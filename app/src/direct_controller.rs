use core::fmt;
use std::sync::Arc;

use futures_util::{SinkExt, StreamExt};
use serde::{Deserialize, Serialize};
use tokio::{
    net::TcpListener,
    sync::{broadcast::Sender, Mutex},
    task::JoinHandle,
};
use tokio_util::sync::CancellationToken;
use tungstenite::protocol::Message;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DirectControlCommand {
    pub controls: String,
    pub input_value: f32,
    pub hold: Option<bool>,
}

pub struct DirectController {
    server: Arc<TcpListener>,
}

impl fmt::Display for DirectControlCommand {
    /**
     * Formats the direct control command
     * {control_name},{input_value},{flag|flag}
     */
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let hold_flag = match self.hold {
            Some(true) => "hold".to_string(),
            _ => "".to_string(),
        };
        let flags = vec![hold_flag].iter().filter(|x| !x.is_empty()).map(|x| x.to_string()).collect::<Vec<String>>();
        write!(f, "{},{},{}", self.controls, self.input_value, flags.join("|"))
    }
}

impl DirectController {
    pub async fn new() -> Self {
        let direct_control_server = TcpListener::bind("0.0.0.0:63241").await.unwrap();

        Self {
            server: Arc::new(direct_control_server),
        }
    }

    pub fn start(&self, cancel_token: CancellationToken, direct_control_command_tx: Arc<Mutex<Sender<DirectControlCommand>>>) -> JoinHandle<()> {
        let server = Arc::clone(&self.server);

        let accept_incoming_clients_server = Arc::clone(&server);
        let accept_incoming_clients_cancel_token = cancel_token.clone();
        tokio::task::spawn(async move {
            let cancel_token_clone: CancellationToken = accept_incoming_clients_cancel_token.clone();
            loop {
                tokio::select! {
                    _ = cancel_token_clone.cancelled() => {
                      break;
                  },
                  Ok((tcp_stream, _)) = accept_incoming_clients_server.accept() => {
                    println!("[DC] New client connected");
                    let socket_cancel_token = cancel_token_clone.clone();
                    /* create a new subscriber for each client */
                    let direct_control_command_tx_lock = direct_control_command_tx.lock().await;
                    let mut client_direct_control_command_receiver = direct_control_command_tx_lock.subscribe();
                    drop(direct_control_command_tx_lock);
                    tokio::task::spawn(async move {
                      let ws_stream = match tokio_tungstenite::accept_async(tcp_stream).await {
                        Ok(ws_stream) => ws_stream,
                        Err(e) => {
                          eprintln!("[DC] Error during the websocket handshake occurred: {}", e);
                          return;
                        }
                      };
                      let (mut write, mut read) = ws_stream.split();

                      loop {
                        tokio::select! {
                          _ = socket_cancel_token.cancelled() => {
                            break;
                          },
                          Some(next) = read.next() => {
                            match next {
                              Ok(message) => match message {
                                tungstenite::Message::Close(_) => { break },
                                _ => {},
                              },
                              Err(e) => {
                                eprintln!("Client error: {}", e);
                                break;
                              }
                            }
                          },
                          Ok(message) = client_direct_control_command_receiver.recv() => {
                            let command_to_send = format!("direct_control,{}", message);
                            println!("[DC] Sending command: {:?}", command_to_send);
                            match write.send(Message::text(command_to_send.clone())).await {
                              Ok(_) => {
                                println!("[DC] Message sent: {}", command_to_send.clone());
                              },
                              Err(e) => {
                                eprintln!("[DC] Error sending message: {:?}", e);
                              }
                            }
                          }
                        }
                      }
                    });
                  }
                }
            }
        })
    }
}
