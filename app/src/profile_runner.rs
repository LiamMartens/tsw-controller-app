use std::{collections::HashMap, sync::Arc};

use tokio::sync::{broadcast::Sender, Mutex};

use crate::{
    action_sequencer::ActionSequencer,
    config_defs::controller_profile::{
        ControllerProfileControlAssignment, ControllerProfileControlAssignmentAction,
        ControllerProfileControlLinearAssignmentThreshold,
    },
    config_loader::ConfigLoader,
    controller_manager::{ControllerManagerChangeEvent, ControllerManagerControllerControlState},
    direct_controller::DirectControlCommand,
};

#[derive(Clone)]
pub struct ProfileRunnerAssignmentCall {
    pub control_name: String,
    pub control_state: ControllerManagerControllerControlState,
    pub assignment: ControllerProfileControlAssignment,
    pub action: ControllerProfileControlAssignmentAction,
}

pub struct ProfileRunner {
    config: Arc<ConfigLoader>,
    sequencer: Arc<ActionSequencer>,
    direct_control_sender: Arc<Mutex<Sender<DirectControlCommand>>>,
    profile_name: Option<String>,
    /* keeps track of the last called assignments */
    control_calls: HashMap<String, ProfileRunnerAssignmentCall>,
}

impl ProfileRunner {
    pub fn new(
        config: Arc<ConfigLoader>,
        sequencer: Arc<ActionSequencer>,
        direct_control_sender: Arc<Mutex<Sender<DirectControlCommand>>>,
    ) -> ProfileRunner {
        ProfileRunner {
            config,
            sequencer,
            direct_control_sender,
            profile_name: None,
            control_calls: HashMap::new(),
        }
    }

    pub fn reset_profile(&mut self) -> Result<(), String> {
        self.profile_name = None;
        return Ok(());
    }

    pub fn set_profile<T: AsRef<str>>(&mut self, name: T) -> Result<(), String> {
        let name = name.as_ref();
        if &Some(String::from(name)) == &self.profile_name {
            return Ok(());
        }

        let profile = self.config.find_controller_profile(name, None);
        match profile {
            Some(_) => {
                self.profile_name = Some(name.to_string());
                Ok(())
            }
            None => Err(format!("Profile {} not found", name)),
        }
    }

    pub async fn call_assignment_action<T: AsRef<str>>(
        &mut self,
        control_name: T,
        control_state: &ControllerManagerControllerControlState,
        control_assignment: &ControllerProfileControlAssignment,
        assignment_action: &ControllerProfileControlAssignmentAction,
        should_release_keys: Option<bool>,
    ) {
        match assignment_action {
            ControllerProfileControlAssignmentAction::Keys(action) => {
                let sequencer_action = super::action_sequencer::ActionSequencerAction {
                    keys: action.keys.clone(),
                    press_time: action.press_time,
                    wait_time: action.wait_time,
                    release: Some(should_release_keys.unwrap_or(false)),
                };
                self.control_calls.insert(
                    String::from(control_name.as_ref()),
                    ProfileRunnerAssignmentCall {
                        control_name: String::from(control_name.as_ref()),
                        control_state: control_state.clone(),
                        assignment: control_assignment.clone(),
                        action: assignment_action.clone(),
                    },
                );
                self.sequencer.add_action(sequencer_action).await;
            }
            ControllerProfileControlAssignmentAction::DirectControl(action) => {
                self.control_calls.insert(
                    String::from(control_name.as_ref()),
                    ProfileRunnerAssignmentCall {
                        control_name: String::from(control_name.as_ref()),
                        control_state: control_state.clone(),
                        assignment: control_assignment.clone(),
                        action: assignment_action.clone(),
                    },
                );

                let direct_control_sender = self.direct_control_sender.lock().await;
                direct_control_sender
                    .send(DirectControlCommand {
                        controls: action.controls.clone(),
                        input_value: action.value,
                    })
                    .unwrap();
            }
        }
    }

