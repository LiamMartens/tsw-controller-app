use std::sync::Arc;

use futures_util::{SinkExt, StreamExt};
use serde::{Deserialize, Serialize};
use tokio::{
    net::TcpListener,
    sync::{broadcast::Receiver, Mutex},
    task::JoinHandle,
};
use tokio_util::sync::CancellationToken;
use tungstenite::protocol::Message;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DirectControlCommand {
    pub controls: String,
    pub input_value: f32,
}

pub struct DirectController {
    server: Arc<TcpListener>,
}

impl DirectController {
    pub async fn new() -> Self {
        let direct_control_server = TcpListener::bind("0.0.0.0:63241").await.unwrap();

        Self {
            server: Arc::new(direct_control_server),
        }
    }

    pub fn start(
        &self,
        cancel_token: CancellationToken,
        command_receiver: Arc<Mutex<Receiver<DirectControlCommand>>>,
    ) -> JoinHandle<()> {
        let server = Arc::clone(&self.server);

        let command_receiver_clone = Arc::clone(&command_receiver);
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
                    println!("New client connected");
                    let socket_cancel_token = cancel_token_clone.clone();
                    let command_receiver_clone = Arc::clone(&command_receiver_clone);
                    tokio::task::spawn(async move {
                      let mut command_receiver_lock = command_receiver_clone.lock().await;
                      let ws_stream = tokio_tungstenite::accept_async(tcp_stream)
                        .await
                        .expect("Error during the websocket handshake occurred");
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
                          Ok(message) = command_receiver_lock.recv() => {
                            let command_to_send = format!("direct_control,{},{}", message.controls, message.input_value);
                            println!("Sending command: {:?}", command_to_send);
                            write.send(Message::text(command_to_send)).await.unwrap();
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
