package profile_runner

import (
	"sync"
	"tsw_controller_app/action_sequencer"
	"tsw_controller_app/cabdebugger"
	"tsw_controller_app/config"
	"tsw_controller_app/controller_mgr"
	"tsw_controller_app/map_utils"
	"tsw_controller_app/tswapi"
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
	API                               tswapi.ITSWAPI
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
