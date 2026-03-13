package main

import (
	"sort"
	"tsw_controller_app/controller_mgr"
)

func (a *App) GetControllers() []Interop_GenericController {
	var controllers []Interop_GenericController = []Interop_GenericController{}
	a.sdl_controller_manager.ConfiguredControllers.ForEach(func(c controller_mgr.SDL_ControllerManager_ConfiguredController, _ controller_mgr.DeviceUniqueID) bool {
		controllers = append(controllers, Interop_GenericController{
			UniqueID:     c.Device().UniqueID(),
			DeviceID:     c.Device().DeviceID(),
			Name:         c.Name,
			IsConfigured: true,
			IsVirtual:    false,
		})
		return true
	})
	a.sdl_controller_manager.UnconfiguredControllers.ForEach(func(c controller_mgr.SDL_ControllerManager_UnconfiguredController, _ controller_mgr.DeviceUniqueID) bool {
		controllers = append(controllers, Interop_GenericController{
			UniqueID:     c.Joystick.UniqueID(),
			DeviceID:     c.Joystick.DeviceID(),
			Name:         c.Joystick.Name(),
			IsConfigured: false,
			IsVirtual:    false,
		})
		return true
	})
	a.virtual_controller_manager.Controllers().ForEach(func(c *controller_mgr.VirtualControllerManager_Controller, key controller_mgr.DeviceUniqueID) bool {
		controllers = append(controllers, Interop_GenericController{
			UniqueID:     c.Device().UniqueID(),
			DeviceID:     c.Device().DeviceID(),
			Name:         c.Device().Name(),
			IsConfigured: true,
			IsVirtual:    true,
		})
		return true
	})
	sort.Slice(controllers, func(i, j int) bool {
		return controllers[i].Name < controllers[j].Name
	})
	return controllers
}

func (a *App) GetControllerConfiguration(unique_id controller_mgr.DeviceUniqueID) *Interop_ControllerConfiguration {
	if controller, has_controller := a.sdl_controller_manager.ConfiguredControllers.Get(unique_id); has_controller {
		// /* when configured the SDL map and calibration always exist */
		sdl_mapping, _ := controller.Manager.Config().SDLMappingsByDeviceID.Get(controller.Joystick.DeviceID())
		interop_calibration := Interop_ControllerCalibration{
			Name:     sdl_mapping.Name,
			DeviceID: sdl_mapping.UsbID,
			Controls: []Interop_ControllerCalibration_Control{},
		}
		controller.Controls().ForEach(func(c controller_mgr.IControllerManager_Controller_Control, key string) bool {
			if control, ok := c.(*controller_mgr.SDL_ControllerManager_Controller_JoyControl); ok {
				sdl_mapping := control.SDLMapping()
				calibration_data := control.Calibration()
				calibration := Interop_ControllerCalibration_Control{
					Kind:        sdl_mapping.Kind,
					Index:       sdl_mapping.Index,
					Name:        control.Name(),
					Min:         calibration_data.Min,
					Max:         calibration_data.Max,
					Idle:        0,
					Deadzone:    0,
					Invert:      false,
					EasingCurve: []float64{0.0, 0.0, 1.0, 1.0},
				}
				if calibration_data.Idle != nil {
					calibration.Idle = *calibration_data.Idle
				}
				if calibration_data.Deadzone != nil {
					calibration.Deadzone = *calibration_data.Deadzone
				}
				if calibration_data.Invert != nil {
					calibration.Invert = *calibration_data.Invert
				}
				if calibration_data.EasingCurve != nil {
					calibration.EasingCurve = *calibration_data.EasingCurve
				}
				interop_calibration.Controls = append(interop_calibration.Controls, calibration)
			}
			return true
		})
		return &Interop_ControllerConfiguration{
			SDLMapping:  sdl_mapping,
			Calibration: interop_calibration,
		}
	}
	return nil
}
