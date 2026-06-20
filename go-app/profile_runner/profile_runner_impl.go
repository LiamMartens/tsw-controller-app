package profile_runner

import (
	"context"
	"sync"
	"time"
	"tsw_controller_app/action_sequencer"
	"tsw_controller_app/cabdebugger"
	"tsw_controller_app/config"
	"tsw_controller_app/controller_mgr"
	"tsw_controller_app/map_utils"
	"tsw_controller_app/tswapi"
)

func NewProfileRunner(
	api tswapi.ITSWAPI,
	action_sequencer *action_sequencer.ActionSequencer,
	sdl_controller_manager *controller_mgr.SDLControllerManager,
	virtual_controller_manager *controller_mgr.VirtualControllerManager,
	direct_controller *DirectController,
	sync_controller *SyncController,
	api_controller *ApiController,
	cab_debugger *cabdebugger.CabDebugger,
) *ProfileRunner {
	return &ProfileRunner{
		API:                      api,
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

func (p *ProfileRunner) SetPreferredControlMode(mode config.PreferredControlMode) {
	p.Settings.Update(func(s *ProfileRunnerSettings) {
		s.PreferredControlMode = mode
	})
}

func (p *ProfileRunner) Run(ctx context.Context) context.CancelFunc {
	/*
		the runner handles a few different things:
		1. Listen to the controller manager and send the appropriate values to the sequencer or direct controller
		2. Listen to the sync controller and sequence the appropriate actions to reach the target value
		3. Trigger any listener actions as necessary
	*/
	context_with_cancel, cancel := context.WithCancel(ctx)

	go func() {
		ticker := time.NewTicker(333 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-context_with_cancel.Done():
				return
			case <-ticker.C:
				go p.processActiveProfileApiListeners()
			}
		}
	}()

	/* normal action sequencing */
	go func() {
		sdl_channel, sdl_unsubscribe := p.SDLControllerManager.SubscribeChangeEvent()
		virtual_channel, virtual_unsubscribe := p.VirtualControllerManager.SubscribeChangeEvent()
		defer sdl_unsubscribe()
		defer virtual_unsubscribe()

		for {
			select {
			case <-context_with_cancel.Done():
				return
			case change_event := <-virtual_channel:
				p.processControllerChangeEvent(change_event)
				go p.processActiveProfileControlListeners(change_event)
			case change_event := <-sdl_channel:
				p.processControllerChangeEvent(change_event)
				go p.processActiveProfileControlListeners(change_event)
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
				p.processSyncControllerState(sync_control_state)
			}
		}
	}()

	return cancel
}
