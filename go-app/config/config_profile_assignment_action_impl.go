package config

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
)

func (c *Config_Controller_Profile_Control_Assignment_Action) UnmarshalJSON(data []byte) error {
	var peek struct {
		Type     *string  `json:"type,omitempty"`
		Controls *string  `json:"controls,omitempty"`
		ApiValue *float64 `json:"api_value,omitempty"`
	}
	if err := json.Unmarshal(data, &peek); err != nil {
		return err
	}

	v := validator.New()

	if peek.Type != nil && *peek.Type == "virtual" {
		var virtual_action Config_Controller_Profile_Control_Assignment_Action_Virtual
		if err := json.Unmarshal(data, &virtual_action); err != nil {
			return err
		}
		if err := v.Struct(virtual_action); err != nil {
			return err
		}
		c.Virtual = &virtual_action
		return nil
	}

	/* if api value is defined; try to unmarshall as API control action */
	if peek.ApiValue != nil {
		var ac_action Config_Controller_Profile_Control_Assignment_Action_ApiControl
		if err := json.Unmarshal(data, &ac_action); err != nil {
			return err
		}
		if err := v.Struct(ac_action); err != nil {
			return err
		}
		c.ApiControl = &ac_action
		return nil
	}

	/* if controls is defined; try to unmarshal it as a direct control action */
	if peek.Controls != nil {
		var dc_action Config_Controller_Profile_Control_Assignment_Action_DirectControl
		if err := json.Unmarshal(data, &dc_action); err != nil {
			return err
		}
		if err := v.Struct(dc_action); err != nil {
			return err
		}
		c.DirectControl = &dc_action
		return nil
	}

	/* default to a keys action */
	var keys_action Config_Controller_Profile_Control_Assignment_Action_Keys
	if err := json.Unmarshal(data, &keys_action); err != nil {
		return err
	}
	if err := v.Struct(keys_action); err != nil {
		return err
	}
	c.Keys = &keys_action
	return nil
}

func (c Config_Controller_Profile_Control_Assignment_Action) MarshalJSON() ([]byte, error) {
	if c.Virtual != nil {
		return json.Marshal(c.Virtual)
	}
	if c.DirectControl != nil {
		return json.Marshal(c.DirectControl)
	}
	if c.Keys != nil {
		return json.Marshal(c.Keys)
	}
	return nil, fmt.Errorf("unable to marshal control assignment action; has to be one of direct_control or keys but neither was found")
}

func (c *Config_Controller_Profile_Control_Assignment_Action) ToString() string {
	if c.Keys != nil {
		return c.Keys.Keys
	}
	if c.DirectControl != nil {
		return c.DirectControl.ToString()
	}
	return ""
}
