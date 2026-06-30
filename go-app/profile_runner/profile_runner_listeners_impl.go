package profile_runner

import (
	"errors"
	"fmt"
	"strings"
	"tsw_controller_app/config"
	"tsw_controller_app/controller_mgr"
	"tsw_controller_app/logger"
	"tsw_controller_app/sdl_mgr"
)

var ErrNoAction = errors.New("no action defined")
var ErrNoHIDDevice = errors.New("no hid device")
var ErrHIDFailure = errors.New("hid failure")

var ErrInvalidAction = errors.New("invalid action definition")
var ErrNoListenerValue = errors.New("could not find listener value")

func (p *ProfileRunner) executeProfileListenerAction(
	device controller_mgr.IControllerManager_Device,
	action config.Config_Controller_Profile_Listener_Action,
) error {
	if action.HIDOutputReport != nil {
		sdl_device, is_sdl_device := device.(*sdl_mgr.SDLMgr_Joystick)
		/* hid reports are only supported in sdl devices which have an associated HID device */
		if !is_sdl_device || sdl_device.HIDDevice == nil {
			logger.Logger.Error("[ProfileRunner::executeProfileListenerAction] unable to execute hid_output_report action on non SDL or non HID device")
			return fmt.Errorf("unable to execute HID Output Report action due to missing SDL device: %w", ErrNoHIDDevice)
		}

		report, err := sdl_device.HIDDevice.GetOutputReport(byte(action.HIDOutputReport.ReportID), uint8(len(action.HIDOutputReport.Mask)))
		if err != nil {
			logger.Logger.Error("[ProfileRunner::executeProfileListenerAction] could not read output report", "id", action.HIDOutputReport.ReportID, "error", err)
			return fmt.Errorf("could not read output report: %w: %w", ErrHIDFailure, err)
		}

		for ix := range report {
			if action.HIDOutputReport.Operation == "and" {
				report[ix] = report[ix] & action.HIDOutputReport.Mask[ix]
			} else if action.HIDOutputReport.Operation == "or" {
				report[ix] = report[ix] | action.HIDOutputReport.Mask[ix]
			}
		}

		err = sdl_device.HIDDevice.SetOutputReport(byte(action.HIDOutputReport.ReportID), report)
		if err != nil {
			logger.Logger.Error("[ProfileRunner::executeProfileListenerAction] could not send output report", "id", action.HIDOutputReport.ReportID, "error", err)
			return fmt.Errorf("could not send output report: %w: %w", ErrHIDFailure, err)
		}

		return nil
	}

	if action.HIDFeatureReport != nil {
		sdl_device, is_sdl_device := device.(*sdl_mgr.SDLMgr_Joystick)
		/* hid reports are only supported in sdl devices which have an associated HID device */
		if !is_sdl_device || sdl_device.HIDDevice == nil {
			logger.Logger.Error("[ProfileRunner::executeProfileListenerAction] unable to execute hid_output_report action on non SDL or non HID device")
			return fmt.Errorf("unable to execute HID Output Report action due to missing SDL device: %w", ErrNoHIDDevice)
		}

		report, err := sdl_device.HIDDevice.ReadFeatureReport(byte(action.HIDFeatureReport.ReportID), uint8(len(action.HIDFeatureReport.Mask)))
		if err != nil {
			logger.Logger.Error("[ProfileRunner::executeProfileListenerAction] could not read feature report", "id", action.HIDFeatureReport.ReportID, "error", err)
			return fmt.Errorf("could not read feature report: %w: %w", ErrHIDFailure, err)
		}

		for ix := range report {
			if action.HIDFeatureReport.Operation == "and" {
				report[ix] = report[ix] & action.HIDFeatureReport.Mask[ix]
			} else if action.HIDFeatureReport.Operation == "or" {
				report[ix] = report[ix] | action.HIDFeatureReport.Mask[ix]
			}
		}

		err = sdl_device.HIDDevice.SendFeatureReport(byte(action.HIDFeatureReport.ReportID), report)
		if err != nil {
			logger.Logger.Error("[ProfileRunner::executeProfileListenerAction] could not send feature report", "id", action.HIDFeatureReport.ReportID, "error", err)
			return fmt.Errorf("could not send feature report: %w: %w", ErrHIDFailure, err)
		}

		return nil
	}

	return fmt.Errorf("no valid action to execute: %w", ErrNoAction)
}

func (p *ProfileRunner) filterMatchingAPIListenerActions(listener *config.Config_Controller_Profile_Listener) ([]config.Config_Controller_Profile_Listener_Action, error) {
	if listener.API == nil {
		return nil, fmt.Errorf("listener has invalid action definition: %w", ErrInvalidAction)
	}

	values, err := p.API.GetByPath(listener.API.Path)
	if err != nil {
		logger.Logger.Error("[ProfileRunner::processActiveProfileApiListeners] failed to get value from API path", "path", listener.API.Path, "error", err)
		return nil, err
	}

	value, has_value := values[listener.API.ValuesKey]
	if !has_value {
		/* can't do anything if there is no value */
		return nil, fmt.Errorf("could not find listener value: %w", ErrNoListenerValue)
	}

	matching_actions := []config.Config_Controller_Profile_Listener_Action{}
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
		matching_actions = append(matching_actions, action)
	}

	return matching_actions, nil
}

