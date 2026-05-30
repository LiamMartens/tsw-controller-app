package controller_mgr

import (
	"context"
	"time"
	"tsw_controller_app/config"
	"tsw_controller_app/logger"
	"tsw_controller_app/map_utils"
	"tsw_controller_app/math_utils"
	"tsw_controller_app/pubsub_utils"
	"tsw_controller_app/sdl_mgr"

	"github.com/veandco/go-sdl2/sdl"
)

const DEFAULT_CHANNEL_BUFFER_SIZE = 50
const DIRECTION_CHANGE_THRESHOLD = 0.05

type SDL_ControllerManager_Controller_VirtualControl struct {
	manager    *SDLControllerManager
	controller *SDL_ControllerManager_ConfiguredController
	device     *sdl_mgr.SDLMgr_Joystick
	name       string
	state      ControllerManager_Controller_ControlState
}

type SDL_ControllerManager_Controller_JoyControl struct {
	SDL_ControllerManager_Controller_VirtualControl
	kind        sdl_mgr.SDLMgr_Control_Kind
	index       int
	sdlMapping  config.Config_Controller_SDLMap_Control
	calibration config.Config_Controller_CalibrationData
}

var _ IControllerManager_Controller_Control = &SDL_ControllerManager_Controller_VirtualControl{}
var _ IControllerManager_Controller_Control = &SDL_ControllerManager_Controller_JoyControl{}

type SDL_ControllerManager_ConfiguredController struct {
	Name            string
	Manager         *SDLControllerManager
	Joystick        *sdl_mgr.SDLMgr_Joystick
	controls        *map_utils.LockMap[string, IControllerManager_Controller_Control]
	virtualControls *map_utils.LockMap[string, IControllerManager_Controller_Control]
	/* this just stores the raw SDL mapping and calibration used */
	SDLMapping  *config.Config_Controller_SDLMap
	Calibration *config.Config_Controller_Calibration
}

type SDL_ControllerManager_UnconfiguredController struct {
	Joystick *sdl_mgr.SDLMgr_Joystick
	/* may have  partial configuration */
	SDLMapping  *config.Config_Controller_SDLMap
	Calibration *config.Config_Controller_Calibration
}

type SDL_ControllerManager_Config struct {
	SDLMappingsByDeviceID  *map_utils.LockMap[string, config.Config_Controller_SDLMap]
	CalibrationsByDeviceID *map_utils.LockMap[string, config.Config_Controller_Calibration]
	SDLMappingsByUniqueID  *map_utils.LockMap[string, config.Config_Controller_SDLMap]
	CalibrationsByUniqueID *map_utils.LockMap[string, config.Config_Controller_Calibration]
}

type SDLControllerManager struct {
	SDL                     *sdl_mgr.SDLMgr
	context                 context.Context
	config                  SDL_ControllerManager_Config
	ConfiguredControllers   *map_utils.LockMap[DeviceUniqueID, SDL_ControllerManager_ConfiguredController]
	UnconfiguredControllers *map_utils.LockMap[DeviceUniqueID, SDL_ControllerManager_UnconfiguredController]

	rawEventChannels          *pubsub_utils.PubSubSlice[IControllerManager_RawEvent]
	changeEventChannels       *pubsub_utils.PubSubSlice[ControllerManager_Control_ChangeEvent]
	joyDevicesUpdatedChannels *pubsub_utils.PubSubSlice[ControllerManager_Control_DevicesUpdated]
}

var _ IControllerManager = &SDLControllerManager{}

func (state *ControllerManager_Controller_ControlState) updateDirection() {
	last_direction_change_value := state.Direction.ChangeValue
	value_diff := state.NormalizedValues.Value - last_direction_change_value
	if value_diff > DIRECTION_CHANGE_THRESHOLD {
		state.Direction = ControllerManager_Controller_ControlState_DirectionChangeMarker{
			Direction:   1,
			ChangeValue: state.NormalizedValues.Value,
		}
	} else if value_diff < DIRECTION_CHANGE_THRESHOLD {
		state.Direction = ControllerManager_Controller_ControlState_DirectionChangeMarker{
			Direction:   -1,
			ChangeValue: state.NormalizedValues.Value,
		}
	}
}

