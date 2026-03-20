package config

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

type PreferredControlMode = string

const (
	PreferredControlMode_DirectControl PreferredControlMode = "direct_control"
	PreferredControlMode_SyncControl   PreferredControlMode = "sync_control"
	PreferredControlMode_ApiControl    PreferredControlMode = "api_control"
)

type Config_Threshold_Value struct {
	Value     float64
	Reference *string
}

type Config_Controller_Profile_Control_Assignment_Action_Keys struct {
	Keys      string   `json:"keys" validate:"required" example:"ctrl+a"`
	PressTime *float64 `json:"press_time,omitempty"`
	WaitTime  *float64 `json:"wait_time,omitempty"`
}

type Config_Controller_Profile_Control_Assignment_Action_Virtual struct {
	Type    string  `json:"type" validate:"oneof=virtual"`
	Control string  `json:"control" validate:"required,startswith=virtual:"`
	Value   float64 `json:"value"`
}

type Config_Controller_Profile_Control_Assignment_Action_DirectControl struct {
	Controls      string   `json:"controls" validate:"required"`
	Value         float64  `json:"value"`
	MaxChangeRate *float64 `json:"max_change_rate,omitempty"`
	/* sets this value to be a relative adjustment as opposed to an absolute one */
	Relative *bool `json:"relative,omitempty"`
	/* determine whether to hold the value or not; meaning the value will be sent continuously */
	Hold *bool `json:"hold,omitempty"`
	/* whether to apply raw or normalized values */
	UseNormalized *bool `json:"use_normalized,omitempty"`
	/* whether to notify the game for value changes (defaults to true) */
	Notify *bool `json:"notify,omitempty"`
	/* whether to enable fallback to the TSW API if available */
	EnableAPIFallback *bool `json:"enable_api_fallback,omitempty"`
}

type Config_Controller_Profile_Control_Assignment_Action_ApiControl struct {
	Controls      string   `json:"controls" validate:"required"`
	ApiValue      float64  `json:"api_value"`
	Hold          *bool    `json:"hold,omitempty"`
	MaxChangeRate *float64 `json:"max_change_rate,omitempty"`
}

type Config_Controller_Profile_Control_Assignment_Action struct {
	Keys          *Config_Controller_Profile_Control_Assignment_Action_Keys          `json:"-"`
	Virtual       *Config_Controller_Profile_Control_Assignment_Action_Virtual       `json:"-"`
	DirectControl *Config_Controller_Profile_Control_Assignment_Action_DirectControl `json:"-"`
	ApiControl    *Config_Controller_Profile_Control_Assignment_Action_ApiControl    `json:"-"`
}

type Config_Controller_Profile_Control_Assignment_Condition struct {
	/* this is the other control name to depend on */
	Control  string  `json:"control" validate:"required"`
	Operator string  `json:"operator" validate:"required,oneof=gte lte gt lt eq"`
	Value    float64 `json:"value"`
}

type Config_Controller_Profile_Control_Assignment_Shared struct {
	RailClassInformation *[]Config_Controller_Profile_RailClassInformationEntry    `json:"rail_class_information,omitempty"`
	Conditions           *[]Config_Controller_Profile_Control_Assignment_Condition `json:"conditions,omitempty"`
}

type Config_Controller_Profile_Control_Assignment_Momentary struct {
	Config_Controller_Profile_Control_Assignment_Shared
	Type      string  `json:"type" validate:"required,eq=momentary"`
	Threshold float64 `json:"threshold"`
	Match     *string `json:"match,omitempty" validate:"omitempty,oneof=exceeds equals"`
	/* which action to perform once the threshold is exceeded */
	ActionActivate Config_Controller_Profile_Control_Assignment_Action `json:"action_activate" validate:"required"`
	/* which action to perform once the threshold is not exceeded anymore; defaults to releasing the activate action if keys */
	ActionDeactivate *Config_Controller_Profile_Control_Assignment_Action `json:"action_deactivate,omitempty"`
}

type Config_Controller_Profile_Control_Assignment_Linear_Threshold struct {
	Value float64 `json:"value"`
	/* ValueEnd and ValueStep can be used to automatically generate a set of thresholds while keeping the same action (ie: throttle) */
	ValueEnd  *float64 `json:"value_end,omitempty"`
	ValueStep *float64 `json:"value_step,omitempty"`
	/* which action to perform once the linear threshold is exceeded */
	ActionActivate   Config_Controller_Profile_Control_Assignment_Action  `json:"action_activate" validate:"required"`
	ActionDeactivate *Config_Controller_Profile_Control_Assignment_Action `json:"action_deactivate,omitempty"`
}

