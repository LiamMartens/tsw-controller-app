package profile_runner

import (
	"tsw_controller_app/controller_mgr"
	"tsw_controller_app/logger"
)

func (p *ProfileRunner) processActiveProfileApiListeners() {
	p.SDLControllerManager.ConfiguredControllers.ForEach(func(controller controller_mgr.SDL_ControllerManager_ConfiguredController, key controller_mgr.DeviceUniqueID) bool {
		selected_profile, has_selected_profile := p.getSelectedProfileForDevice(controller.Device())
		if !has_selected_profile {
			/* skip if no profile selected for controller */
			return true
		}

		for _, listener := range selected_profile.Profile.Listeners {
			if listener.API != nil {
				values, err := p.API.GetByPath(listener.API.Path)
				if err != nil {
					logger.Logger.Error("[ProfileRunner::processActiveProfileApiListeners] failed to get value from API path", "path", listener.API.Path, "error", err)
					continue
				}
			}
		}

		return true
	})
}

func (p *ProfileRunner) processActiveProfileControlListeners(change_event controller_mgr.ControllerManager_Control_ChangeEvent) {
	selected_profile, has_selected_profile := p.getSelectedProfileForDevice(change_event.Device)
	if !has_selected_profile {
		/* skip if no profile selected for controller */
		return
	}

	for _, listener := range selected_profile.Profile.Listeners {
		if listener.Control != nil {
		}
	}
}
