package profile_runner

import (
	"tsw_controller_app/action_sequencer"
	"tsw_controller_app/config"
	"tsw_controller_app/controller_mgr"
	"tsw_controller_app/logger"
)

func (p *ProfileRunner) processControllerChangeEvent(change_event controller_mgr.ControllerManager_Control_ChangeEvent) {
	logger.Logger.Debug("[ProfileRunner::Run] received change event", "event", change_event)

	selected_profile, has_selected_profile := p.getSelectedProfileForDevice(change_event.Device)
	if !has_selected_profile {
		logger.Logger.Debug("[ProfileRunner::Run] skipping event, no profile selected", "event", change_event)
		return
	} else {
		logger.Logger.Debug("[ProfileRunner::Run] executing profile", "name", selected_profile.Profile.Name)
	}

	control_name := change_event.ControlName
	device_id := change_event.Device.DeviceID
	defined_thresholds := map[string]float64{}

	if selected_profile.Profile.Controller != nil && selected_profile.Profile.Controller.Mapping != nil {
		if joy_control, is_joy_control := change_event.Control.(*controller_mgr.SDL_ControllerManager_Controller_JoyControl); is_joy_control {
			root_mapping := joy_control.SDLMapping()
			calibration := joy_control.Calibration()
			override_mapping := selected_profile.Profile.Controller.Mapping
			override_control, find_override_control_err := override_mapping.FindByKindAndIndex(root_mapping.Kind, root_mapping.Index)
			if find_override_control_err == nil {
				control_name = override_control.Name
			}
			/* collect the named thresholds and their values */
			for _, threshold := range calibration.Thresholds {
				defined_thresholds[threshold.Name] = threshold.Value
			}
		}
	}

	control_profile := selected_profile.Profile.FindControlByName(device_id, control_name)
	if control_profile == nil {
		logger.Logger.Debug("[ProfileRunner::Run] skipping event, control not found in profile", "event", change_event)
		return
	}

	assignments := p.GetAssignments(control_profile, &change_event)

	previous_control_assignments_call_list, has_previous_control_assignments_call_list := p.PreviousControlAssignmentCallList.Get(control_name)
	for assignment_index, control_assignment_item := range assignments {
		logger.Logger.Debug("[ProfileRunner::Run] executing assignment", "assignment", control_assignment_item)
		var previous_assignment_call *ProfileRunnerAssignmentCall = nil
		if has_previous_control_assignments_call_list && len(*previous_control_assignments_call_list) > assignment_index {
			previous_assignment_call = (*previous_control_assignments_call_list)[assignment_index]
		}

		if control_assignment_item.Momentary != nil {
			if control_assignment_item.Momentary.IsMatch(change_event.ControlState.NormalizedValues.Value) {
				// call if there was no prior call or if the prior call was not this threshold
				should_call_activation := previous_assignment_call == nil || !control_assignment_item.Momentary.IsMatch(previous_assignment_call.ControlState.NormalizedValues.Value)
				if should_call_activation {
					action_to_call := p.AssignmentActionToAssignmentCall(change_event.ControlState, control_assignment_item.Momentary.ActionActivate, false)
					p.CallAssignmentActionForControl(control_name, assignment_index, change_event.Controller, change_event.ControlState, control_assignment_item, action_to_call)
				}
			} else if previous_assignment_call != nil && control_assignment_item.Momentary.IsMatch(previous_assignment_call.ControlState.NormalizedValues.Value) {
				// when below the threshold only call action if the last call was above or equal to the threshold
				if control_assignment_item.Momentary.ActionDeactivate != nil {
					action_to_call := p.AssignmentActionToAssignmentCall(change_event.ControlState, *control_assignment_item.Momentary.ActionDeactivate, false)
					p.CallAssignmentActionForControl(control_name, assignment_index, change_event.Controller, change_event.ControlState, control_assignment_item, action_to_call)
				} else if control_assignment_item.Momentary.ActionActivate.Keys != nil {
					/* only release if keys -> can't "release" direct control actions */
					action_to_call := p.AssignmentActionToAssignmentCall(change_event.ControlState, control_assignment_item.Momentary.ActionActivate, true)
					p.CallAssignmentActionForControl(control_name, assignment_index, change_event.Controller, change_event.ControlState, control_assignment_item, action_to_call)
				} else {
					/* clear previuous call so momentary can be re-triggered */
					p.CallAssignmentActionForControl(control_name, assignment_index, change_event.Controller, change_event.ControlState, control_assignment_item, nil)
				}
			}
		}
		if control_assignment_item.Linear != nil {
			initial_state_value := control_assignment_item.Linear.CalculateNeutralizedValue(change_event.ControlState.NormalizedValues.InitialValue)
			control_state_value := control_assignment_item.Linear.CalculateNeutralizedValue(change_event.ControlState.NormalizedValues.Value)
			var thresholds_currently_exceeding []config.Config_Controller_Profile_Control_Assignment_Linear_Threshold
			var thresholds_previously_passed []config.Config_Controller_Profile_Control_Assignment_Linear_Threshold
			for _, threshold := range control_assignment_item.Linear.GenerateThresholds(defined_thresholds) {
				if threshold.IsExceedingThreshold(control_state_value, defined_thresholds) {
					thresholds_currently_exceeding = append(thresholds_currently_exceeding, threshold)
				}
				/* threshold was previously passed if the last assignment call was exceeding the threshold OR if there was no last call if the initial value exceeded it*/
				if previous_assignment_call != nil && threshold.IsExceedingThreshold(
					control_assignment_item.Linear.CalculateNeutralizedValue(previous_assignment_call.ControlState.NormalizedValues.Value),
					defined_thresholds,
				) || previous_assignment_call == nil && threshold.IsExceedingThreshold(initial_state_value, defined_thresholds) {
					thresholds_previously_passed = append(thresholds_previously_passed, threshold)
				}
			}

			if len(thresholds_currently_exceeding) > len(thresholds_previously_passed) {
				// activate the intermediate thresholds
				thresholds_to_activate := thresholds_currently_exceeding[len(thresholds_previously_passed):]
				for _, threshold := range thresholds_to_activate {
					action_to_call := p.AssignmentActionToAssignmentCall(change_event.ControlState, threshold.ActionActivate, false)
					p.CallAssignmentActionForControl(control_name, assignment_index, change_event.Controller, change_event.ControlState, control_assignment_item, action_to_call)
				}
			} else if len(thresholds_currently_exceeding) < len(thresholds_previously_passed) {
				// deactivate the intermediate thresholds by iterating from end of previously passed up until but not including the currently exceeding threshold
				for i := len(thresholds_previously_passed) - 1; i > len(thresholds_currently_exceeding)-1; i-- {
					threshold := thresholds_previously_passed[i]
					if threshold.ActionDeactivate != nil {
						action_to_call := p.AssignmentActionToAssignmentCall(change_event.ControlState, *threshold.ActionDeactivate, false)
						p.CallAssignmentActionForControl(control_name, assignment_index, change_event.Controller, change_event.ControlState, control_assignment_item, action_to_call)
					} else if threshold.ActionActivate.Keys != nil {
						/* only release if keys -> can't "release" direct control actions */
						action_to_call := p.AssignmentActionToAssignmentCall(change_event.ControlState, threshold.ActionActivate, true)
						p.CallAssignmentActionForControl(control_name, assignment_index, change_event.Controller, change_event.ControlState, control_assignment_item, action_to_call)
					} else {
						/* clear previuous call so threshold can be re-triggered */
						p.CallAssignmentActionForControl(control_name, assignment_index, change_event.Controller, change_event.ControlState, control_assignment_item, nil)
					}
				}
			}
		}
		if control_assignment_item.Toggle != nil {
			if control_assignment_item.Toggle.IsMatch(change_event.ControlState.NormalizedValues.Value) {
				// call if there was no prior call or if the prior call was not this threshold
				action_to_call := p.AssignmentActionToAssignmentCall(change_event.ControlState, control_assignment_item.Toggle.ActionActivate, false)
				if previous_assignment_call != nil && action_to_call.IsSameAction(previous_assignment_call) {
					/* if the previous call is the same as the activation call -> toggle to deactivation action */
					action_to_call = p.AssignmentActionToAssignmentCall(change_event.ControlState, control_assignment_item.Toggle.ActionDeactivate, false)
				}
				p.CallAssignmentActionForControl(control_name, assignment_index, change_event.Controller, change_event.ControlState, control_assignment_item, action_to_call)
			} else if previous_assignment_call != nil && control_assignment_item.Toggle.IsMatch(previous_assignment_call.ControlState.NormalizedValues.Value) && previous_assignment_call.ActionSequencerAction != nil {
				// when below the threshold only call action if the last call was above or equal to the threshold
				// this is only used for releasing key actions
				p.CallAssignmentActionForControl(control_name, assignment_index, change_event.Controller, change_event.ControlState, control_assignment_item, &ProfileRunnerAssignmentCall{
					ControlState: change_event.ControlState,
					ActionSequencerAction: &action_sequencer.ActionSequencerAction{
						Keys:      previous_assignment_call.ActionSequencerAction.Keys,
						PressTime: previous_assignment_call.ActionSequencerAction.PressTime,
						WaitTime:  previous_assignment_call.ActionSequencerAction.WaitTime,
						Release:   true,
					},
					ApiControlCommand:    nil,
					DirectControlCommand: nil,
				})
			}
		}
		if control_assignment_item.DirectControl != nil {
			control_value := change_event.Control.GetState().NormalizedValues.Value
			if control_assignment_item.DirectControl.ControlRange != nil {
				control_value = control_assignment_item.DirectControl.ControlRange.Clamp(control_value)
			}
			output_value := control_assignment_item.DirectControl.InputValue.CalculateOutputValue(control_value, defined_thresholds)

			max_change_rate := DEFAULT_MAX_CHANGE_RATE
			should_hold := control_assignment_item.DirectControl.Hold != nil && *control_assignment_item.DirectControl.Hold
			enable_api_fallback := control_assignment_item.DirectControl.EnableAPIFallback != nil && *control_assignment_item.DirectControl.EnableAPIFallback

			flags := []string{}
			if should_hold {
				flags = append(flags, "hold")
			}
			if control_assignment_item.DirectControl.InputValue.MaxChangeRate != nil {
				max_change_rate = *control_assignment_item.DirectControl.InputValue.MaxChangeRate
			}
			if control_assignment_item.DirectControl.Notify == nil || *control_assignment_item.DirectControl.Notify {
				flags = append(flags, "notify")
			}
			if control_assignment_item.DirectControl.UseNormalized != nil && *control_assignment_item.DirectControl.UseNormalized {
				flags = append(flags, "normalized")
			}
			if output_value != nil {
				assignment_call := &ProfileRunnerAssignmentCall{
					ControlState:          change_event.ControlState,
					ActionSequencerAction: nil,
					ApiControlCommand:     nil,
					DirectControlCommand: &DirectController_Command{
						Controls:      control_assignment_item.DirectControl.Controls,
						InputValue:    *output_value,
						MaxChangeRate: max_change_rate,
						Flags:         flags,
					},
				}
				if enable_api_fallback && !p.DirectController.Connector.IsActive() {
					assignment_call.DirectControlCommand = nil
					assignment_call.ApiControlCommand = &ApiController_Command{
						Controls:      control_assignment_item.DirectControl.Controls,
						InputValue:    *output_value,
						MaxChangeRate: max_change_rate,
						Hold:          should_hold,
					}
				}
				p.CallAssignmentActionForControl(control_name, assignment_index, change_event.Controller, change_event.ControlState, control_assignment_item, assignment_call)
			}
		}
		if control_assignment_item.ApiControl != nil {
			max_change_rate := DEFAULT_MAX_CHANGE_RATE
			hold := false
			control_value := change_event.Control.GetState().NormalizedValues.Value
			if control_assignment_item.ApiControl.InputValue.MaxChangeRate != nil {
				max_change_rate = *control_assignment_item.ApiControl.InputValue.MaxChangeRate
			}
			if control_assignment_item.ApiControl.ControlRange != nil {
				control_value = control_assignment_item.ApiControl.ControlRange.Clamp(control_value)
			}
			if control_assignment_item.ApiControl.Hold != nil {
				hold = *control_assignment_item.ApiControl.Hold
			}

			output_value := control_assignment_item.ApiControl.InputValue.CalculateOutputValue(control_value, defined_thresholds)
			if output_value != nil {
				p.CallAssignmentActionForControl(control_name, assignment_index, change_event.Controller, change_event.ControlState, control_assignment_item, &ProfileRunnerAssignmentCall{
					ControlState:          change_event.ControlState,
					ActionSequencerAction: nil,
					DirectControlCommand:  nil,
					ApiControlCommand: &ApiController_Command{
						Controls:      control_assignment_item.ApiControl.Controls,
						InputValue:    *output_value,
						MaxChangeRate: max_change_rate,
						Hold:          hold,
					},
				})
			}
		}
		if control_assignment_item.SyncControl != nil {
			control_value := change_event.Control.GetState().NormalizedValues.Value
			if control_assignment_item.SyncControl.ControlRange != nil {
				control_value = control_assignment_item.SyncControl.ControlRange.Clamp(control_value)
			}
			output_value := control_assignment_item.SyncControl.InputValue.CalculateOutputValue(control_value, defined_thresholds)
			if output_value != nil {
				p.SyncController.UpdateControlStateTargetValue(
					control_assignment_item.SyncControl.Identifier,
					*output_value,
					control_assignment_item.SyncControl,
					&change_event,
				)
			}
		}
	}
}
