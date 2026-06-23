package main

import (
	"tsw_controller_app/config"
	"tsw_controller_app/sdl_mgr"
)

type Interop_GenericController struct {
	UniqueID      string
	DeviceID      string /* this is the USB ID for SDL controllers */
	Name          string
	IsConfigured  bool
	IsVirtual     bool
	HasThresholds bool
}

type Interop_Profile_Metadata struct {
	Path       string
	IsEmbedded bool
	UpdatedAt  string
	Warnings   []string
}

type Interop_Profile struct {
	Id         string
	Name       string
	DeviceID   string
	AutoSelect *bool
	Apps       *[]string
	Metadata   Interop_Profile_Metadata
}

type Interop_RawEvent struct {
	UniqueID  string
	DeviceID  string
	Kind      sdl_mgr.SDLMgr_Control_Kind
	Index     int
	Value     float64
	Timestamp int
}

type Interop_ChangeEvent struct {
	UniqueID    string
	DeviceID    string
	ControlName string
	Value       float64
}

type Interop_ControllerCalibration_Control struct {
	Kind        sdl_mgr.SDLMgr_Control_Kind
	Index       int
	Name        string
	Min         float64
	Max         float64
	Idle        float64
	Deadzone    float64
	AntiJitter  float64
	EasingCurve []float64
	Invert      bool
	Thresholds  []config.Config_Controller_Calibration_Threshold
}

type Interop_ControllerCalibration struct {
	Name     string
	DeviceID string
	UniqueID string
	Controls []Interop_ControllerCalibration_Control
}

type Interop_ControllerConfiguration struct {
	Name        string
	DeviceID    string
	UniqueID    string
	Calibration Interop_ControllerCalibration
	SDLMapping  config.Config_Controller_SDLMap
}

type Interop_Cab_ControlState_Control struct {
	Identifier             string
	PropertyName           string
	CurrentValue           float64
	CurrentNormalizedValue float64
}

type Interop_Cab_ControlState struct {
	Name     string
	Controls []Interop_Cab_ControlState_Control
}

type Interop_SharedProfile_Author struct {
	Name string
	Url  *string
}

type Interop_SharedProfile struct {
	Name                string
	DeviceID            string
	Url                 string
	AutoSelect          *bool
	Apps                *[]string
	ContainsCalibration *bool
	Author              *Interop_SharedProfile_Author
}

type Interop_SelectedProfileInfo struct {
	Id   string
	Name string
}

type Interop_SaveControlMapping struct {
	ProfileJSON  string
	ExistingPath string
}

type Interop_SaveControlMapping_Options struct {
	ReplaceExistingAssignments bool
}
