use std::{collections::HashMap, sync::Arc};

use futures_util::{SinkExt, StreamExt};
use serde::{Deserialize, Serialize};
use tokio::{
    net::TcpListener,
    sync::{broadcast::Sender, Mutex},
    task::JoinHandle,
};
use tokio_util::sync::CancellationToken;
use tungstenite::protocol::Message;

pub struct SyncControllerControlState {
    pub identifier: String,
    pub current_value: f32,
    pub target_value: f32,
    /** [-1,0,1] -> decreasing, idle, increasing */
    pub moving: i8,
}

pub struct SyncController {
    server: Arc<TcpListener>,
    controls_state: Arc<Mutex<HashMap<String, SyncControllerControlState>>>,
}

impl SyncController {
    pub async fn new() -> Self {
        let direct_control_server = TcpListener::bind("0.0.0.0:63242").await.unwrap();

        Self {
            server: Arc::new(direct_control_server),
            controls_state: Arc::new(Mutex::new(HashMap::new())),
        }
    }

    pub fn start(&self, cancel_token: CancellationToken) -> JoinHandle<()> {
        let server = Arc::clone(&self.server);

        let controls_state = Arc::clone(&self.controls_state);
        let accept_incoming_clients_server = Arc::clone(&server);
        let accept_incoming_clients_cancel_token = cancel_token.clone();

        tokio::task::spawn(async move {
            let cancel_token_clone: CancellationToken =
                accept_incoming_clients_cancel_token.clone();
            loop {
                tokio::select! {
                    _ = cancel_token_clone.cancelled() => {
                      break;
                  },
                  Ok((tcp_stream, _)) = accept_incoming_clients_server.accept() => {
                    println!("[SC] New client connected");

                    let controls_state = Arc::clone(&controls_state);
                    let socket_cancel_token = cancel_token_clone.clone();

                    tokio::task::spawn(async move {
                      let ws_stream = tokio_tungstenite::accept_async(tcp_stream)
                        .await
                        .expect("[SC] Error during the websocket handshake occurred");
                      let (_, mut read) = ws_stream.split();

                      loop {
                        tokio::select! {
                          _ = socket_cancel_token.cancelled() => {
                            break;
                          },
                          Some(next) = read.next() => {
                            match next {
                              Ok(message) => match message {
                                tungstenite::Message::Text(text) => {
                                  println!("[SC] Received message: {}", text);
                                  /* message should follow format sync_control,{identifier},{value} */
                                  let parts = text.split(",").collect::<Vec<&str>>();
                                  /* skip message if not sync_control message */
                                  if parts[0] != "sync_control" || parts.len() != 3 {
                                    continue;
                                  }

                                  let mut controls_state_lock = controls_state.lock().await;
                                  match controls_state_lock.get_mut(parts[1]) {
                                   Some(control_state) => {
                                     control_state.current_value = parts[2].parse::<f32>().unwrap();
                                   },
                                   None => {
                                    controls_state_lock.insert(String::from(parts[1]), SyncControllerControlState {
                                      identifier: parts[1].to_string(),
                                      current_value: parts[2].parse::<f32>().unwrap(),
                                      target_value: parts[2].parse::<f32>().unwrap(),
                                      moving: 0,
                                    });
                                   },
                                  };
                                },
                                tungstenite::Message::Close(_) => { break },
                                _ => {},
                              },
                              Err(e) => {
                                eprintln!("Client error: {}", e);
                                break;
                              }
                            }
                          },
                        }
                      }
                    });
                  }
                }
            }
        })
    }
}
