use log::debug;
use std::{collections::HashMap, sync::Arc};
use tokio::sync::broadcast::{Receiver, Sender};
use tokio::sync::Mutex;
use tokio_util::sync::CancellationToken;

use sdl2::{event::Event, joystick::Joystick, Sdl};

use crate::{
    config_defs::{
        controller_calibration::ControllerCalibrationData,
        controller_sdl_map::{ControllerSdlMapControl, SDLControlKind},
    },
    config_loader::ConfigLoader,
};

#[derive(Clone, Debug)]
pub struct ControllerManagerChangeEvent {
    pub joystick_guid: String,
    pub control_name: String,
    pub control_state: ControllerManagerControllerControlState,
}

#[derive(Clone, Debug)]
pub struct ControllerManagerRawEvent {
    pub joystick_index: u32,
    pub joystick_guid: String,
    pub joystick_name: String,
    pub event: Event,
}

#[derive(Debug, Clone, Copy)]
pub struct ControllerManagerControllerControlState {
    /* can be -1 | 0 | 1 depending on the direction; also contains the value at which the direction last changed */
    pub direction: (isize, f32),
    pub value: f32,
    pub previous_value: f32,
    pub initial_value: f32,

    pub raw_value: i16,
    pub raw_previous_value: i16,
    pub raw_initial_value: i16,
}

pub struct ControllerManagerControllerControl {
    joystick: Arc<Joystick>,
    name: String,
    sdl_mapping: ControllerSdlMapControl,
    calibration: Option<ControllerCalibrationData>,
    state: ControllerManagerControllerControlState,

    change_event_channel: (Arc<Sender<ControllerManagerChangeEvent>>, Arc<Mutex<Receiver<ControllerManagerChangeEvent>>>),
}

pub struct ControllerManagerController {
    config: Arc<ConfigLoader>,
    joystick: Arc<Joystick>,
    controls: HashMap<String, ControllerManagerControllerControl>,

    change_event_channel: (Arc<Sender<ControllerManagerChangeEvent>>, Arc<Mutex<Receiver<ControllerManagerChangeEvent>>>),
}

pub struct ControllerManager {
    config: Arc<ConfigLoader>,
    sdl_context: Arc<Sdl>,
    joystick_subsystem: Arc<sdl2::JoystickSubsystem>,
    devices: HashMap<u32, ControllerManagerController>,

    change_event_channel: (Arc<Sender<ControllerManagerChangeEvent>>, Arc<Mutex<Receiver<ControllerManagerChangeEvent>>>),
    raw_event_channel: (Arc<Sender<ControllerManagerRawEvent>>, Arc<Mutex<Receiver<ControllerManagerRawEvent>>>),
}

impl ControllerManagerChangeEvent {
    pub fn has_changed(&self) -> bool {
        self.control_state.value != self.control_state.previous_value && self.control_state.direction.0 != 0
    }
}

impl ControllerManagerControllerControlState {
    pub fn new(idle_value: Option<f32>) -> ControllerManagerControllerControlState {
        ControllerManagerControllerControlState {
            value: idle_value.unwrap_or(0.0),
            previous_value: idle_value.unwrap_or(0.0),
            direction: (0, idle_value.unwrap_or(0.0)),
            initial_value: idle_value.unwrap_or(0.0),
            raw_value: 0,
            raw_previous_value: 0,
            raw_initial_value: 0,
        }
    }
}

impl ControllerManagerControllerControl {
    pub fn new(
        joystick: Arc<Joystick>,
        name: String,
        sdl_mapping: ControllerSdlMapControl,
        calibration: Option<&ControllerCalibrationData>,

        change_event_channel: (Arc<Sender<ControllerManagerChangeEvent>>, Arc<Mutex<Receiver<ControllerManagerChangeEvent>>>),
    ) -> ControllerManagerControllerControl {
        ControllerManagerControllerControl {
            joystick,
            name,
            sdl_mapping,
            calibration: match calibration {
                Some(x) => Some(x.clone()),
                None => None,
            },
            state: ControllerManagerControllerControlState::new(match calibration {
                Some(x) => Some(x.idle),
                None => None,
            }),
            change_event_channel,
        }
    }

    pub fn round_to_margin_of_error(&self, value: f32) -> f32 {
        (value * 10000.0).round() / 10000.0
    }

