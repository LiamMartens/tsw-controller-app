use futures_util::{SinkExt, StreamExt};
use mlua::{Function, Lua, Result, Table};
use once_cell::sync::Lazy;
use std::{
    collections::{HashMap, VecDeque},
    sync::Arc,
    thread,
    time::Duration,
};
use tokio::{
    runtime::Runtime,
    sync::{mpsc, Mutex},
};
use tokio_tungstenite::connect_async;
use tokio_util::sync::CancellationToken;
use tungstenite::Utf8Bytes;

const DIRECT_CONTROL_WS_ADDR: &str = "ws://127.0.0.1:63241";
const SYNC_CONTROL_WS_ADDR: &str = "ws://127.0.0.1:63242";
static TOKIO_RUNTIME: Lazy<Runtime> = Lazy::new(|| {
    tokio::runtime::Builder::new_multi_thread()
        .enable_all()
        .build()
        .expect("Failed to create runtime")
});

fn init(lua: Arc<&Lua>) -> Result<Table> {
    let restart_token = CancellationToken::new();

    let exports = lua.create_table()?;
    let (socket_callback_tx, socket_callback_rx) =
        tokio::sync::watch::channel::<Option<Function>>(None);
    let direct_control_message_queue_arc: Arc<Mutex<VecDeque<String>>> =
        Arc::new(Mutex::new(VecDeque::new()));
    let (sync_control_channel_tx, mut sync_control_channel_rx) = mpsc::channel::<String>(10000);

    /* thread to send states up to Lua */
    let task_cancel_token = restart_token.child_token();
    let direct_control_message_queue_arc_clone = Arc::clone(&direct_control_message_queue_arc);
    TOKIO_RUNTIME.spawn(async move {
        loop {
            /* stop task if cancelled */
            if task_cancel_token.is_cancelled() {
                break;
            }

            let callback_lock = socket_callback_rx.borrow().clone();
            if callback_lock.is_none() {
                continue;
            }

            let mut message_queue = direct_control_message_queue_arc_clone.lock().await;
            // only take latest of current tick
            let mut control_value_map: HashMap<String, String> = HashMap::new();
            while message_queue.len() > 0 {
                let message = message_queue.pop_front().unwrap();
                let parts: Vec<&str> = message.split(",").collect();
                if parts.len() == 3 && parts[0] == "direct_control" {
                    control_value_map.insert(parts[1].to_string(), parts[2].to_string());
                }
            }
            // send all control values to lua - wait a tick each time
            for (control, value) in control_value_map.iter() {
                callback_lock
                    .as_ref()
                    .unwrap()
                    .call::<()>(format!("direct_control,{},{}", control, value))
                    .unwrap();
                // wait between each sending of the message
                tokio::time::sleep(Duration::from_millis(1000 / 30)).await;
            }
            // drop locks before waiting
            drop(callback_lock);
            drop(message_queue);
            tokio::time::sleep(Duration::from_millis(1000 / 30)).await;
        }
    });

    /* thread to listen to DC WS and propagate into VecDequeue */
    let task_cancel_token = restart_token.child_token();
    let direct_control_message_queue_arc_clone = Arc::clone(&direct_control_message_queue_arc);
    TOKIO_RUNTIME.spawn(async move {
        loop {
            /* don't reconnect if cancelled */
            if task_cancel_token.is_cancelled() {
                break;
            }

            let inner_cancel_token = task_cancel_token.child_token();
            match connect_async(DIRECT_CONTROL_WS_ADDR).await {
                Ok((mut socket, _)) => {
                    println!("[DC] Connected..");

                    loop {
                        tokio::select! {
                            _ = inner_cancel_token.cancelled() => {
                                /* stop continuous read loop and close socket if cancelled */
                                socket.close(None).await.unwrap();
                                break;
                            },
                            Some(msg) = socket.next() => {
                                match msg {
                                    Ok(msg) => match msg {
                                        tungstenite::Message::Text(text) => {
                                            println!("[DC] Queueing Message: {}", text.to_string());
                                            let mut message_queue_lock =
                                                direct_control_message_queue_arc_clone.lock().await;
                                            message_queue_lock.push_back(text.to_string());
                                            drop(message_queue_lock);
                                        }
                                        tungstenite::Message::Close(_) => {
                                            socket.close(None).await.unwrap();
                                            break;
                                        }
                                        _ => {}
                                    },
                                    Err(e) => {
                                        eprintln!("[DC] Error receiving message: {}", e);
                                        break;
                                    }
                                }
                            }
                        }
                    }
                }
                Err(e) => {
                    println!("[DC] Connection error: {}", e);
                }
            }

            tokio::time::sleep(Duration::from_secs(5)).await;
            println!("[DC] Reconnecting...");
        }
    });

    /* thread to listen to SC channel and forward state to SC WS */
    let task_cancel_token = restart_token.child_token();
    TOKIO_RUNTIME.spawn(async move {
        loop {
            /* don't reconnect if cancelled */
            if task_cancel_token.is_cancelled() {
                break;
            }

            let inner_cancel_token = task_cancel_token.child_token();
            match connect_async(SYNC_CONTROL_WS_ADDR).await {
                Ok((mut socket, _)) => {
                    println!("[SC] Connected..");
                    loop {
                        tokio::select! {
                            _ = inner_cancel_token.cancelled() => {
                                /* stop continuous read loop and close socket if cancelled */
                                socket.close(None).await.unwrap();
                                break;
                            },
                            Some(msg) = sync_control_channel_rx.recv() => {
                                println!("[SC] Sending Message: {}", msg.clone());
                                match socket.send(tungstenite::Message::Text(Utf8Bytes::from(msg))).await {
                                    Err(e) => {
                                        eprintln!("[SC] Error sending message: {}", e);
                                        /* break out of loop to allow re-connecting */
                                        break;
                                    }
                                    Ok(_) => {
                                        println!("[SC] Sent Message");
                                    }
                                };
                            },
                            Some(msg) = socket.next() => {
                                match msg {
                                    Ok(msg) => match msg {
                                        tungstenite::Message::Close(_) => {
                                            socket.close(None).await.unwrap();
                                            break;
                                        }
                                        _ => {}
                                    },
                                    Err(e) => {
                                        eprintln!("[DC] Error receiving message: {}", e);
                                        break;
                                    }
                                }
                            }
                        }
                    }
                }
                Err(e) => {
                    println!("[SC] Connection error: {}", e);
                }
            }

            thread::sleep(Duration::from_secs(5));
            println!("[SC] Reconnecting...");
        }
    });

    exports.set(
        "restart",
        lua.create_function(move |_, _: ()| {
            restart_token.cancel();
            Ok(())
        })?,
    )?;

    exports.set(
        "set_callback",
        lua.create_function(move |_, callback: Function| {
            match socket_callback_tx.send(Some(callback)) {
                Ok(_) => {}
                Err(e) => {
                    eprintln!("Error setting callback: {}", e);
                }
            };
            Ok(())
        })?,
    )?;

    exports.set(
        "send_sync_control_state",
        lua.create_function(move |_, message: String| {
            println!("Sending SC Message: {}", message.clone());
            match sync_control_channel_tx.blocking_send(format!("sync_control,{}", message.clone()))
            {
                Ok(_) => {
                    println!("[SC] Sent message");
                }
                Err(e) => {
                    eprintln!("[SC] Error sending SC message: {}", e);
                }
            }
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
