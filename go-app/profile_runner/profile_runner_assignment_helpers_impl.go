package profile_runner

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"tsw_controller_app/action_sequencer"
	"tsw_controller_app/chan_utils"
	"tsw_controller_app/config"
	"tsw_controller_app/controller_mgr"
	"tsw_controller_app/logger"
)

func (p *ProfileRunner) CallAssignmentActionForControl(
	control_name string,
	assignment_index int,
	controller controller_mgr.IControllerManager_Controller,
	control_state_at_call controller_mgr.ControllerManager_Controller_ControlState,
	assignment config.Config_Controller_Profile_Control_Assignment,
	action *ProfileRunnerAssignmentCall,
) error {
	if action != nil {
		logger.Logger.Debug("[ProfileRunner::CallAssignmentActionForControl] executing assignment action", "sequencer_action", action.ActionSequencerAction, "direct_control_action", action.DirectControlCommand, "api_control_action", action.ApiControlCommand)
	}
	previous_control_assignments_call_list, has_previous_control_call := p.PreviousControlAssignmentCallList.Get(control_name)
	if !has_previous_control_call {
		previous_control_assignments_call_list = &[]*ProfileRunnerAssignmentCall{}
		p.PreviousControlAssignmentCallList.Set(control_name, previous_control_assignments_call_list)
	}
	for len(*previous_control_assignments_call_list) <= assignment_index {
		*previous_control_assignments_call_list = append(*previous_control_assignments_call_list, nil)
	}

	if action == nil && (*previous_control_assignments_call_list)[assignment_index] == nil {
		/* no action and no previous call - don't do anything */
		return fmt.Errorf("no action or previous call list entry")
	}

	/* add updated call entry in previous assignment call list */
	assignment_call := &ProfileRunnerAssignmentCall{
		ControlState:          control_state_at_call,
		ActionSequencerAction: nil,
		VirtualAction:         nil,
		DirectControlCommand:  nil,
		ApiControlCommand:     nil,
	}
	if action != nil {
		assignment_call.ActionSequencerAction = action.ActionSequencerAction
		assignment_call.VirtualAction = action.VirtualAction
		assignment_call.DirectControlCommand = action.DirectControlCommand
		assignment_call.ApiControlCommand = action.ApiControlCommand
	} else {
		/* should always be available - None action should only be set as none for deactivation calls */
		assignment_call.ActionSequencerAction = (*previous_control_assignments_call_list)[assignment_index].ActionSequencerAction
		assignment_call.VirtualAction = (*previous_control_assignments_call_list)[assignment_index].VirtualAction
		assignment_call.DirectControlCommand = (*previous_control_assignments_call_list)[assignment_index].DirectControlCommand
		assignment_call.ApiControlCommand = (*previous_control_assignments_call_list)[assignment_index].ApiControlCommand
	}
	(*previous_control_assignments_call_list)[assignment_index] = assignment_call

	if action != nil {
		if action.ActionSequencerAction != nil {
			logger.Logger.Debug("[ProfileRunner::CallAssignmentActionForControl] queueing sequencer action", "action", action.ActionSequencerAction)
			p.ActionSequencer.Enqueue(*action.ActionSequencerAction)
		} else if action.VirtualAction != nil {
			logger.Logger.Debug("[ProfileRunner::CallAssignmentActionForControl] updating virtual control", "action", action.VirtualAction)
			virtual_control, has_virtual_control := controller.VirtualControls().Get(action.VirtualAction.Control)
			if !has_virtual_control {
				controller.RegisterVirtualControl(action.VirtualAction.Control, action.VirtualAction.Value)
				virtual_control, _ = controller.VirtualControls().Get(action.VirtualAction.Control)
			}
			virtual_control.UpdateValue(action.VirtualAction.Value, false)
			controller.VirtualControls().Set(action.VirtualAction.Control, virtual_control)
		} else if action.DirectControlCommand != nil {
			logger.Logger.Debug("[ProfileRunner::CallAssignmentActionForControl] sending direct control command", "command", action.DirectControlCommand)
			chan_utils.SendTimeout(p.DirectController.ControlChannel, time.Second, *action.DirectControlCommand)
		} else if action.ApiControlCommand != nil {
			logger.Logger.Debug("[ProfileRunner::CallAssignmentActionForControl] sending api control command", "command", action.ApiControlCommand)
			chan_utils.SendTimeout(p.ApiController.ControlChannel, time.Second, *action.ApiControlCommand)
		}
	}
	return nil
}

