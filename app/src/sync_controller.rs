use std::{collections::HashMap, sync::Arc};

use futures_util::StreamExt;
use serde::{Deserialize, Serialize};
use tokio::{
    net::TcpListener,
    sync::broadcast::{self, Receiver, Sender},
    sync::Mutex,
    task::JoinHandle,
};
use tokio_util::sync::CancellationToken;

use crate::{
    action_sequencer::{ActionSequencer, ActionSequencerAction},
    config_defs::controller_profile::{
        ControllerProfileControlAssignment, ControllerProfileControlAssignmentKeysAction,
        ControllerProfileDirectControAssignmentSyncMode, PreferredControlMode,
    },
    config_loader::ConfigLoader,
    controller_manager::ControllerManagerChangeEvent,
    profile_runner::ProfileRunner,
};

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct SyncControllerControlState {
    pub identifier: String,
    pub current_value: f32,
    pub target_value: f32,
    /** [-1,0,1] -> decreasing, idle, increasing */
    pub moving: i8,
    pub target_profile: Option<ControllerProfileDirectControAssignmentSyncMode>,
}

pub struct SyncController {
    config: Arc<ConfigLoader>,
    sequencer: Arc<ActionSequencer>,
    profile_runner: Arc<Mutex<ProfileRunner>>,
    server: Arc<TcpListener>,
    controls_state_profile: Arc<Mutex<Option<String>>>,
    controls_state: Arc<Mutex<HashMap<String, SyncControllerControlState>>>,
    control_state_changed_channel: (
        Sender<SyncControllerControlState>,
        Receiver<SyncControllerControlState>,
    ),
}

impl SyncController {
    pub async fn new(
        config: Arc<ConfigLoader>,
        sequencer: Arc<ActionSequencer>,
        profile_runner: Arc<Mutex<ProfileRunner>>,
    ) -> Self {
        let direct_control_server = TcpListener::bind("0.0.0.0:63242").await.unwrap();

        Self {
            config,
            sequencer,
            profile_runner,
            server: Arc::new(direct_control_server),
            controls_state_profile: Arc::new(Mutex::new(None)),
            controls_state: Arc::new(Mutex::new(HashMap::new())),
            control_state_changed_channel: broadcast::channel::<SyncControllerControlState>(10000),
        }
    }

