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
	case "hid_report":
		var hr Config_Controller_Profile_Listener_Action_HIDReport
		if err := json.Unmarshal(data, &hr); err != nil {
			return err
		}
		if err := v.Struct(hr); err != nil {
			return err
		}
		c.HIDReport = &hr
		return nil
	}

	return fmt.Errorf("invalid action type (%s)", peek.Type)
}

func (c Config_Controller_Profile_Listener_Action) MarshalJSON() ([]byte, error) {
	if c.HIDReport != nil {
		return json.Marshal(c.HIDReport)
	}
	return nil, fmt.Errorf("unable to marshal action; no valid action type found")
}

func (c *Config_Controller_Profile_Listener) UnmarshalJSON(data []byte) error {
	v := validator.New()

	var peek struct {
		Type string `type:"type"`
	}
	if err := json.Unmarshal(data, &peek); err != nil {
		return err
	}

	switch peek.Type {
	case "api_value":
		var api Config_Controller_Profile_Listener_Type_APIValue
		if err := json.Unmarshal(data, &api); err != nil {
			return err
		}
		if err := v.Struct(api); err != nil {
			return err
		}
		c.API = &api
		return nil
	}
	return fmt.Errorf("invalid listener type (%s)", peek.Type)
}

func (c Config_Controller_Profile_Listener) MarshalJSON() ([]byte, error) {
	if c.API != nil {
		return json.Marshal(c.API)
	}
	return nil, fmt.Errorf("unable to marshal listener; no valid listener type found")
}
