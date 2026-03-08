package profile_runner

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
	"tsw_controller_app/action_sequencer"
	"tsw_controller_app/cabdebugger"
	"tsw_controller_app/chan_utils"
	"tsw_controller_app/config"
	"tsw_controller_app/controller_mgr"
	"tsw_controller_app/logger"
	"tsw_controller_app/map_utils"
)

type ProfileRunner_AssignmentScore = int

const DEFAULT_MAX_CHANGE_RATE = 999.0
const ASSIGNMENT_SCORE_IS_PREFERRED_CONTROL_MODE ProfileRunner_AssignmentScore = 10
const ASSIGNMENT_SCORE_DIRECT_CONTROL_MODE ProfileRunner_AssignmentScore = 3
const ASSIGNMENT_SCORE_API_CONTROL_MODE ProfileRunner_AssignmentScore = 2
const ASSIGNMENT_SCORE_SYNC_CONTROL_MODE ProfileRunner_AssignmentScore = 1

type ProfileRunner_ScoredProfileEntry struct {
	Id    string
	Score int
}

type ProfileRunner_ScoredAssignmentsListEntry struct {
	Score       int
	Assignments []config.Config_Controller_Profile_Control_Assignment
}

type ProfileRunner_ScoredAssignmentCallEntry struct {
	Score          int
	AssignmentCall ProfileRunnerAssignmentCall
}

type ProfileRunnerSettings_SelectedProfile struct {
	Profile config.Config_Controller_Profile
}

type ProfileRunnerSettings struct {
	Mutex                      sync.RWMutex
	SelectedProfilesByUniqueID *map_utils.LockMap[controller_mgr.DeviceUniqueID, ProfileRunnerSettings_SelectedProfile]
	PreferredControlMode       config.PreferredControlMode
}

type ProfileRunnerAssignmentCall struct {
	ControlState          controller_mgr.ControllerManager_Controller_ControlState
	ActionSequencerAction *action_sequencer.ActionSequencerAction
	VirtualAction         *config.Config_Controller_Profile_Control_Assignment_Action_Virtual
	DirectControlCommand  *DirectController_Command
	ApiControlCommand     *ApiController_Command
}

type ProfileRunner struct {
	ActionSequencer                   *action_sequencer.ActionSequencer
	SDLControllerManager              *controller_mgr.SDLControllerManager
	VirtualControllerManager          *controller_mgr.VirtualControllerManager
	DirectController                  *DirectController
	SyncController                    *SyncController
	ApiController                     *ApiController
	CabDebugger                       *cabdebugger.CabDebugger
	Profiles                          *map_utils.LockMap[string, config.Config_Controller_Profile]
	Settings                          ProfileRunnerSettings
	PreviousControlAssignmentCallList *map_utils.LockMap[string, *[]*ProfileRunnerAssignmentCall]
}

func (s *ProfileRunnerSettings) Update(mutator func(s *ProfileRunnerSettings)) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	mutator(s)
}

func (s *ProfileRunnerSettings) GetSelectedProfiles() *map_utils.LockMap[controller_mgr.DeviceUniqueID, ProfileRunnerSettings_SelectedProfile] {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return s.SelectedProfilesByUniqueID
}

func (s *ProfileRunnerSettings) GetPreferredControlMode() config.PreferredControlMode {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return s.PreferredControlMode
}

func (s *ProfileRunnerSettings) SetPreferredControlMode(mode config.PreferredControlMode) {
	s.Mutex.Lock()
	s.PreferredControlMode = mode
	defer s.Mutex.Unlock()
}

func New(
	action_sequencer *action_sequencer.ActionSequencer,
	sdl_controller_manager *controller_mgr.SDLControllerManager,
	virtual_controller_manager *controller_mgr.VirtualControllerManager,
	direct_controller *DirectController,
	sync_controller *SyncController,
	api_controller *ApiController,
	cab_debugger *cabdebugger.CabDebugger,
) *ProfileRunner {
	return &ProfileRunner{
		ActionSequencer:          action_sequencer,
		SDLControllerManager:     sdl_controller_manager,
		VirtualControllerManager: virtual_controller_manager,
		DirectController:         direct_controller,
		SyncController:           sync_controller,
		ApiController:            api_controller,
		CabDebugger:              cab_debugger,
		Profiles:                 map_utils.NewLockMap[string, config.Config_Controller_Profile](),
		Settings: ProfileRunnerSettings{
			Mutex:                      sync.RWMutex{},
			SelectedProfilesByUniqueID: map_utils.NewLockMap[controller_mgr.DeviceUniqueID, ProfileRunnerSettings_SelectedProfile](),
			PreferredControlMode:       config.PreferredControlMode_DirectControl,
		},
		PreviousControlAssignmentCallList: map_utils.NewLockMap[string, *[]*ProfileRunnerAssignmentCall](),
	}
}

