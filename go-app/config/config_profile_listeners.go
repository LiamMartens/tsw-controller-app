package config

type Config_Controller_Profile_Listener_Action_Condition struct {
	Operator string `json:"operator" validate:"required,oneof=eq gt lt gte lte"`
	Value    any    `json:"value" validate:"required"`
}

type Config_Controller_Profile_Listener_SharedAction struct {
	Conditions []Config_Controller_Profile_Listener_Action_Condition `json:"conditions,omitempty"`
}

type Config_Controller_Profile_Listener_Action_HIDOutputReport struct {
	Config_Controller_Profile_Listener_SharedAction
	Type      string `json:"type" validate:"required,eq=hid_output_report"`
	ReportID  uint8  `json:"report_id" validate:"required"`
	Mask      uint64 `json:"mask" validate:"required"`
	Operation string `json:"operation" validate:"required,oneof=and or"`
}

type Config_Controller_Profile_Listener_Action struct {
	HIDReport *Config_Controller_Profile_Listener_Action_HIDOutputReport `json:"-"`
}

type Config_Controller_Profile_Listener_Type_APIValue struct {
	Type      string                                      `json:"type" validate:"required,eq=api_value"`
	Path      string                                      `json:"path" validate:"required"`
	ValuesKey string                                      `json:"values_key" validate:"required"`
	Actions   []Config_Controller_Profile_Listener_Action `json:"actions" validate:"required"`
}

type Config_Controller_Profile_Listener_Type_ControlValue struct {
	Type    string                                      `json:"type" validate:"required,eq=control_value"`
	Name    string                                      `json:"name" validate:"required"`
	Actions []Config_Controller_Profile_Listener_Action `json:"actions" validate:"required"`
}

type Config_Controller_Profile_Listener struct {
	API     *Config_Controller_Profile_Listener_Type_APIValue     `json:"-"`
	Control *Config_Controller_Profile_Listener_Type_ControlValue `json:"-"`
}