type Config_Controller_Profile_Control_Assignment_Linear struct {
	Config_Controller_Profile_Control_Assignment_Shared
	Type       string                                                          `json:"type" validate:"required,eq=linear"`
	Neutral    *float64                                                        `json:"neutral,omitempty"`
	Thresholds []Config_Controller_Profile_Control_Assignment_Linear_Threshold `json:"thresholds" validate:"required"`
}

type Config_Controller_Profile_Control_Assignment_Toggle struct {
	Config_Controller_Profile_Control_Assignment_Shared
	Type      string  `json:"type" validate:"required,eq=toggle"`
	Threshold float64 `json:"threshold"`
	Match     *string `json:"match,omitempty" validate:"omitempty,oneof=exceeds equals"`
	/* which action to perform once the threshold is exceeded */
	ActionActivate   Config_Controller_Profile_Control_Assignment_Action `json:"action_activate" validate:"required"`
	ActionDeactivate Config_Controller_Profile_Control_Assignment_Action `json:"action_deactivate" validate:"required"`
}

type Config_Controller_Profile_Control_Assignment_DirectLike_ControlRange struct {
	Start float64 `json:"start"` /* depending on the value direction this is the min or maximum value */
	End   float64 `json:"end"`
}

type Config_Controller_Profile_Control_Assignment_DirectLike_InputValue_StepThreshold struct {
	Threshold          Config_Threshold_Value  `json:"threshold"`                     /* The actual threshold of this corresponding step. Can be combined with threshold tolerance */
	ThresholdEnd       *Config_Threshold_Value `json:"threshold_end,omitempty"`       /* Defines the end value of the corresponding step */
	ThresholdTolerance *float64                `json:"threshold_tolerance,omitempty"` /* Defines the tolerance of the threshold and threshold_end; defaults to 0 for free range zones and the default tolerance for normal steps */
}

type Config_Controller_Profile_Control_Assignment_DirectLike_InputValue struct {
	Min           float64  `json:"min"`
	Max           float64  `json:"max"`
	MaxChangeRate *float64 `json:"max_change_rate,omitempty"`
	Step          *float64 `json:"step,omitempty"`
	/** steps can be combined with null values to create automatic interpolation */
	Steps *[]*float64 `json:"steps,omitempty"`
	/* if defined should have the same number of elements as the steps */
	StepThresholds *[]Config_Controller_Profile_Control_Assignment_DirectLike_InputValue_StepThreshold `json:"step_thresholds,omitempty"`
	Invert         *bool                                                                               `json:"invert,omitempty"`
}

type Config_Controller_Profile_Control_Assignment_DirectControl struct {
	Config_Controller_Profile_Control_Assignment_Shared
	Type string `json:"type" validate:"required,eq=direct_control"`
	/* the HID control component as per the UE4SS API */
	Controls     string                                                                `json:"controls" validate:"required"`
	InputValue   Config_Controller_Profile_Control_Assignment_DirectLike_InputValue    `json:"input_value" validate:"required"`
	ControlRange *Config_Controller_Profile_Control_Assignment_DirectLike_ControlRange `json:"control_range,omitempty"`
	/* will hold the control in changing */
	Hold *bool `json:"hold,omitempty"`
	/* whether to apply raw or normalized values */
	UseNormalized *bool `json:"use_normalized,omitempty"`
	/* whether to enable fallback to the TSW API if available */
	EnableAPIFallback *bool `json:"enable_api_fallback,omitempty"`
	/* whether to send the notify flag */
	Notify *bool `json:"notify,omitempty"`
}

type Config_Controller_Profile_Control_Assignment_ApiControl struct {
	Config_Controller_Profile_Control_Assignment_Shared
	Type string `json:"type" validate:"required,eq=api_control"`
	/* the HID control component as per the UE4SS API / HTTP API - they are the same */
	Controls     string                                                                `json:"controls" validate:"required"`
	Hold         *bool                                                                 `json:"hold,omitempty"`
	InputValue   Config_Controller_Profile_Control_Assignment_DirectLike_InputValue    `json:"input_value" validate:"required"`
	ControlRange *Config_Controller_Profile_Control_Assignment_DirectLike_ControlRange `json:"control_range,omitempty"`
}

