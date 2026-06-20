package profile_runner

import (
	"math"
	"tsw_controller_app/config"
)

func (p *ProfileRunner) processSyncControllerState(sync_control_state SyncController_ControlState) {
	/* sync control only works when a profile is distinctly selected - also skip if not in sync control */
	if sync_control_state.SourceEvent == nil || p.Settings.GetPreferredControlMode() != config.PreferredControlMode_SyncControl {
		return
	}

	selected_profile, has_selected_profile := p.getSelectedProfileForDevice(sync_control_state.SourceEvent.Device)
	if !has_selected_profile {
		/* skip if no profile selected for controller */
		return
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
		return
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