func (pc *ProfileRunnerAssignmentCall) IsSameAction(other *ProfileRunnerAssignmentCall) bool {
	if pc.ActionSequencerAction != nil && other.ActionSequencerAction != nil {
		return pc.ActionSequencerAction.Keys == other.ActionSequencerAction.Keys
	}
	if pc.VirtualAction != nil && other.VirtualAction != nil {
		return pc.VirtualAction.Control == other.VirtualAction.Control && pc.VirtualAction.Value == other.VirtualAction.Value
	}
	if pc.DirectControlCommand != nil && other.DirectControlCommand != nil {
		return pc.DirectControlCommand.ToSocketMessage().ToString() == other.DirectControlCommand.ToSocketMessage().ToString()
	}
	if pc.ApiControlCommand != nil && other.ApiControlCommand != nil {
		return pc.ApiControlCommand.Controls == other.ApiControlCommand.Controls && pc.ApiControlCommand.InputValue == other.ApiControlCommand.InputValue
	}
	return false
}

func (p *ProfileRunner) getSelectedProfileForDevice(device *controller_mgr.ControllerManager_ChangeEvent_Device) (ProfileRunnerSettings_SelectedProfile, bool) {
	selected_profile, has_selected_profile := p.Settings.GetSelectedProfiles().Get(device.UniqueID)

	/* try auto-selection */
	current_rail_class := p.CabDebugger.State.DrivableActorName
	if !has_selected_profile && current_rail_class != "" {
		scored_profiles := []ProfileRunner_ScoredProfileEntry{}

		p.Profiles.ForEach(func(profile config.Config_Controller_Profile, id string) bool {
			if (profile.AutoSelect == nil || !*profile.AutoSelect) ||
				profile.RailClassInformation == nil ||
				(profile.Controller != nil && *profile.Controller.UsbID != device.DeviceID) {
				/* skip if not-autoselect, rail class information is missing or the controller doesn't match */
				return true
			}

			/* we'll score any match which is not embedded higher than their embedded counterpart */
			score_factor := 1
			if !profile.Metadata.IsEmbedded {
				score_factor = 10
			}

			for _, rc_info := range *profile.RailClassInformation {
				if rc_info.ClassName == current_rail_class {
					is_controller_match := profile.Controller != nil && *profile.Controller.UsbID == device.DeviceID
					if is_controller_match {
						scored_profiles = append(scored_profiles, ProfileRunner_ScoredProfileEntry{Id: id, Score: 20 * score_factor})
					} else {
						scored_profiles = append(scored_profiles, ProfileRunner_ScoredProfileEntry{Id: id, Score: 10 * score_factor})
					}
					break
				}
			}

			return true
		})
		sort.Slice(scored_profiles, func(i, j int) bool {
			return scored_profiles[i].Score > scored_profiles[j].Score
		})

		if len(scored_profiles) > 0 {
			profile, _ := p.Profiles.Get(scored_profiles[0].Id)
			has_selected_profile = true
			selected_profile = ProfileRunnerSettings_SelectedProfile{
				Profile: profile,
			}
		}
	}

	return selected_profile, has_selected_profile
}

func (p *ProfileRunner) GetProfileNameToIdMap() map[string][]string {
	id_map_by_name := map[string][]string{}
	p.Profiles.ForEach(func(profile config.Config_Controller_Profile, id string) bool {
		if existing_ids, has_key := id_map_by_name[profile.Name]; has_key {
			id_map_by_name[profile.Name] = append(existing_ids, id)
		} else {
			id_map_by_name[profile.Name] = []string{id}
		}
		return true
	})
	return id_map_by_name
}

func (p *ProfileRunner) RegisterProfile(profile config.Config_Controller_Profile) {
	p.Profiles.Set(profile.Id(), profile)
}