func (ctrl *SDL_ControllerManager_Controller_JoyControl) reset() {
	switch ctrl.sdlMapping.Kind {
	case sdl_mgr.SDLMgr_Control_Kind_Axis:
		axis_value := ctrl.device.InternalJoystick.Axis(ctrl.sdlMapping.Index)
		ctrl.UpdateValue(float64(axis_value), true)
	case sdl_mgr.SDLMgr_Control_Kind_Button:
		button_value := int(ctrl.device.InternalJoystick.Button(ctrl.sdlMapping.Index))
		ctrl.UpdateValue(float64(int(button_value)), true)
	case sdl_mgr.SDLMgr_Control_Kind_Hat:
		hat_value := int(ctrl.device.InternalJoystick.Hat(sdl.JoystickHat(ctrl.sdlMapping.Index)))
		ctrl.UpdateValue(float64(int(hat_value)), true)
	}
}

func (ctrl *SDL_ControllerManager_Controller_JoyControl) Manager() IControllerManager {
	return ctrl.manager
}

func (ctrl *SDL_ControllerManager_Controller_JoyControl) Controller() IControllerManager_Controller {
	return ctrl.controller
}

func (ctrl *SDL_ControllerManager_Controller_JoyControl) Device() IControllerManager_Device {
	return ctrl.device
}

func (ctrl *SDL_ControllerManager_Controller_JoyControl) GetState() ControllerManager_Controller_ControlState {
	return ctrl.state
}

func (ctrl *SDL_ControllerManager_Controller_JoyControl) UpdateValue(value float64, is_reset bool) {
	/* update raw values */
	if is_reset {
		ctrl.state.RawValues.PreviousValue = value
		ctrl.state.RawValues.InitialValue = value
	} else {
		ctrl.state.RawValues.PreviousValue = ctrl.state.RawValues.Value
	}
	ctrl.state.RawValues.Value = value

	/* update normal values */
	normalized_value := ctrl.calibration.NormalizeRawValue(value)
	/* only update if value is not within margin error or  if this is a reset value */
	if is_reset {
		ctrl.state.NormalizedValues.InitialValue = normalized_value.Value
		ctrl.state.NormalizedValues.PreviousValue = normalized_value.Value
		ctrl.state.NormalizedValues.Value = normalized_value.Value
	} else if !math_utils.IsWithinMarginOfError(normalized_value.Value, ctrl.state.NormalizedValues.Value) {
		ctrl.state.NormalizedValues.PreviousValue = ctrl.state.NormalizedValues.Value
		ctrl.state.NormalizedValues.Value = normalized_value.Value
	}

	/* update direction */
	if is_reset {
		ctrl.state.Direction = ControllerManager_Controller_ControlState_DirectionChangeMarker{
			Direction:   0,
			ChangeValue: ctrl.state.NormalizedValues.Value,
		}
	} else {
		ctrl.state.updateDirection()
	}

	ctrl.manager.changeEventChannels.EmitTimeout(time.Second, ControllerManager_Control_ChangeEvent{
		Device: &ControllerManager_ChangeEvent_Device{
			UniqueID: ctrl.device.UniqueID(),
			DeviceID: ctrl.device.DeviceID(),
		},
		Controller:   ctrl.controller,
		Control:      ctrl,
		ControlName:  ctrl.name,
		ControlState: ctrl.state,
	})
}

func (ctrl *SDL_ControllerManager_Controller_JoyControl) Kind() sdl_mgr.SDLMgr_Control_Kind {
	return ctrl.kind
}

func (ctrl *SDL_ControllerManager_Controller_JoyControl) Index() int {
	return ctrl.index
}

func (ctrl *SDL_ControllerManager_Controller_JoyControl) SDLMapping() config.Config_Controller_SDLMap_Control {
	return ctrl.sdlMapping
}

