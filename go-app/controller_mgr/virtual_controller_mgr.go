package controller_mgr

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
	"tsw_controller_app/logger"
	"tsw_controller_app/map_utils"
	"tsw_controller_app/pubsub_utils"
	"tsw_controller_app/tswconnector"
)

const VIRTUAL_CONTROLLER_DEVICE_CONNECTED_CONNECTOR_EVENT_NAME = "virtual_device_connected"
const VIRTUAL_CONTROLLER_DEVICE_DISCONNECTED_CONNECTOR_EVENT_NAME = "virtual_device_disconnected"
const VIRTUAL_CONTROLLER_AXIS_VALUE_CONNECTOR_EVENT_NAME = "virtual_device_axis_value"
const VIRTUAL_CONTROLLER_BUTTON_VALUE_CONNECTOR_EVENT_NAME = "virtual_device_button_value"
const VIRTUAL_CONTROLLER_HAT_VALUE_CONNECTOR_EVENT_NAME = "virtual_device_hat_value"

var InvalidVirtualDeviceIDsError = errors.New("invalid unique or device ID reported by virtual device")
var MissingVirtualDeviveError = errors.New("virtual device is not available")

type VirtualControllerManager_Device struct {
	uniqueID string
	deviceID string
	name     string
}

type VirtualControllerManager_Controller struct {
	manager         *VirtualControllerManager
	device          *VirtualControllerManager_Device
	controls        *map_utils.LockMap[string, IControllerManager_Controller_Control]
	virtualControls *map_utils.LockMap[string, IControllerManager_Controller_Control]
}

type VirtualControllerManager_Controller_Control struct {
	IControllerManager_Controller_Control
	manager    *VirtualControllerManager
	controller *VirtualControllerManager_Controller
	device     *VirtualControllerManager_Device
	name       string
	state      ControllerManager_Controller_ControlState
}

type VirtualControllerManager struct {
	context                context.Context
	connector              tswconnector.TSWConnector
	controllers            *map_utils.LockMap[DeviceUniqueID, *VirtualControllerManager_Controller]
	rawEventChannels       *pubsub_utils.PubSubSlice[IControllerManager_RawEvent]
	changeEventChannels    *pubsub_utils.PubSubSlice[ControllerManager_Control_ChangeEvent]
	devicesUpdatedChannels *pubsub_utils.PubSubSlice[ControllerManager_Control_DevicesUpdated]
}

var _ IControllerManager_Device = &VirtualControllerManager_Device{}

var _ IControllerManager_Controller_Control = &VirtualControllerManager_Controller_Control{}

var _ IControllerManager_Controller = &VirtualControllerManager_Controller{}

var _ IControllerManager = &VirtualControllerManager{}

func (d *VirtualControllerManager_Device) UniqueID() string {
	return d.uniqueID
}

func (d *VirtualControllerManager_Device) DeviceID() string {
	return d.deviceID
}

func (d *VirtualControllerManager_Device) Name() string {
	return d.name
}

func (c *VirtualControllerManager_Controller_Control) Manager() IControllerManager {
	return c.manager
}

func (c *VirtualControllerManager_Controller_Control) Controller() IControllerManager_Controller {
	return c.controller
}

func (c *VirtualControllerManager_Controller_Control) Device() IControllerManager_Device {
	return c.device
}

func (c *VirtualControllerManager_Controller_Control) Name() string {
	return c.name
}

func (ctrl *VirtualControllerManager_Controller_Control) UpdateValue(value float64, is_reset bool) {
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
		Device:       ctrl.device,
		Controller:   ctrl.controller,
		Control:      ctrl,
		ControlName:  ctrl.name,
		ControlState: ctrl.state,
	})
}

func (c *VirtualControllerManager_Controller_Control) GetState() ControllerManager_Controller_ControlState {
	return c.state
}

func (ctrl *VirtualControllerManager_Controller) Device() IControllerManager_Device {
	return ctrl.device
}

func (ctrl *VirtualControllerManager_Controller) Controls() *map_utils.LockMap[string, IControllerManager_Controller_Control] {
	return ctrl.controls
}

func (ctrl *VirtualControllerManager_Controller) VirtualControls() *map_utils.LockMap[string, IControllerManager_Controller_Control] {
	return ctrl.virtualControls
}