    pub fn is_within_margin_of_error(&self, one: f32, two: f32) -> bool {
        let margin_of_error = 0.0005;
        let diff = (one - two).abs();
        diff < margin_of_error
    }

    pub fn reset(&mut self) {
        match self.sdl_mapping.kind {
            SDLControlKind::Axis => {
                if let Ok(value) = self.joystick.axis(self.sdl_mapping.index as u32) {
                    self.update_value(value, true);
                }
            }
            SDLControlKind::Button => {
                if let Ok(value) = self.joystick.button(self.sdl_mapping.index as u32) {
                    self.update_value(if value { 1 } else { 0 }, true);
                }
            }
            SDLControlKind::Hat => {
                if let Ok(value) = self.joystick.hat(self.sdl_mapping.index as u32) {
                    self.update_value(value as i16, true);
                }
            }
        }
    }

    pub fn update_value(&mut self, value: i16, is_reset: bool) {
        self.state.raw_previous_value = match is_reset {
            true => value,
            false => self.state.raw_value,
        };
        self.state.raw_initial_value = match is_reset {
            true => value,
            false => self.state.raw_initial_value,
        };
        self.state.raw_value = value;

        match self.sdl_mapping.kind {
            SDLControlKind::Axis => match &self.calibration {
                Some(calibration) => {
                    let normalized_value = calibration.normalize(value);

                    match normalized_value {
                        Some(value) => {
                            let rounded = self.round_to_margin_of_error(value);
                            self.state.initial_value = match is_reset {
                                true => rounded,
                                false => self.state.initial_value,
                            };
                            self.state.previous_value = match is_reset {
                                true => rounded,
                                false => self.state.value,
                            };
                            self.state.value = rounded;
                        }
                        None => { /* deadzone ignore */ }
                    }
                }
                None => {
                    self.state.initial_value = match is_reset {
                        true => value as f32,
                        false => self.state.initial_value,
                    };
                    self.state.previous_value = match is_reset {
                        true => value as f32,
                        false => self.state.value,
                    };
                    self.state.value = match is_reset {
                        true => value as f32,
                        false => match self.is_within_margin_of_error(self.state.value, value as f32) {
                            true => self.state.value,
                            false => value as f32,
                        },
                    };
                }
            },
            SDLControlKind::Button => {
                self.state.initial_value = match is_reset {
                    true => value as f32,
                    false => self.state.initial_value,
                };
                self.state.previous_value = match is_reset {
                    true => value as f32,
                    false => self.state.value,
                };
                self.state.value = value as f32;
            }
            SDLControlKind::Hat => {
                self.state.initial_value = match is_reset {
                    true => value as f32,
                    false => self.state.initial_value,
                };
                self.state.previous_value = match is_reset {
                    true => value as f32,
                    false => self.state.value,
                };
                self.state.value = value as f32;
            }
        }

        self.state.direction = match is_reset {
            true => (0, self.state.value),
            false => {
                let last_direction_change_value = self.state.direction.1;
                let direction_change_threshold = 0.05;
                match self.state.value - last_direction_change_value {
                    x if x > direction_change_threshold => (1, self.state.value),
                    x if x < -direction_change_threshold => (-1, self.state.value),
                    _ => self.state.direction,
                }
            }
        };

        match self.change_event_channel.0.send(ControllerManagerChangeEvent {
            joystick_guid: self.joystick.guid().string(),
            control_name: self.name.clone(),
            control_state: self.state.clone(),
        }) {
            Ok(_) => {}
            Err(err) => {
                debug!("Failed to send controller change event: {}", err);
            }
        };
    }

    pub fn process(&mut self, event: sdl2::event::Event) {
        use sdl2::event::Event;

        match event {
            Event::JoyAxisMotion { value, .. } => self.update_value(value, false),
            Event::JoyButtonDown { .. } => self.update_value(1, false),
            Event::JoyButtonUp { .. } => self.update_value(0, false),
            Event::JoyHatMotion { state, .. } => self.update_value(state as i16, false),
            _ => {}
        }
    }
}

