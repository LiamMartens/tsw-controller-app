use futures_util::{SinkExt, StreamExt};
use std::{ffi::CStr, sync::Arc, time::Duration};
use tungstenite::Utf8Bytes;

use tokio::{
    runtime::Runtime,
    sync::{mpsc, Mutex},
};
use tokio_tungstenite::connect_async;

const SYNC_CONTROL_WS_ADDR: &str = "ws://127.0.0.1:63242";

pub struct SyncControllerTask {
    tokio_runtime: &'static Runtime,
    /* channel for sending and receiving */
    sync_control_channel_tx: Arc<mpsc::Sender<String>>,
    sync_control_channel_rx: Arc<Mutex<mpsc::Receiver<String>>>,
}

impl SyncControllerTask {
    /* this task handles reading messages from the message channel and sends them to the SC WS connection */
    pub fn spawn_sc_forwarding_task(&self) {
        let message_channel_rx = Arc::clone(&self.sync_control_channel_rx);

        self.tokio_runtime.spawn(async move {
            let mut message_channel_rx_lock = message_channel_rx.lock().await;
            loop {
                match connect_async(SYNC_CONTROL_WS_ADDR).await {
                    Ok((mut socket, _)) => {
                        println!("[SyncControllerTask] Connected..");
                        loop {
                            tokio::select! {
                                Some(msg) = message_channel_rx_lock.recv() => {
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

    pub unsafe fn send(&self, raw: *const std::ffi::c_char) {
        let sync_control_channel_tx = Arc::clone(&self.sync_control_channel_tx);

        let has_capacity = sync_control_channel_tx.capacity() > 0;
        if !has_capacity {
            eprintln!("[SyncControllerTask] Channel is full, dropping message");
            return;
        }

        let message = String::from(CStr::from_ptr(raw).to_str().unwrap());
        println!(
            "[SyncControllerTask] Sending SC Message: {}",
            message.clone()
        );

        match sync_control_channel_tx.blocking_send(format!("sync_control,{}", message.clone())) {
            Ok(_) => {}
            Err(e) => {
                eprintln!("[SyncControllerTask] Error sending SC message: {}", e);
            }
        }
    }

    pub fn new(tokio_runtime: &'static Runtime) -> SyncControllerTask {
        let (sync_control_channel_tx, sync_control_channel_rx) = mpsc::channel::<String>(50);
        SyncControllerTask {
            tokio_runtime,
            sync_control_channel_tx: Arc::new(sync_control_channel_tx),
            sync_control_channel_rx: Arc::new(Mutex::new(sync_control_channel_rx)),
        }
    }
}
