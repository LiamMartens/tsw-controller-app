use super::serde_sdl_guid::SDLGuid;
use bezier_easing::bezier_easing;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ControllerCalibrationData {
    /** the ID of the controller button or trigger as named in the controller mapping config (see other file - eg: "throttle1", "throttle2", "button1") */
    pub id: String,
    pub deadzone: Option<f32>,
    pub invert: Option<bool>,
    pub min: f32,
    pub max: f32,
    pub idle: f32,
    pub easing_curve: Option<[f32; 4]>,
}

/**
 * This struct defines the controller calibration data that is stored in the config file.
 */
#[derive(Clone, Serialize, Deserialize)]
pub struct ControllerCalibration {
    pub id: SDLGuid,
    pub data: Vec<ControllerCalibrationData>,
}

impl ControllerCalibrationData {
    pub fn new<T: AsRef<str>>(name: T) -> ControllerCalibrationData {
        ControllerCalibrationData {
            id: String::from(name.as_ref()),
            deadzone: None,
            invert: None,
            min: 0.0,
            max: 0.0,
            idle: 0.0,
            easing_curve: None,
        }
    }

    /**
     * Normalize will return None if the value is within the deadzone
     */
    pub fn normalize(&self, incoming_value: i16) -> Option<f32> {
        let idle_range = match self.deadzone {
            Some(deadzone) => [self.idle - deadzone, self.idle + deadzone],
            None => [self.idle, self.idle],
        };

        let value = match self.invert {
            Some(true) => -(incoming_value as f32),
            Some(false) => incoming_value as f32,
            None => incoming_value as f32,
        };

        /* within deadzone - None */
        if value >= idle_range[0] && value <= idle_range[1] {
            return None;
        }

        let easing_curve = match self.easing_curve {
            Some(curve) => curve,
            None => [0.0, 0.0, 1.0, 1.0],
        };
        let ease = bezier_easing(
            easing_curve[0],
            easing_curve[1],
            easing_curve[2],
            easing_curve[3],
        )
        .unwrap();

        /* below idle range -- negative value */
        if value < idle_range[0] && self.min != self.idle {
            let abs_value = ((value - idle_range[0]) / (self.min - idle_range[0]))
                .clamp(0.0, 1.0)
                .abs();
            return Some(ease(abs_value) * -1.0f32);
        }

        let abs_value = ((value - idle_range[1]) / (self.max - idle_range[1]))
            .clamp(0.0, 1.0)
            .abs();
        return Some(ease(abs_value));
    }
}

impl ControllerCalibration {
    pub fn control_data<T: AsRef<str>>(&self, id: T) -> Option<ControllerCalibrationData> {
        match self.data.iter().find(|x| x.id == id.as_ref()) {
            Some(data) => Some(data.clone()),
            None => None,
        }
    }
}
