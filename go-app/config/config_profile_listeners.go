package config

type Config_Controller_Profile_Listener_Action_Condition struct {
	Type     string `json:"type" validate:"oneof=cab_state_value api_value control_value"`
	Name     string `json:"name" validate:"required"`
	Operator string `json:"operator" validate:"required,oneof=eq gt lt gte lte"`
	Value    any    `json:"value" validate:"required"`
}

type Config_Controller_Profile_Listener_SharedAction struct {
	/* defaults to "all" */
	ConditionsEvaluationStrategy string                                                `json:"conditions_evaluation_strategy,omitempty" validate:"omitempty,oneof=all any"`
	Conditions                   []Config_Controller_Profile_Listener_Action_Condition `json:"conditions,omitempty"`
}

type Config_Controller_Profile_Listener_Action_HIDOutputReport struct {
	Config_Controller_Profile_Listener_SharedAction
	Type      string `json:"type" validate:"required,eq=hid_output_report"`
	ReportID  uint8  `json:"report_id" validate:"required"`
	Mask      []byte `json:"mask" validate:"required"`
	Operation string `json:"operation" validate:"required,oneof=and or"`
}

type Config_Controller_Profile_Listener_Action_HIDFeatureReport struct {
	Config_Controller_Profile_Listener_SharedAction
	Type      string `json:"type" validate:"required,eq=hid_feature_report"`
	ReportID  uint8  `json:"report_id" validate:"required"`
	Mask      []byte `json:"mask" validate:"required"`
	Operation string `json:"operation" validate:"required,oneof=and or"`
}

/**
 * Can match either HIDOutputReport or HIDFeatureReport depending on the internal type
 */
type Config_Controller_Profile_Listener_Action struct {
	HIDOutputReport  *Config_Controller_Profile_Listener_Action_HIDOutputReport  `json:"-"`
	HIDFeatureReport *Config_Controller_Profile_Listener_Action_HIDFeatureReport `json:"-"`
}

/**
 * Can match either API or Control struct depending on the internal type
 */
type Config_Controller_Profile_Listener struct {
	Actions []Config_Controller_Profile_Listener_Action `json:"actions" validate:"required"`
}