func (ctrl *SDL_ControllerManager_Controller_JoyControl) Calibration() config.Config_Controller_CalibrationData {
	return ctrl.calibration
}

func (ctrl *SDL_ControllerManager_Controller_JoyControl) ProcessEvent(event sdl.Event) {
	switch e := event.(type) {
	case *sdl.JoyAxisEvent:
		ctrl.UpdateValue(float64(e.Value), false)
	case *sdl.JoyButtonEvent:
		switch e.State {
		case sdl.PRESSED:
			ctrl.UpdateValue(1.0, false)
		case sdl.RELEASED:
			ctrl.UpdateValue(0.0, false)
		}
	case *sdl.JoyHatEvent:
		ctrl.UpdateValue(float64(e.Value), false)
	}
}

func (ctrl *SDL_ControllerManager_Controller_VirtualControl) Manager() IControllerManager {
	return ctrl.manager
}
func (ctrl *SDL_ControllerManager_Controller_VirtualControl) Device() IControllerManager_Device {
	return ctrl.device
}

func (ctrl *SDL_ControllerManager_Controller_VirtualControl) Controller() IControllerManager_Controller {
	return ctrl.controller
}

func (ctrl *SDL_ControllerManager_Controller_VirtualControl) Name() string {
	return ctrl.name
}

func (ctrl *SDL_ControllerManager_Controller_VirtualControl) GetState() ControllerManager_Controller_ControlState {
	return ctrl.state
}

func (ctrl *SDL_ControllerManager_Controller_VirtualControl) UpdateValue(value float64, is_reset bool) {
	if is_reset {
		ctrl.state.RawValues.PreviousValue = value
		ctrl.state.RawValues.InitialValue = value
		ctrl.state.NormalizedValues.InitialValue = value
		ctrl.state.NormalizedValues.PreviousValue = value
	} else {
		ctrl.state.RawValues.PreviousValue = ctrl.state.RawValues.Value
		ctrl.state.NormalizedValues.PreviousValue = ctrl.state.NormalizedValues.Value
	}
	ctrl.state.RawValues.Value = value
	ctrl.state.NormalizedValues.Value = value

	/* update direction */
	if is_reset {
		ctrl.state.Direction = ControllerManager_Controller_ControlState_DirectionChangeMarker{
			Direction:   0,
			ChangeValue: ctrl.state.NormalizedValues.Value,
		}
	} else {
		ctrl.state.updateDirection()
	}

	ctrl.manager.changeEventChannels.EmitTimeout(time.Second, ControllerManager_Control_ChangeEvent{
		Device: &ControllerManager_ChangeEvent_Device{
			UniqueID: ctrl.device.UniqueID(),
			DeviceID: ctrl.device.DeviceID(),
		},
		Controller:   ctrl.controller,
		Control:      ctrl,
		ControlName:  ctrl.name,
		ControlState: ctrl.state,
	})
}

func (controller *SDL_ControllerManager_ConfiguredController) Device() IControllerManager_Device {
	return controller.Joystick
}

func (controller *SDL_ControllerManager_ConfiguredController) Controls() *map_utils.LockMap[string, IControllerManager_Controller_Control] {
	return controller.controls
}

func (controller *SDL_ControllerManager_ConfiguredController) VirtualControls() *map_utils.LockMap[string, IControllerManager_Controller_Control] {
	return controller.virtualControls
}

func (controller *SDL_ControllerManager_ConfiguredController) RegisterVirtualControl(name string, initial_value float64) {
	controller.virtualControls.Set(name, &SDL_ControllerManager_Controller_VirtualControl{
		manager:    controller.Manager,
		controller: controller,
		device:     controller.Joystick,
		name:       name,
		state: ControllerManager_Controller_ControlState{
			Direction: ControllerManager_Controller_ControlState_DirectionChangeMarker{
				Direction:   0,
				ChangeValue: initial_value,
			},
			NormalizedValues: ControllerManager_Controller_ControlStateValues{
				Value:         initial_value,
				PreviousValue: initial_value,
				InitialValue:  initial_value,
			},
			RawValues: ControllerManager_Controller_ControlStateValues{
				Value:         initial_value,
				PreviousValue: initial_value,
				InitialValue:  initial_value,
			},
		},
	})
}

