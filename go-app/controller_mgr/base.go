package controller_mgr

import (
	"context"
	"time"
	"tsw_controller_app/map_utils"
)

type DeviceUniqueID = string

type IControllerManager_Device interface {
	UniqueID() string
	DeviceID() string
	Name() string
}

type IControllerManager_Controller_Control interface {
	Manager() IControllerManager
	Controller() IControllerManager_Controller
	Device() IControllerManager_Device
	Name() string
	UpdateValue(value float64, is_reset bool)
	GetState() ControllerManager_Controller_ControlState
}

type IControllerManager_Controller interface {
	Device() IControllerManager_Device
	Controls() *map_utils.LockMap[string, IControllerManager_Controller_Control]
	VirtualControls() *map_utils.LockMap[string, IControllerManager_Controller_Control]
	RegisterVirtualControl(name string, initialvalue float64)
}
type IControllerManager interface {
	Attach(c context.Context) context.CancelFunc
	SubscribeRaw() (chan IControllerManager_RawEvent, func())
	SubscribeChangeEvent() (chan ControllerManager_Control_ChangeEvent, func())
	SubscribeDevicesUpdated() (chan ControllerManager_Control_DevicesUpdated, func())
}

type ControllerManager_Controller_ControlState_DirectionChangeMarker struct {
	/* the actual direction of travel; -1 | 0 | 1 depending on the direction */
	Direction int8
	/* the value at which the direction changed */
	ChangeValue float64
}

type ControllerManager_Controller_ControlStateValues struct {
	/* the current value of the control state */
	Value float64
	/* the previous value of the control state */
	PreviousValue float64
	/* the value at which the control was initialized */
	InitialValue float64
}

type ControllerManager_Controller_ControlState struct {
	LastUpdatedAt *time.Time
	Direction     ControllerManager_Controller_ControlState_DirectionChangeMarker
	/* the normalized value states are in 0-1 format */
	NormalizedValues ControllerManager_Controller_ControlStateValues
	/* the raw values are in their raw value format coming from sdl */
	RawValues ControllerManager_Controller_ControlStateValues
}