func (p *ProfileRunner) Resolve() {
	/* resolves all the profiles */
	id_name_map := p.GetProfileNameToIdMap()
	p.Profiles.Mutex.Lock()
	defer p.Profiles.Mutex.Unlock()

	var resolve_profile func(profile config.Config_Controller_Profile) config.Config_Controller_Profile
	resolve_profile = func(profile config.Config_Controller_Profile) config.Config_Controller_Profile {
		if profile.Extends != nil && len(*profile.Extends) > 0 && profile.Name != *profile.Extends {
			if extend_from_profile_ids, has_extendable_ids := id_name_map[*profile.Extends]; has_extendable_ids {
				if len(extend_from_profile_ids) == 0 || len(extend_from_profile_ids) > 1 {
					/* only extend if there is one and only one profile to extend from */
					return profile
				}
				extend_from_profile := p.Profiles.Map[extend_from_profile_ids[0]]

				/*
					these are the control names which are defined in the profile we are currently resolving;
					these should be kept as they already have a definition
				*/
				existing_control_definitions := map[string]bool{}
				for _, control := range profile.Controls {
					existing_control_definitions[control.Name] = true
				}

				resolved_extend_from_profile := resolve_profile(extend_from_profile)
				for _, control := range resolved_extend_from_profile.Controls {
					if _, should_not_override := existing_control_definitions[control.Name]; !should_not_override {
						profile.Controls = append(profile.Controls, control)
					}
				}

				if profile.Controller == nil && extend_from_profile.Controller != nil {
					profile.Controller = extend_from_profile.Controller
				}
			}
		}
		return profile
	}

	for profile_id, profile := range p.Profiles.Map {
		p.Profiles.Map[profile_id] = resolve_profile(profile)
	}
}

func (p *ProfileRunner) ClearProfile(unique_id controller_mgr.DeviceUniqueID) {
	p.Settings.Update(func(s *ProfileRunnerSettings) {
		s.SelectedProfilesByUniqueID.Delete(unique_id)
	})
}

func (p *ProfileRunner) SetProfile(unique_id controller_mgr.DeviceUniqueID, id string) error {
	var err error = nil
	p.Settings.Update(func(s *ProfileRunnerSettings) {
		profile, is_valid_profile := p.Profiles.Get(id)
		if is_valid_profile {
			s.SelectedProfilesByUniqueID.Set(unique_id, ProfileRunnerSettings_SelectedProfile{
				Profile: profile,
			})
		} else {
			err = fmt.Errorf("could not find profile by ID %s", id)
		}
	})
	return err
}

func (p *ProfileRunner) SetPreferredControlMode(mode config.PreferredControlMode) {
	p.Settings.Update(func(s *ProfileRunnerSettings) {
		s.PreferredControlMode = mode
	})
}

func (p *ProfileRunner) CallAssignmentActionForControl(
	control_name string,
	assignment_index int,
	controller controller_mgr.IControllerManager_Controller,
	control_state_at_call controller_mgr.ControllerManager_Controller_ControlState,
	assignment config.Config_Controller_Profile_Control_Assignment,
	action *ProfileRunnerAssignmentCall,
) error {
	if action != nil {
		logger.Logger.Info("[ProfileRunner::CallAssignmentActionForControl] executing assignment action", "sequencer_action", action.ActionSequencerAction, "direct_control_action", action.DirectControlCommand, "api_control_action", action.ApiControlCommand)
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
		logger.Logger.Info("no socket or API connection is available - skipping direct/sync and API control")
	}

	return non_control_asssignments
}