func (controller *SDL_ControllerManager_ConfiguredController) ProcessEvent(event sdl.Event) {
	switch e := event.(type) {
	case *sdl.JoyAxisEvent:
		controller.controls.ForEach(func(maybe_axis IControllerManager_Controller_Control, _ string) bool {
			if axis, ok := maybe_axis.(*SDL_ControllerManager_Controller_JoyControl); ok {
				if axis.sdlMapping.Kind == sdl_mgr.SDLMgr_Control_Kind_Axis && axis.sdlMapping.Index == int(e.Axis) {
					axis.ProcessEvent(event)
				}
			}

			return true
		})
	case *sdl.JoyButtonEvent:
		controller.controls.ForEach(func(maybe_axis IControllerManager_Controller_Control, _ string) bool {
			if axis, ok := maybe_axis.(*SDL_ControllerManager_Controller_JoyControl); ok {
				if axis.sdlMapping.Kind == sdl_mgr.SDLMgr_Control_Kind_Button && axis.sdlMapping.Index == int(e.Button) {
					axis.ProcessEvent(event)
				}
			}
			return true
		})
	case *sdl.JoyHatEvent:
		controller.controls.ForEach(func(maybe_axis IControllerManager_Controller_Control, _ string) bool {
			if axis, ok := maybe_axis.(*SDL_ControllerManager_Controller_JoyControl); ok {
				if axis.sdlMapping.Kind == sdl_mgr.SDLMgr_Control_Kind_Hat && axis.sdlMapping.Index == int(e.Hat) {
					axis.ProcessEvent(event)
				}
			}
			return true
		})
	}
}

func NewSDLControllerManager(sdlmgr *sdl_mgr.SDLMgr) *SDLControllerManager {
	return &SDLControllerManager{
		SDL: sdlmgr,
		config: SDL_ControllerManager_Config{
			SDLMappingsByDeviceID:  map_utils.NewLockMap[string, config.Config_Controller_SDLMap](),
			CalibrationsByDeviceID: map_utils.NewLockMap[string, config.Config_Controller_Calibration](),
			SDLMappingsByUniqueID:  map_utils.NewLockMap[string, config.Config_Controller_SDLMap](),
			CalibrationsByUniqueID: map_utils.NewLockMap[string, config.Config_Controller_Calibration](),
		},
		ConfiguredControllers:   map_utils.NewLockMap[DeviceUniqueID, SDL_ControllerManager_ConfiguredController](),
		UnconfiguredControllers: map_utils.NewLockMap[DeviceUniqueID, SDL_ControllerManager_UnconfiguredController](),

		rawEventChannels:          pubsub_utils.NewPubSubSlice[IControllerManager_RawEvent](),
		changeEventChannels:       pubsub_utils.NewPubSubSlice[ControllerManager_Control_ChangeEvent](),
		joyDevicesUpdatedChannels: pubsub_utils.NewPubSubSlice[ControllerManager_Control_DevicesUpdated](),
	}
}

func (mgr *SDLControllerManager) shouldApplyJoystickSDLMap(joystick *sdl_mgr.SDLMgr_Joystick, sdl_map *config.Config_Controller_SDLMap) bool {
	if sdl_map.UniqueID == "" {
		/* only check DeviceID() to UsbID if there is no unique ID in the SDL map */
		return joystick.DeviceID() == sdl_map.UsbID
	}

	/* otherwise we only check the UniqueID */
	return joystick.UniqueID() == sdl_map.UniqueID
}

func (mgr *SDLControllerManager) IsConfigured(deviceid string) bool {
	is_configured := false
	mgr.ConfiguredControllers.ForEach(func(value SDL_ControllerManager_ConfiguredController, key DeviceUniqueID) bool {
		if value.Joystick.DeviceID() == deviceid {
			is_configured = true
			return false
		}
		return true
	})
	return is_configured
}