func (p *ProfileRunner) AssignmentKeysActionToSequencerAction(keys_action config.Config_Controller_Profile_Control_Assignment_Action_Keys, release bool) action_sequencer.ActionSequencerAction {
	var press_time_value float64 = 0
	var wait_time_value float64 = 0
	if keys_action.PressTime != nil {
		press_time_value = *keys_action.PressTime
	}
	if keys_action.WaitTime != nil {
		wait_time_value = *keys_action.WaitTime
	}

	return action_sequencer.ActionSequencerAction{
		Keys:      keys_action.Keys,
		PressTime: press_time_value,
		WaitTime:  wait_time_value,
		Release:   release,
	}
}

func (p *ProfileRunner) AssignmentActionToAssignmentCall(
	control_state controller_mgr.ControllerManager_Controller_ControlState,
	action config.Config_Controller_Profile_Control_Assignment_Action,
	release_if_keys bool,
) *ProfileRunnerAssignmentCall {
	if action.Keys != nil {
		sequencer_action := p.AssignmentKeysActionToSequencerAction(*action.Keys, release_if_keys)
		return &ProfileRunnerAssignmentCall{
			ControlState:          control_state,
			ActionSequencerAction: &sequencer_action,
			VirtualAction:         nil,
			DirectControlCommand:  nil,
			ApiControlCommand:     nil,
		}
	}
	if action.Virtual != nil {
		return &ProfileRunnerAssignmentCall{
			ControlState:          control_state,
			ActionSequencerAction: nil,
			VirtualAction:         action.Virtual,
			DirectControlCommand:  nil,
			ApiControlCommand:     nil,
		}
	}

	preferred_control_mode := p.Settings.GetPreferredControlMode()
	scored_assignment_calls := []ProfileRunner_ScoredAssignmentCallEntry{}

	if action.DirectControl != nil {
		max_change_rate := DEFAULT_MAX_CHANGE_RATE
		enable_api_fallback := action.DirectControl.EnableAPIFallback != nil && *action.DirectControl.EnableAPIFallback
		should_hold := action.DirectControl.Hold != nil && *action.DirectControl.Hold

		if p.DirectController.Connector.IsActive() {
			flags := []string{}
			if should_hold {
				flags = append(flags, "hold")
			}
			if action.DirectControl.MaxChangeRate != nil {
				max_change_rate = *action.DirectControl.MaxChangeRate
			}
			if action.DirectControl.Relative != nil && *action.DirectControl.Relative {
				flags = append(flags, "relative")
			}
			if action.DirectControl.UseNormalized != nil && *action.DirectControl.UseNormalized {
				flags = append(flags, "normalized")
			}
			if action.DirectControl.Notify == nil || *action.DirectControl.Notify {
				flags = append(flags, "notify")
			}

			scored_assignment_call := ProfileRunner_ScoredAssignmentCallEntry{
				Score: 0,
				AssignmentCall: ProfileRunnerAssignmentCall{
					ControlState:          control_state,
					ActionSequencerAction: nil,
					VirtualAction:         nil,
					ApiControlCommand:     nil,
					DirectControlCommand: &DirectController_Command{
						Controls:      action.DirectControl.Controls,
						InputValue:    action.DirectControl.Value,
						MaxChangeRate: max_change_rate,
						Flags:         flags,
					},
				},
			}
			if preferred_control_mode == config.PreferredControlMode_DirectControl {
				scored_assignment_call.Score += 10
			}
			scored_assignment_calls = append(scored_assignment_calls, scored_assignment_call)
		} else if enable_api_fallback && p.ApiController.API.Enabled() {
			scored_assignment_call := ProfileRunner_ScoredAssignmentCallEntry{
				Score: 0,
				AssignmentCall: ProfileRunnerAssignmentCall{
					ControlState:          control_state,
					ActionSequencerAction: nil,
					VirtualAction:         nil,
					ApiControlCommand: &ApiController_Command{
						Controls:      action.DirectControl.Controls,
						InputValue:    action.DirectControl.Value,
						Hold:          should_hold,
						MaxChangeRate: max_change_rate,
					},
				},
			}
			if preferred_control_mode == config.PreferredControlMode_ApiControl {
				scored_assignment_call.Score += 10
			}
			scored_assignment_calls = append(scored_assignment_calls, scored_assignment_call)
		}
	}

	if action.ApiControl != nil && p.ApiController.API.CanConnect() {
		max_change_rate := DEFAULT_MAX_CHANGE_RATE
		hold := false
		if action.ApiControl.MaxChangeRate != nil {
			max_change_rate = *action.ApiControl.MaxChangeRate
		}
		if action.ApiControl.Hold != nil {
			hold = *action.ApiControl.Hold
		}
		scored_assignment_call := ProfileRunner_ScoredAssignmentCallEntry{
			Score: 0,
			AssignmentCall: ProfileRunnerAssignmentCall{
				ControlState:          control_state,
				ActionSequencerAction: nil,
				VirtualAction:         nil,
				DirectControlCommand:  nil,
				ApiControlCommand: &ApiController_Command{
					Controls:      action.ApiControl.Controls,
					InputValue:    action.ApiControl.ApiValue,
					MaxChangeRate: max_change_rate,
					Hold:          hold,
				},
			},
		}
		if preferred_control_mode == config.PreferredControlMode_ApiControl {
			scored_assignment_call.Score += 10
		}
		scored_assignment_calls = append(scored_assignment_calls, scored_assignment_call)
	}

	sort.Slice(scored_assignment_calls, func(i, j int) bool {
		return scored_assignment_calls[i].Score > scored_assignment_calls[j].Score
	})
	if len(scored_assignment_calls) > 0 {
		return &scored_assignment_calls[0].AssignmentCall
	}

	return nil
}