type Config_Controller_Profile_Control_Assignment_SyncControl struct {
	Config_Controller_Profile_Control_Assignment_Shared
	Type string `json:"type" validate:"required,eq=sync_control"`
	/** this is the VHID Identifier Name - differs from the direct control name */
	Identifier     string                                                                `json:"identifier" validate:"required"`
	InputValue     Config_Controller_Profile_Control_Assignment_DirectLike_InputValue    `json:"input_value" validate:"required"`
	ControlRange   *Config_Controller_Profile_Control_Assignment_DirectLike_ControlRange `json:"control_range,omitempty"`
	ActionIncrease Config_Controller_Profile_Control_Assignment_Action_Keys              `json:"action_increase" validate:"required"`
	ActionDecrease Config_Controller_Profile_Control_Assignment_Action_Keys              `json:"action_decrease" validate:"required"`
}

type Config_Controller_Profile_Control_Assignment struct {
	Momentary     *Config_Controller_Profile_Control_Assignment_Momentary     `json:"-"`
	Linear        *Config_Controller_Profile_Control_Assignment_Linear        `json:"-"`
	Toggle        *Config_Controller_Profile_Control_Assignment_Toggle        `json:"-"`
	DirectControl *Config_Controller_Profile_Control_Assignment_DirectControl `json:"-"`
	SyncControl   *Config_Controller_Profile_Control_Assignment_SyncControl   `json:"-"`
	ApiControl    *Config_Controller_Profile_Control_Assignment_ApiControl    `json:"-"`
}

type Config_Controller_Profile_Control struct {
	Name        string                                          `json:"name"`
	Assignment  *Config_Controller_Profile_Control_Assignment   `json:"assignment,omitempty"`
	Assignments *[]Config_Controller_Profile_Control_Assignment `json:"assignments,omitempty"`
}

type Config_Controller_Profile_Controller struct {
	/* if defined ; specifies this profile can only be used with the below SDL controller (by USB ID) */
	UsbID *string `json:"usb_id,omitempty"`
	/* Can be defined to specify a specific SDL mapping for this controller and profile; useful for sharing */
	Mapping     *Config_Controller_SDLMap      `json:"mapping,omitempty"`
	Calibration *Config_Controller_Calibration `json:"calibration,omitempty"`
}

type Config_Controller_Profile_Metadata struct {
	Path       string    `json:"-"`
	IsEmbedded bool      `json:"-"`
	UpdatedAt  time.Time `json:"-"`
	Warnings   []string  `json:"-"`
}

type Config_Controller_Profile_RailClassInformationEntry struct {
	ClassName string `json:"class_name"`
}

type Config_Controller_Profile struct {
	Metadata Config_Controller_Profile_Metadata  `json:"-"`
	Extends  *string                             `json:"extends,omitempty"`
	Name     string                              `json:"name" validate:"required"`
	Controls []Config_Controller_Profile_Control `json:"controls" validate:"required"`
	/* specifies if this profile can be autoselected */
	AutoSelect *bool `json:"auto_select,omitempty"`
	/* specifies which controller this profile is for */
	Controller *Config_Controller_Profile_Controller `json:"controller,omitempty"`
	/** specifies which rail classes this profile can be used on */
	RailClassInformation *[]Config_Controller_Profile_RailClassInformationEntry `json:"rail_class_information,omitempty"`
	/** specifies which apps this profile support(s) */
	Apps *[]string `json:"apps,omitempty"`
}

func (c *Config_Controller_Profile) Id() string {
	id_str := fmt.Sprintf("%s-%s", c.Metadata.Path, c.Name)
	hash := sha1.Sum([]byte(id_str))
	return fmt.Sprintf("%x", hash)
}

/*
Finds a control in the profile by it's name or by it's device ID and name
eg: "{name}" or "{device_id}:{name}"
*/
func (c *Config_Controller_Profile) FindControlByName(device_id string, name string) *Config_Controller_Profile_Control {
	with_device_id := fmt.Sprintf("%s:%s", device_id, name)
	for _, control := range c.Controls {
		if control.Name == with_device_id || control.Name == name {
			return &control
		}
	}
	return nil
}

func ControllerProfileFromJSON(json_str string, metadata Config_Controller_Profile_Metadata) (*Config_Controller_Profile, error) {
	var c Config_Controller_Profile = Config_Controller_Profile{
		Metadata: metadata,
	}
	if err := json.Unmarshal([]byte(json_str), &c); err != nil {
		return nil, err
	}

	v := validator.New()
	if err := v.Struct(c); err != nil {
		return nil, err
	}

	return &c, nil
}