func (mgr *SDLControllerManager) Config() SDL_ControllerManager_Config {
	return mgr.config
}

func (mgr *SDLControllerManager) ConfigureJoystick(joystick *sdl_mgr.SDLMgr_Joystick, sdl_map config.Config_Controller_SDLMap, calibration config.Config_Controller_Calibration) SDL_ControllerManager_ConfiguredController {
	controller := SDL_ControllerManager_ConfiguredController{
		Name:            sdl_map.Name,
		Manager:         mgr,
		Joystick:        joystick,
		controls:        map_utils.NewLockMap[string, IControllerManager_Controller_Control](),
		virtualControls: map_utils.NewLockMap[string, IControllerManager_Controller_Control](),
		SDLMapping:      &sdl_map,
		Calibration:     &calibration,
	}
	for _, control := range sdl_map.Data {
		var calibration_data config.Config_Controller_CalibrationData = config.Config_Controller_CalibrationData{
			Id:           control.Name,
			IsCalibrated: false,
		}
		for _, data := range calibration.Data {
			if data.Id == control.Name {
				calibration_data = data
				calibration_data.IsCalibrated = true
				break
			}
		}

		idle_value := 0.0
		if calibration_data.Idle != nil {
			idle_value = *calibration_data.Idle
		}

		current_raw_value := idle_value
		switch control.Kind {
		case sdl_mgr.SDLMgr_Control_Kind_Axis:
			current_raw_value = float64(joystick.InternalJoystick.Axis(control.Index))
		case sdl_mgr.SDLMgr_Control_Kind_Button:
			current_raw_value = float64(joystick.InternalJoystick.Button(control.Index))
		case sdl_mgr.SDLMgr_Control_Kind_Hat:
			current_raw_value = float64(joystick.InternalJoystick.Hat(sdl.JoystickHat(control.Index)))
		}
		current_normal_value := calibration_data.NormalizeRawValue(current_raw_value).Value

		control := SDL_ControllerManager_Controller_JoyControl{
			SDL_ControllerManager_Controller_VirtualControl: SDL_ControllerManager_Controller_VirtualControl{
				manager:    mgr,
				device:     joystick,
				controller: &controller,
				name:       control.Name,
				state: ControllerManager_Controller_ControlState{
					Direction: ControllerManager_Controller_ControlState_DirectionChangeMarker{
						Direction:   0,
						ChangeValue: current_raw_value,
					},
					NormalizedValues: ControllerManager_Controller_ControlStateValues{
						Value:         current_normal_value,
						PreviousValue: current_normal_value,
						InitialValue:  current_normal_value,
					},
					RawValues: ControllerManager_Controller_ControlStateValues{
						Value:         current_raw_value,
						PreviousValue: current_raw_value,
						InitialValue:  current_raw_value,
					},
				},
			},
			kind:        control.Kind,
			index:       control.Index,
			sdlMapping:  control,
			calibration: calibration_data,
		}
		control.reset()
		controller.controls.Set(control.name, &control)
	}

	return controller
}

