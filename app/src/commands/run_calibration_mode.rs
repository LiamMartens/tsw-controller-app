use std::{collections::HashMap, sync::Arc};
use tokio::io::{self, AsyncBufReadExt, AsyncWriteExt, BufReader};
use tokio::sync::Mutex;
use tokio_util::sync::CancellationToken;

use crate::{
    config_defs::{
        controller_calibration::{ControllerCalibration, ControllerCalibrationData},
        controller_sdl_map::{ControllerSdlMap, ControllerSdlMapControl, SDLControlKind},
        serde_sdl_guid::SDLGuid,
    },
    config_loader,
    controller_manager::{self, ControllerManagerRawEvent},
};

pub async fn run_calibration_mode<T: AsRef<str>>(config_dir: T) {
    println!("Running calibration mode; press Q and hit enter to stop and write config files.");

    let config = config_loader::ConfigLoader::new();
    let config_arc = Arc::new(config);
    let mut controller_manager = controller_manager::ControllerManager::new(Arc::clone(&config_arc));

    let cancel_token = CancellationToken::new();
    let receiver = controller_manager.raw_receiver();

    let controller_sdl_mappings = Arc::new(Mutex::new(HashMap::<String, ControllerSdlMap>::new()));
    let controller_calibrations = Arc::new(Mutex::new(HashMap::<String, ControllerCalibration>::new()));

    let cancel_token_clone = cancel_token.clone();
    let controller_sdl_mappings_task_arc = Arc::clone(&controller_sdl_mappings);
    let controller_calibrations_task_arc = Arc::clone(&controller_calibrations);
    let event_listener_task = tokio::task::spawn(async move {
        let stdin = io::stdin();
        let mut stdout = io::stdout();
        let mut stdin_reader = BufReader::new(stdin);
        let (stdin_read_channel_tx, mut stdin_read_channel_rx) = tokio::sync::mpsc::channel::<Result<String, String>>(10);

        loop {
            let mut last_line_input = String::new();
            tokio::select! {
                _ = cancel_token_clone.cancelled() => {
                        break;
                }
              read_result = stdin_reader.read_line(&mut last_line_input) => {
                match read_result {
                  Ok(_) => {
                    if last_line_input.to_lowercase().trim() == "q" {
                      cancel_token_clone.cancel();
                      break;
                    }
                    stdin_read_channel_tx.send(Ok(last_line_input.clone())).await.unwrap();
                  },
                  Err(_) => {
                    stdin_read_channel_tx.send(Err("Could not read line".to_string())).await.unwrap();
                  }
                };
              },
                _ = async {
                  use sdl2::event::Event;

                  let raw_event: ControllerManagerRawEvent = receiver.lock().await.recv().await.unwrap();

                  let mut sdl_mapping_lock = controller_sdl_mappings_task_arc.lock().await;
                  let mut controller_calibrations_lock = controller_calibrations_task_arc.lock().await;
                  let existing_sdl_map = sdl_mapping_lock.get_mut(&raw_event.joystick_guid);
                  let existing_calibration = controller_calibrations_lock.get_mut(&raw_event.joystick_guid);

                  let vendor_id = unsafe { sdl2_sys::SDL_JoystickGetDeviceVendor(raw_event.joystick_index as i32) };
                  let product_id = unsafe { sdl2_sys::SDL_JoystickGetDeviceProduct(raw_event.joystick_index as i32) };

                  let mut controller_sdl_map: ControllerSdlMap = match &existing_sdl_map {
                    Some(sdl_map) => (*sdl_map).clone(),
                    None => ControllerSdlMap {
                      usb_id: format!("{:04x}:{:04x}", vendor_id, product_id),
                      name: "Unknown".to_string(),
                      data: vec![],
                    }
                  };
                  let mut controller_calibration: ControllerCalibration = match &existing_calibration {
                    Some(calibration) => (*calibration).clone(),
                    None => ControllerCalibration {
                      usb_id: format!("{:04x}:{:04x}", vendor_id, product_id),
                      data: vec![],
                    }
                  };

                  match raw_event.event {
                    Event::JoyAxisMotion { axis_idx, value, .. } => {
                      stdout.write_all(format!("[{}][{}] Axis {} moved to {}\n", raw_event.joystick_guid, raw_event.joystick_name, axis_idx, value).as_bytes()).await.unwrap();
                      stdout.flush().await.unwrap();

                      let control_name = format!("Axis{}", axis_idx);
                      if !controller_sdl_map.data.iter().any(|c| c.kind == SDLControlKind::Axis && c.index == axis_idx) {
                        stdout.write_all(b"Enter common name for this axis: ").await.unwrap();
                        stdout.flush().await.unwrap();
                        let input = stdin_read_channel_rx.recv().await.unwrap().expect("Could not read input");

                        controller_sdl_map.data.push(ControllerSdlMapControl {
                          kind: SDLControlKind::Axis,
                          index: axis_idx,
                          name: match input.trim().to_string() {
                            s if s.is_empty() => control_name.clone(),
                            s => s,
                          },
                        });
                      }

                      let existing_calibration_index = controller_calibration.data.iter().position(|c| c.id == control_name).unwrap_or(controller_calibration.data.len());
                      let mut control_calibration: ControllerCalibrationData = match &controller_calibration.data.get(existing_calibration_index) {
                        Some(calibration) => (*calibration).clone(),
                        None => {
                          ControllerCalibrationData {
                            id: control_name.clone(),
                            min: 0.0f32,
                            max: 1.0f32,
                            deadzone: Some(0.0f32),
                            idle: 0.0f32,
                            easing_curve: None,
                            invert: Some(false),
                          }
                        }
                      };
                      control_calibration.min = control_calibration.min.min(value as f32);
                      control_calibration.idle = control_calibration.idle.min(value as f32);
                      control_calibration.max = control_calibration.max.max(value as f32);
                      if existing_calibration_index < controller_calibration.data.len() {
                          controller_calibration.data[existing_calibration_index] = control_calibration;
                      } else {
                          controller_calibration.data.push(control_calibration);
                      }
                    },
                    Event::JoyButtonDown {button_idx, ..} | Event::JoyButtonUp {button_idx, ..} => {
                      stdout.write_all(format!("[{}][{}] Button {} triggered\n",  raw_event.joystick_guid, raw_event.joystick_name, button_idx).as_bytes()).await.unwrap();
                      stdout.flush().await.unwrap();

                      if !controller_sdl_map.data.iter().any(|c| c.kind == SDLControlKind::Button && c.index == button_idx) {
                        stdout.write_all(b"Enter common name for this button: ").await.unwrap();
                        stdout.flush().await.unwrap();
                        let input = stdin_read_channel_rx.recv().await.unwrap().expect("Could not read input");

                        controller_sdl_map.data.push(ControllerSdlMapControl {
                          kind: SDLControlKind::Button,
                          index: button_idx,
                          name: match input.trim().to_string() {
                            s if s.is_empty() => format!("Button{}", button_idx),
                            s => s,
                          },
                        });
                      }
                    },
                    Event::JoyHatMotion {hat_idx, ..}  => {
                      stdout.write_all(format!("[{}][{}] Hat {} triggered\n", raw_event.joystick_guid, raw_event.joystick_name, hat_idx).as_bytes()).await.unwrap();
                      stdout.flush().await.unwrap();

                      if !controller_sdl_map.data.iter().any(|c| c.kind == SDLControlKind::Hat && c.index == hat_idx) {
                        stdout.write_all(b"Enter common name for this hat: ").await.unwrap();
                        stdout.flush().await.unwrap();
                        let input = stdin_read_channel_rx.recv().await.unwrap().expect("Could not read input");

                        controller_sdl_map.data.push(ControllerSdlMapControl {
                          kind: SDLControlKind::Hat,
                          index: hat_idx,
                          name: match input.trim().to_string() {
                            s if s.is_empty() => format!("Hat{}", hat_idx),
                            s => s,
                          },
                        });
                      }
                    }
                    _ => {},
                  };

                  sdl_mapping_lock.insert(raw_event.joystick_guid.clone(), controller_sdl_map);
                  controller_calibrations_lock.insert(raw_event.joystick_guid.clone(), controller_calibration);
                } => {}
            }
        }
    });

    controller_manager.attach(cancel_token.clone());
    cancel_token.cancel();
    event_listener_task.await.unwrap();

    println!("Writing new config files..");
    let mut config = config_loader::ConfigLoader::new();
    let sdl_mappings_lock = controller_sdl_mappings.lock().await;
    let calibrations_lock = controller_calibrations.lock().await;
    for (_, sdl_map) in sdl_mappings_lock.iter() {
        config.register_sdl_mapping(sdl_map.clone());
    }
    for (_, calibration) in calibrations_lock.iter() {
        config.register_calibration(calibration.clone());
    }
    config.export(config_dir.as_ref());
}
