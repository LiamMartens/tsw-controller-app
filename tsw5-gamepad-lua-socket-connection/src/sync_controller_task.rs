use futures_util::{SinkExt, StreamExt};
use std::time::Duration;
use tungstenite::Utf8Bytes;

use mlua::{Lua, Table};
use tokio::{runtime::Runtime, sync::mpsc};
use tokio_tungstenite::connect_async;

const SYNC_CONTROL_WS_ADDR: &str = "ws://127.0.0.1:63242";

pub struct SyncControllerTask {}

impl SyncControllerTask {
    /* this task handles reading messages from the message channel and sends them to the SC WS connection */
    pub fn spawn_sc_forwarding_task(tokio_runtime: &'static Runtime, mut message_channel_rx: mpsc::Receiver<String>) {
        tokio_runtime.spawn(async move {
            loop {
                match connect_async(SYNC_CONTROL_WS_ADDR).await {
                    Ok((mut socket, _)) => {
                        println!("[SyncControllerTask] Connected..");
                        loop {
                            tokio::select! {
                                Some(msg) = message_channel_rx.recv() => {
                                    println!("[SyncControllerTask] Sending Message: {}", msg.clone());
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

                tokio::time::sleep(Duration::from_secs(5)).await;
                println!("[SC] Reconnecting...");
            }
        });
    }

    pub fn connect(tokio_runtime: &'static Runtime, lua: &Lua) -> Table {
        let table = lua.create_table().unwrap();
        let (sync_control_channel_tx, mut sync_control_channel_rx) = mpsc::channel::<String>(1000);

        SyncControllerTask::spawn_sc_forwarding_task(tokio_runtime, sync_control_channel_rx);

        table
            .set(
                "send",
                lua.create_function(move |_, message: String| {
                    println!("[SyncControllerTask] Sending SC Message: {}", message.clone());
                    match sync_control_channel_tx.blocking_send(format!("sync_control,{}", message.clone())) {
                        Ok(_) => {}
                        Err(e) => {
                            eprintln!("[SyncControllerTask] Error sending SC message: {}", e);
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