    pub async fn run(&mut self, event: ControllerManagerChangeEvent) {
        if !event.has_changed() || !self.profile_name.is_some() {
            return;
        }

        let config_loader = Arc::clone(&self.config);
        let controller_config = config_loader.find_controller_profile(
            self.profile_name.as_ref().unwrap(),
            Some(event.joystick_guid),
        );

        if controller_config.is_none() {
            return;
        }

        let control_name = event.control_name.clone();
        let control_state = event.control_state.clone();
        let control_definition = controller_config
            .unwrap()
            .find_control(control_name.clone());
        let last_called_assignment = match self.control_calls.get(&control_name.clone()) {
            Some(last_call) => Some(last_call.clone()),
            None => None,
        };

        match control_definition {
            Some(control) => match &control.assignment {
                ControllerProfileControlAssignment::Momentary(assignment) => {
                    if control_state.value >= assignment.threshold {
                        // call if there was no prior call or if the prior call was not this threshold
                        let should_call = last_called_assignment.as_ref().is_none()
                            || last_called_assignment.clone().unwrap().control_state.value
                                < assignment.threshold;
                        if should_call {
                            self.call_assignment_action(
                                control_name.clone(),
                                &control_state,
                                &control.assignment,
                                &assignment.action_activate,
                                None,
                            )
                            .await;
                        }
                    } else if last_called_assignment.as_ref().is_some()
                        && last_called_assignment.clone().unwrap().control_state.value
                            >= assignment.threshold
                    {
                        // when below the threshold only call action if the last call was above or equal to the threshold
                        self.call_assignment_action(
                            control_name.clone(),
                            &control_state,
                            &control.assignment,
                            match &assignment.action_deactivate {
                                Some(action) => action,
                                None => &assignment.action_activate,
                            },
                            None,
                        )
                        .await;
                    }
                }
                ControllerProfileControlAssignment::Linear(assignment) => {
                    let control_state_value =
                        assignment.calculate_neutralized_value(control_state.value);
                    let generated_thresholds = assignment.generated_thresholds();
                    let thresholds: Vec<&ControllerProfileControlLinearAssignmentThreshold> =
                        generated_thresholds
                            .iter()
                            .filter(|t| {
                                if control_state_value < 0.0 {
                                    return t.value < 0.0;
                                }
                                return t.value >= 0.0;
                            })
                            .collect();

                    let exceeding_thresholds: Vec<
                        &&ControllerProfileControlLinearAssignmentThreshold,
                    > = thresholds
                        .iter()
                        .filter(|t| t.is_exceeding_threshold(control_state_value))
                        .collect();
                    let thresholds_passed: Vec<
                        &&ControllerProfileControlLinearAssignmentThreshold,
                    > = match last_called_assignment.as_ref() {
                        Some(last_call) => thresholds
                            .iter()
                            .filter(|t| {
                                t.is_exceeding_threshold(
                                    assignment
                                        .calculate_neutralized_value(last_call.control_state.value),
                                )
                            })
                            .collect(),
                        None => {
                            /* if there was no last call we'll consider all thresholds passed up until the initial value */
                            let thresholds: Vec<
                                &&ControllerProfileControlLinearAssignmentThreshold,
                            > = thresholds
                                .iter()
                                .filter(|t| control_state.initial_value >= t.value)
                                .collect();
                            thresholds
                        }
                    };
                    if exceeding_thresholds.len() > thresholds_passed.len() {
                        // activate the intermediate thresholds
                        let thresholds_to_activate =
                            &thresholds[thresholds_passed.len()..exceeding_thresholds.len()];
                        for threshold in thresholds_to_activate {
                            self.call_assignment_action(
                                control_name.clone(),
                                &control_state,
                                &control.assignment,
                                &threshold.action_activate,
                                None,
                            )
                            .await;
                        }
                    } else if exceeding_thresholds.len() < thresholds_passed.len() {
                        // deactivate the intermediate thresholds
                        let thresholds_to_deactivate: &Vec<
                            &&ControllerProfileControlLinearAssignmentThreshold,
                        > = &thresholds[exceeding_thresholds.len()..thresholds_passed.len()]
                            .iter()
                            .rev()
                            .collect();
                        for threshold in thresholds_to_deactivate {
                            self.call_assignment_action(
                                control_name.clone(),
                                &control_state,
                                &control.assignment,
                                match &threshold.action_deactivate {
                                    Some(action) => action,
                                    None => &threshold.action_activate,
                                },
                                None,
                            )
                            .await;
                        }
                    }
                }
                ControllerProfileControlAssignment::Toggle(assignment) => {
                    if control_state.value >= assignment.threshold {
                        let action_to_call = match last_called_assignment.as_ref() {
                            Some(last_call) => {
                                /* if last call was below threshold -> activate */
                                match &last_call.action {
                                    ControllerProfileControlAssignmentAction::Keys(action) => match &assignment.action_activate {
                                        ControllerProfileControlAssignmentAction::Keys(activate_action) => match activate_action.keys == action.keys {
                                            true => &assignment.action_deactivate,
                                            false => &assignment.action_activate,
                                        }
                                        ControllerProfileControlAssignmentAction::DirectControl(_) => {
                                            &assignment.action_activate
                                        }
                                    }
                                    ControllerProfileControlAssignmentAction::DirectControl(
                                        action,
                                    ) =>  match &assignment.action_activate {
                                        ControllerProfileControlAssignmentAction::Keys(_) => &assignment.action_activate,
                                        ControllerProfileControlAssignmentAction::DirectControl(action_activate) => match action_activate.value == action.value {
                                            true => &assignment.action_deactivate,
                                            false => &assignment.action_activate,
                                        }
                                    }
                                }
                            }
                            None => &assignment.action_activate,
                        };

                        self.call_assignment_action(
                            control_name.clone(),
                            &control_state,
                            &control.assignment,
                            action_to_call,
                            None,
                        )
                        .await;
                    } else if last_called_assignment.as_ref().is_some()
                        && last_called_assignment.clone().unwrap().control_state.value
                            >= assignment.threshold
                    {
                        let last_action_taken = last_called_assignment.clone().unwrap().action;
                        // when below the threshold only call action if the last call was above or equal to the threshold
                        self.call_assignment_action(
                            control_name.clone(),
                            &control_state,
                            &control.assignment,
                            &last_action_taken.clone(),
                            Some(true),
                        )
                        .await;
                    }
                }
                ControllerProfileControlAssignment::DirectControl(assignment) => {
                    let input_value = assignment
                        .input_value
                        .calculate_normal_value(control_state.value);
                    let direct_control_sender = self.direct_control_sender.lock().await;
                    direct_control_sender
                        .send(DirectControlCommand {
                            controls: assignment.controls.clone(),
                            input_value,
                        })
                        .unwrap();
                }
            },
            None => {}
        }
    }
}
