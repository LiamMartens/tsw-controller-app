use futures_util::{SinkExt, StreamExt};
use once_cell::sync::Lazy;
use std::ffi::{CStr, CString};
use std::sync::{Arc};
use tokio::sync::{RwLock};
use tokio::runtime::Runtime;
use tokio::sync::mpsc::{self, Sender};
use tokio_tungstenite::connect_async;
use tungstenite::{protocol::Message, Utf8Bytes};

static WS_PORT_OPTIONS: &[u16] = &[63241, 63242, 63243];

/// C callback signature: void (*MessageCallback)(const char*)
pub type MessageCallback = extern "C" fn(*const std::ffi::c_char);

/// Holds state of the DLL
struct DLLState {
    rt: Option<Runtime>,
    stop_tx: Option<Sender<()>>,
    outgoing_tx: Option<Sender<String>>,
    callback: Option<MessageCallback>,
    current_port_index: usize,
}

static STATE: Lazy<Arc<RwLock<DLLState>>> = Lazy::new(|| {
    Arc::new(RwLock::new(DLLState {
        rt: None,
        stop_tx: None,
        outgoing_tx: None,
        callback: None,
        current_port_index: 0,
    }))
});

async fn cycle_current_port_index() {
    let mut state_guard = STATE.write().await;
    state_guard.current_port_index = (state_guard.current_port_index + 1) % WS_PORT_OPTIONS.len();
}

/// Start WebSocket loop inside a Tokio runtime
#[no_mangle]
pub extern "C" fn tsw_controller_mod_start() {
    println!("[socket_connection_lib][info] starting tsw_controller_mod");

    if STATE.blocking_write().rt.is_some() {
        return; // already running
    }

    // create tokio runtime
    let rt =  tokio::runtime::Builder::new_multi_thread().worker_threads(1)
        .enable_all().build().expect("Failed to create runtime");
    // create channels
    let (stop_tx, mut stop_rx) = mpsc::channel::<()>(1);
    let (out_tx, mut out_rx) = mpsc::channel::<String>(64);

    rt.spawn(async move {
        loop {
            let current_port = {
                STATE.read().await.current_port_index
            };

            let ws_url = format!("ws://127.0.0.1:{}", WS_PORT_OPTIONS[current_port]);
            println!("[socket_connection_lib][info] attempting to connect to socket using port {}", WS_PORT_OPTIONS[current_port]);

            tokio::select! {
                _ = stop_rx.recv() => {
                    break;
                }
                conect_res = connect_async(ws_url.as_str()) => {
                    match conect_res {
                        Ok((ws_stream, response)) => {
                            let header = response.headers().get("X-TSW-Version");
                            if header.is_none() || header.unwrap().is_empty() {
                                println!("[socket_connection_lib][error] connected to unknown socket server - switching and retrying in 3s");
                                tokio::time::sleep(std::time::Duration::from_secs(3)).await;
                                /* update port index to next one */
                                cycle_current_port_index().await;
                                continue;
                            }

                            let (mut ws_write, mut ws_read) = ws_stream.split();

                            let (reconnect_tx, mut reconnect_rx) = mpsc::channel::<()>(1);

                            // Forward incoming WS messages to callback
                            tokio::spawn(async move {
                                while let Some(Ok(msg)) = ws_read.next().await {
                                  match msg {
                                    tungstenite::Message::Text(text) => {
                                        let state_guard = STATE.read().await;
                                        if let Some(cb) = state_guard.callback {
                                            if let Ok(cstr) = CString::new(text.to_string()) {
                                                println!("[socket_connection_lib][info] received message from socket | {}", text);
                                                let boxed_cstr = Box::new(cstr);
                                                cb(boxed_cstr.as_ptr());
                                                // ⚠️ Important: must keep CString alive until cb returns
                                                // that's why cstr lives inside this block
                                            }
                                        }
                                    },
                                    tungstenite::Message::Close(_) => {
                                        break;
                                    },
                                    _ => {},
                                  }
                                }
                                /* if this while ends - the read resulted in an error - try send reconnect_tx */
                                println!("[socket_connection_lib][info] closing connection and reconnecting due to error or close message");
                                let _ = reconnect_tx.try_send(());
                            });

                            // Outgoing loop
                            loop {
                                tokio::select! {
                                    Some(msg) = out_rx.recv() => {
                                        println!("[socket_connection_lib][info] sending message | {}", msg);
                                        if let Err(e) = ws_write.send(Message::Text(Utf8Bytes::from(msg))).await {
                                           println!("[socket_connection_lib][info] failed to send message | {}", e);
                                            break; // reconnect
                                        }
                                    }
                                    _ = reconnect_rx.recv() => {
                                      break;
                                    },
                                    _ = stop_rx.recv() => {
                                        let _ = ws_write.send(Message::Close(None)).await;
                                        return;
                                    }
                                }
                            }
                        }
                        Err(e) => {
                            println!("[socket_connection_lib][error] failed to connect to socket - retrying in 3s | {}", e);
                            tokio::time::sleep(std::time::Duration::from_secs(3)).await;
                            cycle_current_port_index().await;
                            continue;
                        }
                    }
                }
            }
        }
    });

    {
        let mut state_guard = STATE.blocking_write();
        state_guard.rt = Some(rt);
        state_guard.stop_tx = Some(stop_tx);
        state_guard.outgoing_tx = Some(out_tx);
    }
}

/// Stop the module
#[no_mangle]
pub extern "C" fn tsw_controller_mod_stop() {
    let mut st = STATE.blocking_write();
    if let Some(stop_tx) = st.stop_tx.take() {
        let _ = stop_tx.try_send(());
    }
    st.rt.take(); // dropping runtime shuts it down
}

/// Register callback
#[no_mangle]
pub extern "C" fn tsw_controller_mod_set_receive_message_callback(cb: MessageCallback) {
    let mut st = STATE.blocking_write();
    st.callback = Some(cb);
}

/// Send message
#[no_mangle]
pub extern "C" fn tsw_controller_mod_send_message(message: *const std::ffi::c_char) {
    if message.is_null() {
        return;
    }

    let cstr = unsafe {
        let raw_str = CStr::from_ptr(message);
        raw_str.to_str().ok().map(|s| s.to_owned())
    };
    if let Some(msg) = cstr {
        let st = STATE.blocking_read();
        if let Some(tx) = &st.outgoing_tx {
            println!("[socket_connection_lib][info] sending message {}",msg.clone());
            let send_result = tx.try_send(msg);
            if let Err(e) = send_result {
              println!("[socket_connection_lib][error] failed to send message {}", e.to_string());
            }
        }
    } else {
      println!("[socket_connection_lib][error] failed to decode cstr");
    }
}

#[no_mangle]
#[cfg(target_os="windows")]
pub extern "system" fn DllMain(_hinst_dll: *mut u8, _fwd_reason: u32, _lp_reserved: *mut u8) -> i32 {
    1
}