func (p *ProfileRunner) Run(ctx context.Context) context.CancelFunc {
	/*
		the runner handles a few different things:
		1.Listen to the controller manager and send the appropriate values to the sequencer or direct controller
		2. Listen to the sync controller and sequence the appropriate actions to reach the target value
	*/
	context_with_cancel, cancel := context.WithCancel(ctx)

	/* normal action sequencing */
	go func() {
		sdl_channel, sdl_unsubscribe := p.SDLControllerManager.SubscribeChangeEvent()
		virtual_channel, virtual_unsubscribe := p.VirtualControllerManager.SubscribeChangeEvent()
		defer sdl_unsubscribe()
		defer virtual_unsubscribe()

		var handleChangeEvent func(change_event controller_mgr.ControllerManager_Control_ChangeEvent)
		handleChangeEvent = func(change_event controller_mgr.ControllerManager_Control_ChangeEvent) {
			logger.Logger.Debug("[ProfileRunner::Run] received change event", "event", change_event)

			selected_profile, has_selected_profile := p.getSelectedProfileForDevice(change_event.Device)
			if !has_selected_profile {
				logger.Logger.Debug("[ProfileRunner::Run] skipping event, no profile selected", "event", change_event)
				return
			}

			control_name := change_event.ControlName
			device_id := change_event.Device.DeviceID
			if selected_profile.Profile.Controller != nil && selected_profile.Profile.Controller.Mapping != nil {
				if joy_control, is_joy_control := change_event.Control.(*controller_mgr.SDL_ControllerManager_Controller_JoyControl); is_joy_control {
					root_mapping := joy_control.SDLMapping()
					override_mapping := selected_profile.Profile.Controller.Mapping
					override_control, find_override_control_err := override_mapping.FindByKindAndIndex(root_mapping.Kind, root_mapping.Index)
					if find_override_control_err == nil {
						control_name = override_control.Name
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
					for _, threshold := range control_assignment_item.Linear.GenerateThresholds() {
						if threshold.IsExceedingThreshold(control_state_value) {
							thresholds_currently_exceeding = append(thresholds_currently_exceeding, threshold)
						}
						/* threshold was previously passed if the last assignment call was exceeding the threshold OR if there was no last call if the initial value exceeded it*/
						if previous_assignment_call != nil && threshold.IsExceedingThreshold(
							control_assignment_item.Linear.CalculateNeutralizedValue(previous_assignment_call.ControlState.NormalizedValues.Value),
						) || previous_assignment_call == nil && threshold.IsExceedingThreshold(initial_state_value) {
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
					output_value := control_assignment_item.DirectControl.InputValue.CalculateOutputValue(control_value)

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
					output_value := control_assignment_item.ApiControl.InputValue.CalculateOutputValue(control_value)
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
					output_value := control_assignment_item.SyncControl.InputValue.CalculateOutputValue(control_value)
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

		for {
			select {
			case <-context_with_cancel.Done():
				return
			case change_event := <-virtual_channel:
				handleChangeEvent(change_event)
			case change_event := <-sdl_channel:
				handleChangeEvent(change_event)
			}
		}
	}()

	/* sync control action sequencing */
	go func() {
		channel, unsubscribe := p.SyncController.Subscribe()
		defer unsubscribe()

		for {
			select {
			case <-context_with_cancel.Done():
				return
			case sync_control_state := <-channel:
				/* sync control only works when a profile is distinctly selected - also skip if not in sync control */
				if sync_control_state.SourceEvent == nil || p.Settings.GetPreferredControlMode() != config.PreferredControlMode_SyncControl {
					continue
				}

				selected_profile, has_selected_profile := p.getSelectedProfileForDevice(sync_control_state.SourceEvent.Device)
				if !has_selected_profile {
					/* skip if no profile selected for controller */
					continue
				}

				var sync_control_assignment *config.Config_Controller_Profile_Control_Assignment = nil
			control_loop:
				for _, cp := range selected_profile.Profile.Controls {
					assignments := p.GetAssignments(&cp, sync_control_state.SourceEvent)
					for _, assignment := range assignments {
						if assignment.SyncControl != nil && assignment.SyncControl.Identifier == sync_control_state.Identifier {
							sync_control_assignment = &assignment
							break control_loop
						}
					}
				}

				/* only act if a sync control assignment exists for this identifier and is the current preferred control mode */
				if sync_control_assignment == nil {
					continue
				}

				const MARGIN_OF_ERROR = 0.005
				should_stop_moving := (
				/* was increasing and has now exceeded value */
				sync_control_state.CurrentValue >= sync_control_state.TargetValue && sync_control_state.Moving == 1 ||
					/* was decreasing and has now subceeded value */
					sync_control_state.CurrentValue <= sync_control_state.TargetValue && sync_control_state.Moving == -1 ||
					/* otherwise is within margin of error and was moving */
					math.Abs(sync_control_state.CurrentValue-sync_control_state.TargetValue) < MARGIN_OF_ERROR && sync_control_state.Moving != 0)
				should_start_increasing := sync_control_state.TargetValue > sync_control_state.CurrentValue && math.Abs(sync_control_state.TargetValue-sync_control_state.CurrentValue) > MARGIN_OF_ERROR && sync_control_state.Moving == 0
				should_start_decreasing := sync_control_state.TargetValue < sync_control_state.CurrentValue && math.Abs(sync_control_state.TargetValue-sync_control_state.CurrentValue) > MARGIN_OF_ERROR && sync_control_state.Moving == 0

				release_previous_action := func() {
					if sync_control_state.Moving == -1 {
						p.ActionSequencer.Enqueue(p.AssignmentKeysActionToSequencerAction(sync_control_assignment.SyncControl.ActionDecrease, true))
					} else {
						p.ActionSequencer.Enqueue(p.AssignmentKeysActionToSequencerAction(sync_control_assignment.SyncControl.ActionIncrease, true))
					}
				}

				if should_stop_moving {
					release_previous_action()
					p.SyncController.UpdateControlStateMoving(sync_control_state.Identifier, 0)
				}

				if should_start_increasing {
					release_previous_action()
					p.ActionSequencer.Enqueue(p.AssignmentKeysActionToSequencerAction(sync_control_assignment.SyncControl.ActionIncrease, false))
					p.SyncController.UpdateControlStateMoving(sync_control_state.Identifier, 1)
				}

				if should_start_decreasing {
					release_previous_action()
					p.ActionSequencer.Enqueue(p.AssignmentKeysActionToSequencerAction(sync_control_assignment.SyncControl.ActionDecrease, false))
					p.SyncController.UpdateControlStateMoving(sync_control_state.Identifier, -1)
				}
			}
		}
	}()

	return cancel
}