func (p *ProfileRunner) GetAssignments(
	control *config.Config_Controller_Profile_Control,
	source_event *controller_mgr.ControllerManager_Control_ChangeEvent,
) []config.Config_Controller_Profile_Control_Assignment {
	assignments := control.GetAssignments()

	/* filter out conditional assignments */
	current_rail_class := p.CabDebugger.State.DrivableActorName
	preferred_control_mode := p.Settings.GetPreferredControlMode()
	can_use_direct_control_mode := p.DirectController.Connector.IsActive()
	can_use_sync_control_mode := p.SyncController.Connector.IsActive()
	can_use_api_control_mode := p.ApiController.API.CanConnect()

	non_control_asssignments := []config.Config_Controller_Profile_Control_Assignment{}
	scored_control_assignments := map[config.PreferredControlMode]*ProfileRunner_ScoredAssignmentsListEntry{}
	scored_control_assignments[config.PreferredControlMode_DirectControl] = &ProfileRunner_ScoredAssignmentsListEntry{Score: 0, Assignments: []config.Config_Controller_Profile_Control_Assignment{}}
	scored_control_assignments[config.PreferredControlMode_ApiControl] = &ProfileRunner_ScoredAssignmentsListEntry{Score: 0, Assignments: []config.Config_Controller_Profile_Control_Assignment{}}
	scored_control_assignments[config.PreferredControlMode_SyncControl] = &ProfileRunner_ScoredAssignmentsListEntry{Score: 0, Assignments: []config.Config_Controller_Profile_Control_Assignment{}}

collect_assignments_loop:
	for _, assignment := range assignments {
		assignment_rail_class_information := assignment.RailClassInformation()
		if assignment_rail_class_information != nil &&
			len(*assignment_rail_class_information) > 0 {
			/* should check rail class information */
			if current_rail_class == "" {
				continue collect_assignments_loop
			}

			does_match_rail_class := false
			for _, rc := range *assignment_rail_class_information {
				if rc.ClassName == current_rail_class {
					does_match_rail_class = true
					break
				}
			}
			if !does_match_rail_class {
				continue collect_assignments_loop
			}
		}

		/* conditions can only be evaluated if there is a source event */
		assigmment_conditions := assignment.Conditions()
		if source_event != nil && assigmment_conditions != nil && len(*assigmment_conditions) > 0 {
			for _, condition := range *assigmment_conditions {
				var dependecy_control controller_mgr.IControllerManager_Controller_Control = nil
				if strings.HasPrefix(condition.Control, "virtual:") {
					/* virtual controls always exist - they just start at 0 */
					virtual_control, has_dependency_control := source_event.Controller.VirtualControls().Get(condition.Control)
					if !has_dependency_control {
						source_event.Controller.RegisterVirtualControl(condition.Control, 0.0)
						virtual_control, _ = source_event.Controller.VirtualControls().Get(condition.Control)
					}
					dependecy_control = virtual_control
				} else if joy_control, has_dependency_control := source_event.Controller.Controls().Get(condition.Control); has_dependency_control {
					dependecy_control = joy_control
				}

				if dependecy_control == nil {
					logger.Logger.Error("[ProfileRunner::GetAssignments] skipping assignment because dependency control does not exist")
					continue collect_assignments_loop
				}

				state := dependecy_control.GetState()
				switch condition.Operator {
				case "gte":
					if state.NormalizedValues.Value < condition.Value {
						/* condition doesn't match -> skip */
						continue collect_assignments_loop
					}
				case "lte":
					if state.NormalizedValues.Value > condition.Value {
						/* condition doesn't match -> skip */
						continue collect_assignments_loop
					}
				case "gt":
					if state.NormalizedValues.Value <= condition.Value {
						/* condition doesn't match -> skip */
						continue collect_assignments_loop
					}
				case "lt":
					if state.NormalizedValues.Value >= condition.Value {
						/* condition doesn't match -> skip */
						continue collect_assignments_loop
					}
				case "eq":
					if state.NormalizedValues.Value != condition.Value {
						/* condition doesn't match -> skip */
						continue collect_assignments_loop
					}
				}
			}
		}

		if assignment.DirectControl != nil {
			enable_api_fallback := assignment.DirectControl.EnableAPIFallback != nil && *assignment.DirectControl.EnableAPIFallback
			scored_control_assignments[config.PreferredControlMode_DirectControl].Assignments = append(scored_control_assignments[config.PreferredControlMode_DirectControl].Assignments, assignment)
			/* if enabaled as API fallback; also add this assignment to the scored list of API controls */
			if enable_api_fallback {
				scored_control_assignments[config.PreferredControlMode_ApiControl].Assignments = append(scored_control_assignments[config.PreferredControlMode_ApiControl].Assignments, assignment)
			}
		} else if assignment.SyncControl != nil {
			scored_control_assignments[config.PreferredControlMode_SyncControl].Assignments = append(scored_control_assignments[config.PreferredControlMode_SyncControl].Assignments, assignment)
		} else if assignment.ApiControl != nil {
			scored_control_assignments[config.PreferredControlMode_ApiControl].Assignments = append(scored_control_assignments[config.PreferredControlMode_ApiControl].Assignments, assignment)
		} else {
			non_control_asssignments = append(non_control_asssignments, assignment)
		}
	}

	/*
		the scoring is very simple;
		- DC gets 3 points if available
		- API gets 2 points if available
		- Sync gets 1 point if available
		-- Any of these gets 5 points if available and preferred
		--> this means that whichever is preferred and available always gets the most points
		--> if the preferred mode is not available it will fallback to the internally preferred methods of DC, API and Sync
	*/
	if can_use_direct_control_mode {
		scored_control_assignments[config.PreferredControlMode_DirectControl].Score += 3
		if preferred_control_mode == config.PreferredControlMode_DirectControl {
			scored_control_assignments[config.PreferredControlMode_DirectControl].Score += 5
		}
	}
	if can_use_api_control_mode {
		/* this will also mark direct control assignments which have API fallback enabled with scores */
		scored_control_assignments[config.PreferredControlMode_ApiControl].Score += 2
		if preferred_control_mode == config.PreferredControlMode_ApiControl {
			scored_control_assignments[config.PreferredControlMode_ApiControl].Score += 5
		}
	}
	if can_use_sync_control_mode {
		scored_control_assignments[config.PreferredControlMode_SyncControl].Score += 1
		if preferred_control_mode == config.PreferredControlMode_SyncControl {
			scored_control_assignments[config.PreferredControlMode_SyncControl].Score += 5
		}
	}

	/* only check control type assignments if the connector is alive or the API is available */
	if can_use_api_control_mode || can_use_direct_control_mode || can_use_sync_control_mode {
		scored_control_assignments_values_list := []*ProfileRunner_ScoredAssignmentsListEntry{}
		for _, entry := range scored_control_assignments {
			if len(entry.Assignments) > 0 {
				scored_control_assignments_values_list = append(scored_control_assignments_values_list, entry)
			}
		}
		sort.Slice(scored_control_assignments_values_list, func(i, j int) bool {
			return scored_control_assignments_values_list[i].Score > scored_control_assignments_values_list[j].Score
		})
		if len(scored_control_assignments_values_list) > 0 {
			return append(scored_control_assignments_values_list[0].Assignments, non_control_asssignments...)
		}
	} else {
		logger.Logger.Debug("no socket or API connection is available - skipping direct/sync and API control")
	}

	return non_control_asssignments
}