impl ControllerManagerController {
    pub fn new(
        config: Arc<ConfigLoader>,
        usb_id: String,
        joystick: Joystick,
        change_event_channel: (Arc<Sender<ControllerManagerChangeEvent>>, Arc<Mutex<Receiver<ControllerManagerChangeEvent>>>),
    ) -> ControllerManagerController {
        let joystick_arc = Arc::new(joystick);
        let sdl_mapping = config.find_sdl_mapping(&usb_id);
        let calibration = config.find_controller_calibration(&usb_id);
        let all_controls_calibration_data = calibration.map(|x| x.data.clone()).unwrap_or(Vec::new());

        let mut gamepad_controls = HashMap::new();
        if let Some(mapping) = &sdl_mapping {
            mapping.data.iter().for_each(|control| {
                let control_calibration = all_controls_calibration_data.iter().find(|x| x.id == control.name);

                let mut control = ControllerManagerControllerControl::new(
                    Arc::clone(&joystick_arc),
                    control.name.clone(),
                    control.clone(),
                    control_calibration,
                    (Arc::clone(&change_event_channel.0), Arc::clone(&change_event_channel.1)),
                );
                control.reset();
                gamepad_controls.insert(control.name.clone(), control);
            });
        }

        ControllerManagerController {
            config: Arc::clone(&config),
            joystick: joystick_arc,
            controls: gamepad_controls,
            change_event_channel,
        }
    }

    pub fn process(&mut self, event: sdl2::event::Event) {
        debug!("Processing event ({}): {:?}", self.joystick.guid().string(), event);

        use sdl2::event::Event;

        match event {
            Event::JoyAxisMotion { axis_idx, .. } => {
                let axis = self
                    .controls
                    .values_mut()
                    .find(|control| control.sdl_mapping.kind == SDLControlKind::Axis && control.sdl_mapping.index == axis_idx);
                if let Some(axis) = axis {
                    axis.process(event);
                }
            }
            Event::JoyButtonDown { button_idx, .. } | Event::JoyButtonUp { button_idx, .. } => {
                let button = self
                    .controls
                    .values_mut()
                    .find(|control| control.sdl_mapping.kind == SDLControlKind::Button && control.sdl_mapping.index == button_idx);
                if let Some(button) = button {
                    button.process(event);
                }
            }
            Event::JoyHatMotion { hat_idx, .. } => {
                let hat = self
                    .controls
                    .values_mut()
                    .find(|control| control.sdl_mapping.kind == SDLControlKind::Hat && control.sdl_mapping.index == hat_idx);
                if let Some(hat) = hat {
                    hat.process(event);
                }
            }
            _ => {}
        }
    }
}

impl ControllerManager {
    pub fn new(config: Arc<ConfigLoader>) -> ControllerManager {
        let sdl_context = Arc::new(sdl2::init().unwrap());
        let joystick_subsystem = Arc::new(sdl_context.joystick().unwrap());
        let channel_pair = tokio::sync::broadcast::channel(10000);
        let raw_channel_pair = tokio::sync::broadcast::channel(10000);

        ControllerManager {
            config,
            sdl_context,
            joystick_subsystem,
            devices: HashMap::new(),
            change_event_channel: (Arc::new(channel_pair.0), Arc::new(Mutex::new(channel_pair.1))),
            raw_event_channel: (Arc::new(raw_channel_pair.0), Arc::new(Mutex::new(raw_channel_pair.1))),
        }
    }

    fn handle_joy_device_added(&mut self, event: sdl2::event::Event) {
        match event {
            Event::JoyDeviceAdded { which, .. } => {
                let joystick = self.joystick_subsystem.open(which).unwrap();
                let product_id = unsafe { sdl2_sys::SDL_JoystickGetDeviceProduct(which as i32) };
                let vendor_id = unsafe { sdl2_sys::SDL_JoystickGetDeviceVendor(which as i32) };
                let usb_id: String = format!("{:04X}:{:04X}", vendor_id, product_id);
                println!("Joystick Opened ({}) {}", joystick.name(), usb_id);
                let controller = ControllerManagerController::new(
                    Arc::clone(&self.config),
                    usb_id,
                    joystick,
                    (Arc::clone(&self.change_event_channel.0), Arc::clone(&self.change_event_channel.1)),
                );
                self.devices.insert(which, controller);
            }
            _ => panic!("Invalid event type"),
        }
    }

    fn handle_joy_device_removed(&mut self, event: sdl2::event::Event) {
        match event {
            Event::JoyDeviceRemoved { which, .. } => {
                self.devices.remove(&which);
            }
            _ => panic!("Invalid event type"),
        }
    }

