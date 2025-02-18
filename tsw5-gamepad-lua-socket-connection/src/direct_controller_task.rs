use futures_util::StreamExt;
use std::{
    collections::{HashMap, VecDeque},
    sync::Arc,
    time::Duration,
};

use mlua::{Function, Lua, Table};
use tokio::{
    runtime::Runtime,
    sync::{watch, Mutex},
};
use tokio_tungstenite::connect_async;

const DIRECT_CONTROL_WS_ADDR: &str = "ws://127.0.0.1:63241";

pub struct DirectControllerTask {}

impl DirectControllerTask {
    /* this task handles listening to the DC WS and propagates the values into a VecDequeue */
    pub fn spawn_dc_listener_task(tokio_runtime: &'static Runtime, queue: Arc<Mutex<VecDeque<String>>>) {
        tokio_runtime.spawn(async move {
            loop {
                match connect_async(DIRECT_CONTROL_WS_ADDR).await {
                    Ok((mut socket, _)) => {
                        println!("[DirectControllerTask] Connected..");

                        loop {
                            tokio::select! {
                                Some(msg) = socket.next() => {
                                    match msg {
                                        Ok(msg) => match msg {
                                            tungstenite::Message::Text(text) => {
                                                println!("[DirectControllerTask] Queueing Message: {}", text.to_string());
                                                let mut message_queue_lock =
                                                queue.lock().await;
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
                                            eprintln!("[DirectControllerTask] Error receiving message: {}", e);
                                            break;
                                        }
                                    }
                                }
                            }
                        }
                    }
                    Err(e) => {
                        println!("[DirectControllerTask] Connection error: {}", e);
                    }
                }

                tokio::time::sleep(Duration::from_secs(5)).await;
                println!("[DirectControllerTask] Reconnecting...");
            }
        });
    }

    /* this task handles reading messages from the queue periodically and sending them to lua */
    pub fn spawn_queue_propagation_task(tokio_runtime: &'static Runtime, queue: Arc<Mutex<VecDeque<String>>>, callback_channel_rx: watch::Receiver<Option<Function>>) {
        tokio_runtime.spawn(async move {
            loop {
                let callback_lock = callback_channel_rx.borrow().clone();
                if callback_lock.is_none() {
                    continue;
                }

                let mut message_queue = queue.lock().await;
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
                    callback_lock.as_ref().unwrap().call::<()>(format!("direct_control,{},{}", control, value)).unwrap();
                    // wait between each sending of the message
                    tokio::time::sleep(Duration::from_millis(1000 / 30)).await;
                }
                // drop locks before waiting
                drop(callback_lock);
                drop(message_queue);
                tokio::time::sleep(Duration::from_millis(1000 / 30)).await;
            }
        });
    }

    pub fn connect(tokio_runtime: &'static Runtime, lua: &Lua) -> Table {
        let table = lua.create_table().unwrap();
        let (direct_control_callback_tx, direct_control_callback_rx) = watch::channel::<Option<Function>>(None);
        let direct_control_message_queue: Arc<Mutex<VecDeque<String>>> = Arc::new(Mutex::new(VecDeque::new()));

        DirectControllerTask::spawn_dc_listener_task(tokio_runtime, Arc::clone(&direct_control_message_queue));
        DirectControllerTask::spawn_queue_propagation_task(tokio_runtime, Arc::clone(&direct_control_message_queue), direct_control_callback_rx);

        table
            .set(
                "set_callback",
                lua.create_function(move |_, callback: Function| {
                    match direct_control_callback_tx.send(Some(callback)) {
                        Ok(_) => {}
                        Err(e) => {
                            eprintln!("[DirectControllerTask] Failed to set callback: {}", e);
                        }
                    }
                    Ok(())
                })
                .unwrap(),
            )
            .unwrap();

        table
    }
}