func (mgr *SDLControllerManager) RegisterConfig(sdl_map config.Config_Controller_SDLMap, calibration config.Config_Controller_Calibration) {
	/*
	 * sdl_map and calibration are supposed to match the device
	 */

	if sdl_map.UniqueID == "" {
		mgr.config.SDLMappingsByDeviceID.Set(sdl_map.UsbID, sdl_map)
	} else {
		mgr.config.SDLMappingsByUniqueID.Set(sdl_map.UniqueID, sdl_map)
	}

	if calibration.UniqueID == "" {
		mgr.config.CalibrationsByDeviceID.Set(calibration.UsbID, calibration)
	} else {
		mgr.config.CalibrationsByUniqueID.Set(calibration.UniqueID, calibration)
	}

	didConfigureJoystick := false

	/* configure unconfigured controller */
	mgr.UnconfiguredControllers.Mutate(func(unconfigured SDL_ControllerManager_UnconfiguredController, unique_id DeviceUniqueID) map_utils.LockMapMutateAction[DeviceUniqueID, SDL_ControllerManager_UnconfiguredController] {
		if mgr.shouldApplyJoystickSDLMap(unconfigured.Joystick, &sdl_map) {
			configured_controller := mgr.ConfigureJoystick(unconfigured.Joystick, sdl_map, calibration)
			mgr.ConfiguredControllers.Set(unique_id, configured_controller)
			didConfigureJoystick = true
			return map_utils.LockMapMutateAction[DeviceUniqueID, SDL_ControllerManager_UnconfiguredController]{
				Action: map_utils.LockMapMutateActionType_Delete,
				Key:    unique_id,
			}
		}
		return map_utils.LockMapMutateAction[DeviceUniqueID, SDL_ControllerManager_UnconfiguredController]{
			Action: map_utils.LockMapMutateActionType_Noop,
		}
	})

	/* replace configured controller */
	mgr.ConfiguredControllers.Mutate(func(configured SDL_ControllerManager_ConfiguredController, unique_id DeviceUniqueID) map_utils.LockMapMutateAction[DeviceUniqueID, SDL_ControllerManager_ConfiguredController] {
		if mgr.shouldApplyJoystickSDLMap(configured.Joystick, &sdl_map) {
			configured_controller := mgr.ConfigureJoystick(configured.Joystick, sdl_map, calibration)
			didConfigureJoystick = true
			return map_utils.LockMapMutateAction[DeviceUniqueID, SDL_ControllerManager_ConfiguredController]{
				Action: map_utils.LockMapMutateActionType_Replace,
				Key:    unique_id,
				Value:  configured_controller,
			}
		}
		return map_utils.LockMapMutateAction[DeviceUniqueID, SDL_ControllerManager_ConfiguredController]{
			Action: map_utils.LockMapMutateActionType_Noop,
		}
	})

	if didConfigureJoystick {
		mgr.joyDevicesUpdatedChannels.EmitTimeout(time.Second, ControllerManager_Control_DevicesUpdated{})
	}
}

func (mgr *SDLControllerManager) Handler_JoyDeviceAdded(event *sdl.JoyDeviceAddedEvent) error {
	joystick, err := mgr.SDL.GetJoystickByInstanceID(event.Which)
	if err != nil {
		return err
	}

	unique_sdl_map, has_unique_sdl_map := mgr.config.SDLMappingsByUniqueID.Get(joystick.UniqueID())
	unique_calibration, has_unique_calibration := mgr.config.CalibrationsByUniqueID.Get(joystick.UniqueID())
	device_sdl_map, has_device_sdl_map := mgr.config.SDLMappingsByDeviceID.Get(joystick.DeviceID())
	device_calibration, has_device_calibration := mgr.config.CalibrationsByDeviceID.Get(joystick.DeviceID())
	if has_unique_sdl_map && has_unique_calibration {
		configured_controller := mgr.ConfigureJoystick(joystick, unique_sdl_map, unique_calibration)
		mgr.ConfiguredControllers.Set(joystick.UniqueID(), configured_controller)
		mgr.joyDevicesUpdatedChannels.EmitTimeout(time.Second, ControllerManager_Control_DevicesUpdated{})
	} else if has_device_sdl_map && has_device_calibration {
		configured_controller := mgr.ConfigureJoystick(joystick, device_sdl_map, device_calibration)
		mgr.ConfiguredControllers.Set(joystick.UniqueID(), configured_controller)
		mgr.joyDevicesUpdatedChannels.EmitTimeout(time.Second, ControllerManager_Control_DevicesUpdated{})
	} else {
		unconfigured_controller := SDL_ControllerManager_UnconfiguredController{
			Joystick:    joystick,
			SDLMapping:  nil,
			Calibration: nil,
		}

		if has_unique_sdl_map {
			unconfigured_controller.SDLMapping = &unique_sdl_map
		} else if has_device_sdl_map {
			unconfigured_controller.SDLMapping = &device_sdl_map
		}

		if has_unique_calibration {
			unconfigured_controller.Calibration = &unique_calibration
		} else if has_device_calibration {
			unconfigured_controller.Calibration = &device_calibration
		}

		mgr.UnconfiguredControllers.Set(joystick.UniqueID(), unconfigured_controller)
		mgr.joyDevicesUpdatedChannels.EmitTimeout(time.Second, ControllerManager_Control_DevicesUpdated{})
	}

	return nil
}

