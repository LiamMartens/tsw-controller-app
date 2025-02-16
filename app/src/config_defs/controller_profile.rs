use serde::{Deserialize, Serialize};

use super::serde_sdl_guid::SDLGuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type", rename_all = "snake_case")]
pub enum ControllerProfileControlAssignment {
    Momentary(ControllerProfileControlMomentaryAssignment),
    Linear(ControllerProfileControlLinearAssignment),
    Toggle(ControllerProfileControlToggleAssignment),
    DirectControl(ControllerProfileDirectControlAssignment),
    SyncControl(ControllerProfileDirectControAssignmentSyncMode),
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ControllerProfileControlAssignmentKeysAction {
    pub keys: String,
    pub press_time: Option<f32>,
    pub wait_time: Option<f32>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ControllerProfileDirectControlAssignmentInputValue {
    pub min: f32,
    pub max: f32,
    pub step: Option<f32>,
    pub invert: Option<bool>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ControllerProfileDirectControAssignmentSyncMode {
    /** this is the VHID Identifier Name - differs from the direct control name */
    pub identifier: String,
    pub action_increase: ControllerProfileControlAssignmentKeysAction,
    pub action_decrease: ControllerProfileControlAssignmentKeysAction,
}

/* defines a direct UE4ss control -> through websockets */
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ControllerProfileDirectControlAssignment {
    pub controls: String, /* the HID control component as per the UE4SS API */
    pub input_value: ControllerProfileDirectControlAssignmentInputValue,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ControllerProfileControlAssignmentDirectControlAction {
    pub controls: String,
    pub value: f32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(untagged)]
pub enum ControllerProfileControlAssignmentAction {
    Keys(ControllerProfileControlAssignmentKeysAction),
    DirectControl(ControllerProfileControlAssignmentDirectControlAction),
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ControllerProfileControlToggleAssignment {
    pub threshold: f32,
    pub action_activate: ControllerProfileControlAssignmentAction,
    pub action_deactivate: ControllerProfileControlAssignmentAction,
}
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ControllerProfileControlMomentaryAssignment {
    pub threshold: f32,
    pub action_activate: ControllerProfileControlAssignmentAction,
    pub action_deactivate: Option<ControllerProfileControlAssignmentAction>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ControllerProfileControlLinearAssignmentThreshold {
    pub value: f32,
    pub value_end: Option<f32>,
    pub value_step: Option<f32>,
    pub action_activate: ControllerProfileControlAssignmentAction,
    pub action_deactivate: Option<ControllerProfileControlAssignmentAction>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ControllerProfileControlLinearAssignment {
    pub neutral: Option<f32>,
    pub thresholds: Vec<ControllerProfileControlLinearAssignmentThreshold>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ControllerProfileControl {
    pub name: String,
    pub assignment: Option<ControllerProfileControlAssignment>,
    pub assignments: Option<Vec<ControllerProfileControlAssignment>>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ControllerProfile {
    pub name: String,
    pub controls: Vec<ControllerProfileControl>,
    pub controller_id: Option<SDLGuid>,
}

impl ControllerProfileControlAssignmentAction {
    pub fn get_compare_value(&self) -> String {
        match self {
            ControllerProfileControlAssignmentAction::Keys(action) => format!("{}", action.keys),
            ControllerProfileControlAssignmentAction::DirectControl(action) => {
                format!("{}:{}", action.controls, action.value)
            }
        }
    }
}

impl ControllerProfileControlLinearAssignmentThreshold {
    pub fn is_exceeding_threshold(&self, value: f32) -> bool {
        if self.value < 0.0 {
            return value < self.value;
        }
        return value >= self.value;
    }
}

impl ControllerProfileControlLinearAssignment {
    pub fn generated_thresholds(&self) -> Vec<ControllerProfileControlLinearAssignmentThreshold> {
        let mut thresholds: Vec<ControllerProfileControlLinearAssignmentThreshold> = Vec::new();
        for threshold in self.thresholds.iter() {
            if threshold.value_end.is_none() || threshold.value_step.is_none() {
                thresholds.push(threshold.clone());
            } else {
                let mut current_value = threshold.value;
                while current_value <= threshold.value_end.unwrap() {
                    thresholds.push(ControllerProfileControlLinearAssignmentThreshold {
                        value: current_value,
                        value_end: threshold.value_end,
                        value_step: threshold.value_step,
                        action_activate: threshold.action_activate.clone(),
                        action_deactivate: threshold.action_deactivate.clone(),
                    });
                    current_value = ((current_value + threshold.value_step.unwrap()) * 10000.0)
                        .round()
                        / 10000.0;
                }
            }
        }
        thresholds
    }

    pub fn calculate_neutralized_value(&self, value: f32) -> f32 {
        if self.neutral.is_some() && self.neutral.unwrap() > 0.0 {
            return (value - self.neutral.unwrap()) * (1.0 / self.neutral.unwrap());
        }
        return value;
    }
}

impl ControllerProfileDirectControlAssignmentInputValue {
    /**
     * The incoming value here can only be [-1, 1]
     */
    pub fn calculate_normal_value(&self, value: f32) -> f32 {
        let input_value: f32 = match self.invert {
            Some(true) => match value < 0.0 {
                true => -1.0 - value,
                false => 1.0 - value,
            },
            _ => value,
        };
        let total_distance = (self.max - self.min).abs();
        let normal = (input_value * total_distance) + self.min;
        match self.step {
            Some(step) => {
                let step_count = (normal / step).round();
                return (step_count * step).clamp(self.min, self.max);
            }
            None => normal.clamp(self.min, self.max),
        }
    }
}

impl ControllerProfile {
    pub fn find_control<T: AsRef<str>>(&self, name: T) -> Option<&ControllerProfileControl> {
        self.controls.iter().find(|c| c.name == name.as_ref())
    }
}