    fn handle_joy_axis_motion(&mut self, event: sdl2::event::Event) {
        match event {
            Event::JoyAxisMotion { which, .. } => {
                if let Some(controller) = self.devices.get_mut(&which) {
                    self.raw_event_channel
                        .0
                        .send(ControllerManagerRawEvent {
                            joystick_index: which,
                            joystick_guid: controller.joystick.guid().string(),
                            joystick_name: controller.joystick.name(),
                            event: event.clone(),
                        })
                        .unwrap();
                    controller.process(event);
                }
            }
            _ => panic!("Invalid event type"),
        }
    }

    fn handle_joy_button_down_or_up(&mut self, event: sdl2::event::Event) {
        match event {
            Event::JoyButtonDown { which, .. } | Event::JoyButtonUp { which, .. } => {
                if let Some(controller) = self.devices.get_mut(&which) {
                    self.raw_event_channel
                        .0
                        .send(ControllerManagerRawEvent {
                            joystick_index: which,
                            joystick_guid: controller.joystick.guid().string(),
                            joystick_name: controller.joystick.name(),
                            event: event.clone(),
                        })
                        .unwrap();

                    controller.process(event);
                }
            }
            _ => panic!("Invalid event type"),
        }
    }

    fn handle_joy_hat_motion(&mut self, event: sdl2::event::Event) {
        match event {
            Event::JoyHatMotion { which, .. } => {
                if let Some(controller) = self.devices.get_mut(&which) {
                    self.raw_event_channel
                        .0
                        .send(ControllerManagerRawEvent {
                            joystick_index: which,
                            joystick_guid: controller.joystick.guid().string(),
                            joystick_name: controller.joystick.name(),
                            event: event.clone(),
                        })
                        .unwrap();

                    controller.process(event);
                }
            }
            _ => panic!("Invalid event type"),
        }
    }

    pub fn receiver(&self) -> Arc<Mutex<Receiver<ControllerManagerChangeEvent>>> {
        Arc::clone(&self.change_event_channel.1)
    }

    pub fn raw_receiver(&self) -> Arc<Mutex<Receiver<ControllerManagerRawEvent>>> {
        Arc::clone(&self.raw_event_channel.1)
    }

    pub fn attach(&mut self, cancel: CancellationToken) {
        let mut event_pump = self.sdl_context.event_pump().unwrap();

        loop {
            if cancel.is_cancelled() {
                break;
            }

            let possible_event = event_pump.poll_event();
            match possible_event {
                Some(event) => {
                    debug!("Event Received: {:?}", event);

                    use sdl2::event::Event;

                    // SDL on windows sends some initial movement events which causes issues
                    let initial_events_threshold = 500;

                    match event {
                        Event::JoyDeviceAdded { .. } => self.handle_joy_device_added(event),
                        Event::JoyDeviceRemoved { .. } => self.handle_joy_device_removed(event),
                        Event::JoyAxisMotion { .. } => {
                            if event.get_timestamp() > initial_events_threshold {
                                self.handle_joy_axis_motion(event);
                            }
                        }
                        Event::JoyButtonDown { .. } | Event::JoyButtonUp { .. } => {
                            if event.get_timestamp() > initial_events_threshold {
                                self.handle_joy_button_down_or_up(event)
                            }
                        }
                        Event::JoyHatMotion { .. } => {
                            if event.get_timestamp() > initial_events_threshold {
                                self.handle_joy_hat_motion(event)
                            }
                        }
                        Event::Quit { .. } => break,
                        _ => {}
                    }
                }
                None => {}
            }
        }
    }

    pub fn subscribe(&self, forwarder: Sender<ControllerManagerChangeEvent>, cancel_token: CancellationToken) -> tokio::task::JoinHandle<()> {
        let mut receiver = self.change_event_channel.0.subscribe();

        tokio::task::spawn(async move {
            loop {
                tokio::select! {
                    _ = cancel_token.cancelled() => { break; }
                    Ok(event) = receiver.recv() => {
                        forwarder.send(event).unwrap();
                    }
                }
            }
        })
    }

    pub fn subscribe_raw(&self, forwarder: Sender<ControllerManagerRawEvent>) -> tokio::task::JoinHandle<()> {
        let mut receiver = self.raw_event_channel.0.subscribe();

        tokio::spawn(async move {
            loop {
                let event = receiver.recv().await.unwrap();
                forwarder.send(event).unwrap();
            }
        })
    }
}