func (mgr *SDLControllerManager) Handler_JoyDeviceRemoved(event *sdl.JoyDeviceRemovedEvent) error {
	mgr.ConfiguredControllers.Mutate(func(configured_controller SDL_ControllerManager_ConfiguredController, unique_id DeviceUniqueID) map_utils.LockMapMutateAction[DeviceUniqueID, SDL_ControllerManager_ConfiguredController] {
		if configured_controller.Joystick.InstanceID == event.Which {
			logger.Logger.Debug("[ControllerManager:Handler_JoyDeviceRemoved] Removing joy device", "name", configured_controller.Joystick.Name)
			defer func() {
				mgr.joyDevicesUpdatedChannels.EmitTimeout(time.Second, ControllerManager_Control_DevicesUpdated{})
			}()
			return map_utils.LockMapMutateAction[DeviceUniqueID, SDL_ControllerManager_ConfiguredController]{
				Action: map_utils.LockMapMutateActionType_Delete,
				Key:    unique_id,
			}
		}
		return map_utils.LockMapMutateAction[DeviceUniqueID, SDL_ControllerManager_ConfiguredController]{
			Action: map_utils.LockMapMutateActionType_Noop,
		}
	})

	mgr.UnconfiguredControllers.Mutate(func(unconfigured_controller SDL_ControllerManager_UnconfiguredController, unique_id DeviceUniqueID) map_utils.LockMapMutateAction[DeviceUniqueID, SDL_ControllerManager_UnconfiguredController] {
		if unconfigured_controller.Joystick.InstanceID == event.Which {
			logger.Logger.Debug("[ControllerManager:Handler_JoyDeviceRemoved] Removing joy device", "name", unconfigured_controller.Joystick.Name)
			defer func() {
				mgr.joyDevicesUpdatedChannels.EmitTimeout(time.Second, ControllerManager_Control_DevicesUpdated{})
			}()
			return map_utils.LockMapMutateAction[DeviceUniqueID, SDL_ControllerManager_UnconfiguredController]{
				Action: map_utils.LockMapMutateActionType_Delete,
				Key:    unique_id,
			}
		}
		return map_utils.LockMapMutateAction[DeviceUniqueID, SDL_ControllerManager_UnconfiguredController]{
			Action: map_utils.LockMapMutateActionType_Noop,
		}
	})

	return nil
}

func (mgr *SDLControllerManager) Handler_JoyAxisEvent(event *sdl.JoyAxisEvent) error {
	joystick, err := mgr.SDL.GetJoystickByInstanceID(event.Which)
	if err != nil {
		logger.Logger.Error("[ControllerManager::Handler_JoyAxisEvent] could not get joystick", "error", err)
		return err
	}

	/* only send if the channel is being read */
	mgr.rawEventChannels.EmitTimeout(time.Second, &ControllerManager_RawEvent_Axis{
		device: &ControllerManager_RawEvent_Device{
			UniqueID: joystick.UniqueID(),
			DeviceID: joystick.DeviceID(),
		},
		timestamp: int(event.GetTimestamp()),
		axis:      int(event.Axis),
		value:     float64(event.Value),
	})

	/* send for processing if configured */
	configured, is_configured := mgr.ConfiguredControllers.Get(joystick.UniqueID())
	if is_configured {
		configured.ProcessEvent(event)
	}

	return nil
}

