use mlua::{Function, Lua, Result, Table};
use std::{
    collections::{HashMap, VecDeque},
    sync::{Arc, Mutex},
    thread,
    time::Duration,
};
use tungstenite::connect;

static WS_ADDR: &str = "ws://127.0.0.1:63241";

fn init(lua: Arc<&Lua>) -> Result<Table> {
    let exports = lua.create_table()?;
    let callback_arc: Arc<Mutex<Option<Function>>> = Arc::new(Mutex::new(None));
    let incoming_message_queue_arc: Arc<Mutex<VecDeque<String>>> =
        Arc::new(Mutex::new(VecDeque::new()));

    /* loop to send forward incoming messages to Lua */
    let callback_arc_clone = Arc::clone(&callback_arc);
    let incoming_message_queue_arc_clone = Arc::clone(&incoming_message_queue_arc);
    thread::spawn(move || loop {
        loop {
            let callback_lock = callback_arc_clone.lock().unwrap();
            let mut message_queue = incoming_message_queue_arc_clone.lock().unwrap();
            // only take latest of current tick
            let mut control_value_map: HashMap<String, String> = HashMap::new();
            while message_queue.len() > 0 {
                let message = message_queue.pop_front().unwrap();
                let parts: Vec<&str> = message.split(",").collect();
                if parts.len() == 2 {
                    control_value_map.insert(parts[0].to_string(), parts[1].to_string());
                }
            }
            // send all control values to lua - wait a tick each time
            control_value_map.iter().for_each(|(control, value)| {
                callback_lock
                    .as_ref()
                    .unwrap()
                    .call::<()>(format!("{},{}", control, value))
                    .unwrap();
                // wait between each sending of the message
                thread::sleep(Duration::from_millis(1000 / 15));
            });
            // drop locks before waiting
            drop(callback_lock);
            drop(message_queue);
            thread::sleep(Duration::from_millis(1000 / 15));
        }
    });

    /* loop to listen to socket incoming messages */
    let incoming_message_queue_arc_clone = Arc::clone(&incoming_message_queue_arc);
    thread::spawn(move || loop {
        match connect(WS_ADDR) {
            Ok((mut socket, _)) => {
                println!("[Listener] Connected..");

                loop {
                    let msg = socket.read();
                    match msg {
                        Ok(msg) => match msg {
                            tungstenite::Message::Text(text) => {
                                println!("[Listener] Queueing Message: {}", text.to_string());
                                let mut message_queue_lock =
                                    incoming_message_queue_arc_clone.lock().unwrap();
                                message_queue_lock.push_back(text.to_string());
                                drop(message_queue_lock);
                            }
                            tungstenite::Message::Close(_) => {
                                socket.close(None).unwrap();
                                break;
                            }
                            _ => {}
                        },
                        Err(e) => {
                            eprintln!("[Listener] Error receiving message: {}", e);
                            break;
                        }
                    }
                }
            }
            Err(e) => {
                println!("[Listener] Connection error: {}", e);
            }
        }

        thread::sleep(Duration::from_secs(5));
        println!("[Listener] Reconnecting...");
    });

    let callback_arc_clone = Arc::clone(&callback_arc);
    exports.set(
        "set_callback",
        lua.create_function(move |_, callback: Function| {
            let mut callback_lock = callback_arc_clone.lock().unwrap();
            callback_lock.replace(callback);
            drop(callback_lock);
            Ok(())
        })?,
    )?;

    Ok(exports)
}

#[no_mangle]
pub unsafe extern "C" fn luaopen_tsw5_gamepad_lua_socket_connection(
    state: *mut mlua::lua_State,
) -> std::os::raw::c_int {
    mlua::Lua::entrypoint1(state, move |lua| {
        let lua_arc = Arc::new(lua);
        init(lua_arc)
    })
}
