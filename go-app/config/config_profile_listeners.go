package config

type Config_Controller_Profile_Listener_SharedAction struct {
	Operator string  `json:"operator" validate:"required,oneof=eq gt lt gte lte"`
	Value    float64 `json:"value" validate:"required"`
}

type Config_Controller_Profile_Listener_Action_HIDReport struct {
	Config_Controller_Profile_Listener_SharedAction
	Type     string `json:"type" validate:"required,eq=hid_report"`
	ReportID uint8  `json:"report_id" validate:"required"`
	Mask     uint8  `json:"mask" validate:"required"`
}

type Config_Controller_Profile_Listener_Action struct {
	HIDReport *Config_Controller_Profile_Listener_Action_HIDReport `json:"-"`
}

type Config_Controller_Profile_Listener_Type_APIValue struct {
	Type      string                                      `json:"type" validate:"required,eq=api_value"`
	Path      string                                      `json:"path" validate:"required"`
	ValuesKey string                                      `json:"values_key,omitempty"`
	Actions   []Config_Controller_Profile_Listener_Action `json:"actions" validate:"required"`
}

type Config_Controller_Profile_Listener struct {
	API *Config_Controller_Profile_Listener_Type_APIValue `json:"-"`
}