func (p *ProfileRunner) filterMatchingControlListenerActions(
	listener *config.Config_Controller_Profile_Listener,
	controller controller_mgr.IControllerManager_Controller,
) ([]config.Config_Controller_Profile_Listener_Action, error) {
	if listener.Control == nil {
		return nil, fmt.Errorf("listener has invalid action definition: %w", ErrInvalidAction)
	}

	var dependecy_control controller_mgr.IControllerManager_Controller_Control = nil
	if strings.HasPrefix(listener.Control.Name, "virtual:") {
		/* virtual controls always exist - they just start at 0 */
		virtual_control, has_dependency_control := controller.VirtualControls().Get(listener.Control.Name)
		if !has_dependency_control {
			controller.RegisterVirtualControl(listener.Control.Name, 0.0)
			virtual_control, _ = controller.VirtualControls().Get(listener.Control.Name)
		}
		dependecy_control = virtual_control
	} else if joy_control, has_dependency_control := controller.Controls().Get(listener.Control.Name); has_dependency_control {
		dependecy_control = joy_control
	}
	control_value := dependecy_control.GetState().NormalizedValues.Value

	matching_actions := []config.Config_Controller_Profile_Listener_Action{}
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
		matching_actions = append(matching_actions, action)
	}

	return matching_actions, nil
}

func (p *ProfileRunner) filterMatchingCabStateListenerActions(listener *config.Config_Controller_Profile_Listener) ([]config.Config_Controller_Profile_Listener_Action, error) {
	if listener.CabState == nil {
		return nil, fmt.Errorf("listener has invalid action definition: %w", ErrInvalidAction)
	}

	value, has_value := p.CabDebugger.State.Controls.Get(listener.CabState.Name)
	if !has_value {
		/* can't do anything if there is no value */
		return nil, fmt.Errorf("could not find listener value: %w", ErrNoListenerValue)
	}

	matching_actions := []config.Config_Controller_Profile_Listener_Action{}
action_loop:
	for _, action := range listener.CabState.Actions {
		conditions := action.GetConditions()
		for _, condition := range conditions {
			if !condition.Matches(value.CurrentValue) {
				/* if any condition does not match skip action */
				continue action_loop
			}
		}

		/* if all conditions passed - execute action (might need to deduplicate actions?) */
		matching_actions = append(matching_actions, action)
	}

	return matching_actions, nil
}

func (p *ProfileRunner) processActiveProfileApiListeners() {
	p.SDLControllerManager.ConfiguredControllers.ForEach(func(controller controller_mgr.SDL_ControllerManager_ConfiguredController, key controller_mgr.DeviceUniqueID) bool {
		selected_profile, has_selected_profile := p.getSelectedProfileForDevice(controller.Device())
		if !has_selected_profile {
			/* skip if no profile selected for controller */
			return true
		}

		for _, listener := range selected_profile.Profile.Listeners {
			if listener.API == nil {
				continue
			}

			actions, err := p.filterMatchingAPIListenerActions(&listener)
			if err != nil {
				logger.Logger.Error("[ProfileRunner::processActiveProfileApiListeners] could not filter API listener actions", "error", err)
				continue
			}

			for _, action := range actions {
				go p.executeProfileListenerAction(controller.Device(), action)
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
		if listener.Control == nil {
			continue
		}

		actions, err := p.filterMatchingControlListenerActions(&listener, change_event.Controller)
		if err != nil {
			logger.Logger.Error("[ProfileRunner::processActiveProfileApiListeners] could not filter control listener actions", "error", err)
			continue
		}

		for _, action := range actions {
			go p.executeProfileListenerAction(change_event.Controller.Device(), action)
		}
	}
}

func (p *ProfileRunner) processActiveProfileCabStateListeners() {
	p.SDLControllerManager.ConfiguredControllers.ForEach(func(controller controller_mgr.SDL_ControllerManager_ConfiguredController, key controller_mgr.DeviceUniqueID) bool {
		selected_profile, has_selected_profile := p.getSelectedProfileForDevice(controller.Device())
		if !has_selected_profile {
			/* skip if no profile selected for controller */
			return true
		}

		for _, listener := range selected_profile.Profile.Listeners {
			if listener.CabState == nil {
				continue
			}

			actions, err := p.filterMatchingCabStateListenerActions(&listener)
			if err != nil {
				logger.Logger.Error("[ProfileRunner::processActiveProfileApiListeners] could not filter Cab State listener actions", "error", err)
				continue
			}

			for _, action := range actions {
				go p.executeProfileListenerAction(controller.Device(), action)
			}
		}

		return true
	})
}