func (mgr *SDLControllerManager) Handler_JoyButtonEvent(event *sdl.JoyButtonEvent) error {
	joystick, err := mgr.SDL.GetJoystickByInstanceID(event.Which)
	if err != nil {
		logger.Logger.Error("[ControllerManager::Handler_JoyButtonEvent] could not get joystick", "error", err)
		return err
	}

	/* only send if the channel is being read */
	mgr.rawEventChannels.EmitTimeout(time.Second, &ControllerManager_RawEvent_Button{
		device: &ControllerManager_RawEvent_Device{
			UniqueID: joystick.UniqueID(),
			DeviceID: joystick.DeviceID(),
		},
		timestamp: int(event.GetTimestamp()),
		button:    int(event.Button),
		value:     float64(event.State),
	})

	/* send for processing if configured */
	configured, is_configured := mgr.ConfiguredControllers.Get(joystick.UniqueID())
	if is_configured {
		logger.Logger.Debug("[ControllerManager::Handler_JoyButtonEvent] processing button event", "event", event)
		configured.ProcessEvent(event)
	} else {
		logger.Logger.Debug("[ControllerManager::Handler_JoyButtonEvent] skipping processing because of unconfigured controller", "event", event)
	}

	return nil
}

func (mgr *SDLControllerManager) Handler_JoyHatEvent(event *sdl.JoyHatEvent) error {
	joystick, err := mgr.SDL.GetJoystickByInstanceID(event.Which)
	if err != nil {
		logger.Logger.Error("[ControllerManager::Handler_JoyHatEvent] could not get joystick", "error", err)
		return err
	}

	/* only send if the channel is being read */
	mgr.rawEventChannels.EmitTimeout(time.Second, &ControllerManager_RawEvent_Hat{
		device: &ControllerManager_RawEvent_Device{
			UniqueID: joystick.UniqueID(),
			DeviceID: joystick.DeviceID(),
		},
		timestamp: int(event.GetTimestamp()),
		hat:       int(event.Hat),
		value:     float64(event.Value),
	})

	/* send for processing if configured */
	configured, is_configured := mgr.ConfiguredControllers.Get(joystick.UniqueID())
	if is_configured {
		configured.ProcessEvent(event)
	}

	return nil
}

func (mgr *SDLControllerManager) Attach(ctx context.Context) context.CancelFunc {
	ctx_with_cancel, cancel := context.WithCancel(ctx)

	go func() {
		/* returns a cancel but will be cancelled by it's parent context */
		events_channel, _ := mgr.SDL.StartPolling(ctx_with_cancel)
		for {
			select {
			case event := <-events_channel:
				logger.Logger.Debug("[ControllerManager.Attach] Received SDL2 event", "event", event)
				switch e := event.(type) {
				case *sdl.JoyDeviceAddedEvent:
					mgr.Handler_JoyDeviceAdded(e)
				case *sdl.JoyDeviceRemovedEvent:
					mgr.Handler_JoyDeviceRemoved(e)
				case *sdl.JoyAxisEvent:
					mgr.Handler_JoyAxisEvent(e)
				case *sdl.JoyButtonEvent:
					mgr.Handler_JoyButtonEvent(e)
				case *sdl.JoyHatEvent:
					mgr.Handler_JoyHatEvent(e)
				case *sdl.QuitEvent:
					cancel()
				}
			case <-ctx_with_cancel.Done():
				return
			}
		}
	}()
	return cancel
}

func (mgr *SDLControllerManager) SubscribeRaw() (chan IControllerManager_RawEvent, func()) {
	return mgr.rawEventChannels.Subscribe()
}

func (mgr *SDLControllerManager) SubscribeChangeEvent() (chan ControllerManager_Control_ChangeEvent, func()) {
	return mgr.changeEventChannels.Subscribe()
}

func (mgr *SDLControllerManager) SubscribeDevicesUpdated() (chan ControllerManager_Control_DevicesUpdated, func()) {
	return mgr.joyDevicesUpdatedChannels.Subscribe()
}