    pub fn start(
        &self,
        cancel_token: CancellationToken,
        mut controller_receiver: Receiver<ControllerManagerChangeEvent>,
    ) -> JoinHandle<()> {
        let server = Arc::clone(&self.server);

        /* listen to current value changed channel to stop actions */
        let controls_state = Arc::clone(&self.controls_state);
        let sequencer = Arc::clone(&self.sequencer);
        let current_value_changed_handler_cancel_token = cancel_token.clone();
        let control_state_changed_channel_sender = self.control_state_changed_channel.0.clone();
        let mut control_state_changed_channel_receiver =
            control_state_changed_channel_sender.subscribe();
        tokio::task::spawn(async move {
            loop {
                tokio::select! {
                  _ = current_value_changed_handler_cancel_token.cancelled() => {
                    break;
                  },
                  Ok(state) = control_state_changed_channel_receiver.recv() => {
                    /* ignore if there is no target profile */
                    if state.target_profile.is_none() {
                      continue;
                    }

                    let mut control_state_lock = controls_state.lock().await;
                    /* unwrapping since it should always exist */
                    let mut_control_state = control_state_lock.get_mut(state.identifier.as_str()).unwrap();
                    let MARGIN_OF_ERROR = 0.03;
                    let should_stop_moving =
                      /* was increasing and has now exceeded value */
                      (mut_control_state.current_value > mut_control_state.target_value && mut_control_state.moving == 1)
                      /* was decreasing and has now subceeded value */
                      || (mut_control_state.current_value < mut_control_state.target_value && mut_control_state.moving == -1
                      /* otherwise is within margin of error and was moving */
                      || (mut_control_state.current_value - mut_control_state.target_value).abs() < MARGIN_OF_ERROR && mut_control_state.moving != 0);
                    let should_start_increasing = mut_control_state.target_value > mut_control_state.current_value && (mut_control_state.current_value - mut_control_state.target_value).abs() > MARGIN_OF_ERROR && mut_control_state.moving != 1;
                    let should_start_decreasing =  mut_control_state.target_value < mut_control_state.current_value && (mut_control_state.current_value - mut_control_state.target_value).abs() > MARGIN_OF_ERROR && mut_control_state.moving != -1;
                    /* stop moving */
                    if should_stop_moving {
                      let action_to_release = match mut_control_state.moving {
                        1 => mut_control_state.target_profile.clone().unwrap().action_increase,
                        _ => mut_control_state.target_profile.clone().unwrap().action_decrease,
                      };
                      sequencer.add_action(ActionSequencerAction {
                        keys: action_to_release.keys.clone(),
                        press_time: action_to_release.press_time,
                        wait_time: action_to_release.wait_time,
                        release: Some(true),
                      }).await;
                      /* set moving param to 0 */
                      mut_control_state.moving = 0;
                    }

                    /* should start increasing if not already */
                    if should_start_increasing {
                      let action = mut_control_state.target_profile.clone().unwrap().action_increase;
                      sequencer.add_action(ActionSequencerAction {
                        keys: action.keys.clone(),
                        press_time: action.press_time,
                        wait_time: action.wait_time,
                        release: Some(false),
                      }).await;
                      /* set moving param to 0 */
                      mut_control_state.moving = 1;
                    }

                    /* should start decreasing if not already */
                    if should_start_decreasing {
                      let action = mut_control_state.target_profile.clone().unwrap().action_decrease;
                      sequencer.add_action(ActionSequencerAction {
                        keys: action.keys.clone(),
                        press_time: action.press_time,
                        wait_time: action.wait_time,
                        release: Some(false),
                      }).await;
                      /* set moving param to 0 */
                      mut_control_state.moving = -1;
                    }
                  },
                }
            }
        });

        /* listen to incoming controller receiver to update target values */
        let controls_state = Arc::clone(&self.controls_state);
        let controls_state_profile = Arc::clone(&self.controls_state_profile);
        let profile_runner = Arc::clone(&self.profile_runner);
        let target_value_changed_cancel_token = cancel_token.clone();
        let control_state_changed_channel_sender = self.control_state_changed_channel.0.clone();
        tokio::task::spawn(async move {
            loop {
                tokio::select! {
                  _ = target_value_changed_cancel_token.cancelled() => {
                    break;
                  },
                  Ok(event) = controller_receiver.recv() => {
                    let profile_runner_lock = profile_runner.lock().await;
                    let profile = profile_runner_lock.get_current_profile(Some(event.joystick_guid));
                    let preferred_control_mode = profile_runner_lock.get_preferred_control_mode();
                    match profile {
                      Some(profile) => {
                        let controls_state_profile_lock = controls_state_profile.lock().await;
                        if controls_state_profile_lock.is_some() && controls_state_profile_lock.as_ref().unwrap() != &profile.name {
                          /* profile changed - clear current state */
                          let mut controls_state_lock = controls_state.lock().await;
                          controls_state_lock.clear();
                          drop(controls_state_lock);
                        }

                        let control = profile.find_control(event.control_name);
                        if let Some(control_config) = control {
                          let assignments = control_config.get_assignments(preferred_control_mode);
                          for assignment in assignments.iter() {
                            if let ControllerProfileControlAssignment::SyncControl(sync_control_action) = assignment {
                              let target_value = sync_control_action.input_value.calculate_normal_value(event.control_state.value);
                              let mut controls_state_lock = controls_state.lock().await;
                              let updated_state = match controls_state_lock.get_mut(sync_control_action.identifier.as_str()) {
                                Some(control_state) => {
                                  control_state.target_value = target_value;
                                  control_state.target_profile = Some(sync_control_action.clone());
                                  control_state
                                },
                                None => {
                                let new_control_state = SyncControllerControlState {
                                  identifier: sync_control_action.identifier.clone(),
                                  target_value,
                                  current_value: target_value,
                                  moving: 0,
                                  target_profile: Some(sync_control_action.clone()),
                                };
                                controls_state_lock.insert(sync_control_action.identifier.clone(), new_control_state);
                                controls_state_lock.get_mut(sync_control_action.identifier.as_str()).unwrap()
                                },
                              };
                              control_state_changed_channel_sender.send(updated_state.clone()).unwrap();
                            }
                          }
                        }
                      },
                      None => {
                      },
                    }
                  },
                }
            }
        });

        /* listen to incoming WS messags to update current value */
        let controls_state = Arc::clone(&self.controls_state);
        let accept_incoming_clients_server = Arc::clone(&server);
        let accept_incoming_clients_cancel_token = cancel_token.clone();
        let control_state_changed_channel_sender = self.control_state_changed_channel.0.clone();
        tokio::task::spawn(async move {
            println!("[SC] Server started");
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
                    let control_state_changed_channel_sender = control_state_changed_channel_sender.clone();

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
                                  let updated_state = match controls_state_lock.get_mut(parts[1]) {
                                    Some(control_state) => {
                                      control_state.current_value = parts[2].parse::<f32>().unwrap();
                                      control_state
                                    },
                                    None => {
                                    let new_control_state = SyncControllerControlState {
                                      identifier: parts[1].to_string(),
                                      current_value: parts[2].parse::<f32>().unwrap(),
                                      target_value: parts[2].parse::<f32>().unwrap(),
                                      moving: 0,
                                      target_profile: None,
                                    };
                                    controls_state_lock.insert(String::from(parts[1]), new_control_state);
                                    controls_state_lock.get_mut(parts[1]).unwrap()
                                    },
                                  };
                                  control_state_changed_channel_sender.send(updated_state.clone()).unwrap();
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
