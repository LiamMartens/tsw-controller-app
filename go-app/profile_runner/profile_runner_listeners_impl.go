package profile_runner

import (
	"encoding/binary"
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

func (p *ProfileRunner) executeProfileListenerAction(
	device controller_mgr.IControllerManager_Device,
	action config.Config_Controller_Profile_Listener_Action,
) error {
	if action.HIDReport != nil {
		sdl_device, is_sdl_device := device.(*sdl_mgr.SDLMgr_Joystick)
		/* hid reports are only supported in sdl devices which have an associated HID device */
		if !is_sdl_device || sdl_device.HIDDevice == nil {
			logger.Logger.Error("[ProfileRunner::executeProfileListenerAction] unable to execute hid_output_report action on non SDL or non HID device")
			return fmt.Errorf("unable to execute HID Output Report action due to missing SDL device: %w", ErrNoHIDDevice)
		}

		report, err := sdl_device.HIDDevice.ReadFeatureReport(byte(action.HIDReport.ReportID))
		if err != nil {
			logger.Logger.Error("[ProfileRunner::executeProfileListenerAction] could not read feature report", "id", action.HIDReport.ReportID, "error", err)
			return fmt.Errorf("could not read feature report: %w: %w", ErrHIDFailure, err)
		}

		mask := make([]byte, 8)
		/*
		 * Little-endian: least significant byte goes first (index 0), most significant byte goes last (index 7)
		 * This means the mask is applied per byte and any more significant bytes are thrown away - this makes it easy
		 * to create a feature report mask even if the feature report only returns a singular byte
		 */
		binary.LittleEndian.PutUint64(mask, action.HIDReport.Mask)
		for ix := range report {
			if action.HIDReport.Operation == "and" {
				report[ix] = report[ix] & mask[ix]
			} else if action.HIDReport.Operation == "or" {
				report[ix] = report[ix] | mask[ix]
			}
		}

		err = sdl_device.HIDDevice.SendFeatureReport(byte(action.HIDReport.ReportID), report)
		if err != nil {
			logger.Logger.Error("[ProfileRunner::executeProfileListenerAction] could not send feature report", "id", action.HIDReport.ReportID, "error", err)
			return fmt.Errorf("could not send feature report: %w: %w", ErrHIDFailure, err)
		}
	}

	return fmt.Errorf("no valid action to execute: %w", ErrNoAction)
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
					go p.executeProfileListenerAction(controller.Device(), action)
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
				go p.executeProfileListenerAction(change_event.Device, action)
			}
		}
	}
}