func (controller *VirtualControllerManager_Controller) RegisterVirtualControl(name string, initial_value float64) {
	controller.virtualControls.Set(name, &VirtualControllerManager_Controller_Control{
		manager:    controller.manager,
		controller: controller,
		device:     controller.device,
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

func (vm *VirtualControllerManager) registerDeviceFromConnectorEvent(msg tswconnector.TSWConnector_Message) error {
	unique_id := msg.Properties["unique_id"]
	device_id := msg.Properties["device_id"]
	device_name := msg.Properties["device_name"]
	if strings.HasPrefix(unique_id, "virtual:") && strings.HasPrefix(device_id, "virtual:") {
		logger.Logger.Debug("[VirtualControllerManager::registerDeviceFromConnectorEvent] registering device", "unique_id", unique_id, "device_id", device_id)
		if vm.controllers.Contains(unique_id) {
			/* already registered; silently ignore */
			return nil
		}
		vm.controllers.Set(unique_id, &VirtualControllerManager_Controller{
			device: &VirtualControllerManager_Device{
				uniqueID: unique_id,
				deviceID: device_id,
				name:     device_name,
			},
			manager:         vm,
			controls:        map_utils.NewLockMap[string, IControllerManager_Controller_Control](),
			virtualControls: map_utils.NewLockMap[string, IControllerManager_Controller_Control](),
		})
		vm.devicesUpdatedChannels.EmitTimeout(time.Second, ControllerManager_Control_DevicesUpdated{})
		return nil
	}
	logger.Logger.Error("attempted to register a device with invalid IDs", "unique_id", unique_id, "device_id", device_id, "device_name", device_name)
	return InvalidVirtualDeviceIDsError
}

func (vm *VirtualControllerManager) deregisterDeviceFromConnectorEvent(msg tswconnector.TSWConnector_Message) error {
	unique_id := msg.Properties["unique_id"]
	if strings.HasPrefix(unique_id, "virtual:") {
		logger.Logger.Debug("[VirtualControllerManager::deregisterDeviceFromConnectorEvent] deregistering device", "unique_id", unique_id)
		vm.controllers.Delete(unique_id)
		return nil
	}
	logger.Logger.Error("attempted to de-register a device with invalid ID", "unique_id", unique_id)
	return InvalidVirtualDeviceIDsError
}

func (vm *VirtualControllerManager) updateDeviceControlValueFromConnectorEvent(msg tswconnector.TSWConnector_Message) error {
	unique_id := msg.Properties["unique_id"]
	control_name := msg.Properties["control"]
	control_value, _ := strconv.ParseFloat(msg.Properties["value"], 64)
	device, has_device := vm.controllers.Get(unique_id)
	if !has_device {
		return MissingVirtualDeviveError
	}

	control, has_control := device.controls.Get(control_name)
	if !has_control {
		control = &VirtualControllerManager_Controller_Control{
			manager:    vm,
			controller: device,
			device:     device.device,
			name:       control_name,
			state:      ControllerManager_Controller_ControlState{},
		}
		device.controls.Set(control_name, control)
	}
	control.UpdateValue(control_value, false)
	return nil
}

func (vm *VirtualControllerManager) Attach(ctx context.Context) context.CancelFunc {
	childctx, cancel := context.WithCancel(ctx)
	go func() {
		ch, unsubscribe := vm.connector.Subscribe()
		defer unsubscribe()
		for {
			select {
			case <-childctx.Done():
				return
			case msg := <-ch:
				switch msg.EventName {
				case VIRTUAL_CONTROLLER_DEVICE_CONNECTED_CONNECTOR_EVENT_NAME:
					vm.registerDeviceFromConnectorEvent(msg)
				case VIRTUAL_CONTROLLER_DEVICE_DISCONNECTED_CONNECTOR_EVENT_NAME:
					vm.deregisterDeviceFromConnectorEvent(msg)
				case VIRTUAL_CONTROLLER_AXIS_VALUE_CONNECTOR_EVENT_NAME:
					if err := vm.registerDeviceFromConnectorEvent(msg); err == nil {
						vm.updateDeviceControlValueFromConnectorEvent(msg)
					}
				case VIRTUAL_CONTROLLER_BUTTON_VALUE_CONNECTOR_EVENT_NAME:
					if err := vm.registerDeviceFromConnectorEvent(msg); err == nil {
						vm.updateDeviceControlValueFromConnectorEvent(msg)
					}
				case VIRTUAL_CONTROLLER_HAT_VALUE_CONNECTOR_EVENT_NAME:
					if err := vm.registerDeviceFromConnectorEvent(msg); err == nil {
						vm.updateDeviceControlValueFromConnectorEvent(msg)
					}
				}
			}
		}
	}()
	return cancel
}

func (mgr *VirtualControllerManager) SubscribeRaw() (chan IControllerManager_RawEvent, func()) {
	return mgr.rawEventChannels.Subscribe()
}

func (mgr *VirtualControllerManager) SubscribeChangeEvent() (chan ControllerManager_Control_ChangeEvent, func()) {
	return mgr.changeEventChannels.Subscribe()
}

func (mgr *VirtualControllerManager) SubscribeDevicesUpdated() (chan ControllerManager_Control_DevicesUpdated, func()) {
	return mgr.devicesUpdatedChannels.Subscribe()
}

func (mgr *VirtualControllerManager) Controllers() *map_utils.LockMap[DeviceUniqueID, *VirtualControllerManager_Controller] {
	return mgr.controllers
}

func NewVirtualControllerManager(conn tswconnector.TSWConnector) *VirtualControllerManager {
	return &VirtualControllerManager{
		context:                context.Background(),
		connector:              conn,
		controllers:            map_utils.NewLockMap[DeviceUniqueID, *VirtualControllerManager_Controller](),
		rawEventChannels:       pubsub_utils.NewPubSubSlice[IControllerManager_RawEvent](),
		changeEventChannels:    pubsub_utils.NewPubSubSlice[ControllerManager_Control_ChangeEvent](),
		devicesUpdatedChannels: pubsub_utils.NewPubSubSlice[ControllerManager_Control_DevicesUpdated](),
	}
}
