use futures_util::StreamExt;
use std::{
    collections::{HashMap, VecDeque},
    ffi::CString,
    sync::Arc,
    time::Duration,
};

use tokio::{
    runtime::Runtime,
    sync::{watch, Mutex},
};
use tokio_tungstenite::connect_async;

const DIRECT_CONTROL_WS_ADDR: &str = "ws://127.0.0.1:63241";

pub struct DirectControllerTask {
    tokio_runtime: &'static Runtime,
    /* direct control callback watcher values */
    direct_control_callback_tx: Arc<watch::Sender<Option<extern "C" fn(*const std::ffi::c_char)>>>,
    direct_control_callback_rx:
        Arc<watch::Receiver<Option<extern "C" fn(*const std::ffi::c_char)>>>,
    /* message queue */
    direct_control_message_queue: Arc<Mutex<VecDeque<String>>>,
}

impl DirectControllerTask {
    /* this task handles listening to the DC WS and propagates the values into a VecDequeue */
    pub fn spawn_dc_listener_task(&self) {
        let message_queue = Arc::clone(&self.direct_control_message_queue);
        self.tokio_runtime.spawn(async move {
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
                                                message_queue.lock().await;
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
    pub fn spawn_queue_propagation_task(&self) {
        let callback_channel_rx = Arc::clone(&self.direct_control_callback_rx);
        let message_queue = Arc::clone(&self.direct_control_message_queue);
        self.tokio_runtime.spawn(async move {
            loop {
                let callback_option = callback_channel_rx.borrow().clone();
                if callback_option.is_none() {
                    continue;
                }

                let mut message_queue = message_queue.lock().await;
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
                    callback_option.as_ref().unwrap()(
                        CString::new(format!("direct_control,{},{}", control, value))
                            .unwrap()
                            .as_ptr(),
                    );
                    // wait between each sending of the message
                    tokio::time::sleep(Duration::from_millis(1000 / 30)).await;
                }
                // drop locks before waiting
                drop(message_queue);
                tokio::time::sleep(Duration::from_millis(1000 / 30)).await;
            }
        });
    }

    pub fn set_callback(&self, callback: extern "C" fn(*const std::ffi::c_char)) {
        match self.direct_control_callback_tx.send(Some(callback)) {
            Ok(_) => {}
            Err(e) => {
                eprintln!("[DirectControllerTask] Failed to set callback: {}", e);
            }
        }
    }

    pub fn new(tokio_runtime: &'static Runtime) -> DirectControllerTask {
        // let handle = lua.create_table().unwrap();
        let (direct_control_callback_tx, direct_control_callback_rx) =
            watch::channel::<Option<extern "C" fn(*const std::ffi::c_char)>>(None);
        let direct_control_message_queue: Arc<Mutex<VecDeque<String>>> =
            Arc::new(Mutex::new(VecDeque::new()));

        DirectControllerTask {
            tokio_runtime,
            direct_control_message_queue,
            direct_control_callback_tx: Arc::new(direct_control_callback_tx),
            direct_control_callback_rx: Arc::new(direct_control_callback_rx),
        }
    }
}
