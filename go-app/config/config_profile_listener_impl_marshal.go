package config

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
)

func (c *Config_Controller_Profile_Listener_Action) UnmarshalJSON(data []byte) error {
	v := validator.New()

	var peek struct {
		Type string `type:"type"`
	}
	if err := json.Unmarshal(data, &peek); err != nil {
		return err
	}

	switch peek.Type {
	case "hid_output_report":
		var hr Config_Controller_Profile_Listener_Action_HIDOutputReport
		if err := json.Unmarshal(data, &hr); err != nil {
			return err
		}
		if err := v.Struct(hr); err != nil {
			return err
		}
		c.HIDOutputReport = &hr
		return nil
	case "hid_feature_report":
		var hr Config_Controller_Profile_Listener_Action_HIDFeatureReport
		if err := json.Unmarshal(data, &hr); err != nil {
			return err
		}
		if err := v.Struct(hr); err != nil {
			return err
		}
		c.HIDFeatureReport = &hr
		return nil
	}

	return fmt.Errorf("invalid action type (%s)", peek.Type)
}

func (c Config_Controller_Profile_Listener_Action) MarshalJSON() ([]byte, error) {
	if c.HIDOutputReport != nil {
		return json.Marshal(c.HIDOutputReport)
	}
	if c.HIDFeatureReport != nil {
		return json.Marshal(c.HIDFeatureReport)
	}
	return nil, fmt.Errorf("unable to marshal action; no valid action type found")
}
