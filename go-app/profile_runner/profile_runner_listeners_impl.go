package profile_runner

import (
	"strings"
	"tsw_controller_app/config"
	"tsw_controller_app/controller_mgr"
	"tsw_controller_app/logger"

	"github.com/goforj/godump"
)

func (p *ProfileRunner) executeProfileListenerAction(action config.Config_Controller_Profile_Listener_Action) {
	godump.Dump("executing listener action", action)
}

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

				value, has_value := values[listener.API.ValuesKey]
				if !has_value {
					/* can't do anything if there is no value */
					continue
				}

			action_loop:
				for _, action := range listener.API.Actions {
					conditions := action.GetConditions()
					for _, condition := range conditions {
						if !condition.Matches(value) {
							/* if any condition does not match skip action */
							continue action_loop
						}
					}

					/* if all conditions passed - execute action (might need to deduplicate actions?) */
					go p.executeProfileListenerAction(action)
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
			var dependecy_control controller_mgr.IControllerManager_Controller_Control = nil
			if strings.HasPrefix(listener.Control.Name, "virtual:") {
				/* virtual controls always exist - they just start at 0 */
				virtual_control, has_dependency_control := change_event.Controller.VirtualControls().Get(listener.Control.Name)
				if !has_dependency_control {
					change_event.Controller.RegisterVirtualControl(listener.Control.Name, 0.0)
					virtual_control, _ = change_event.Controller.VirtualControls().Get(listener.Control.Name)
				}
				dependecy_control = virtual_control
			} else if joy_control, has_dependency_control := change_event.Controller.Controls().Get(listener.Control.Name); has_dependency_control {
				dependecy_control = joy_control
			}
			control_value := dependecy_control.GetState().NormalizedValues.Value

		action_loop:
			for _, action := range listener.Control.Actions {
				conditions := action.GetConditions()
				for _, condition := range conditions {
					if !condition.Matches(control_value) {
						/* if any condition does not match skip action */
						continue action_loop
					}
				}

				/* if all conditions passed - execute action (might need to deduplicate actions?) */
				go p.executeProfileListenerAction(action)
			}
		}
	}
}
