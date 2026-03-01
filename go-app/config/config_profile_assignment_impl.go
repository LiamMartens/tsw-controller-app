package config

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
)

func (c *Config_Controller_Profile_Control_Assignment) Conditions() *[]Config_Controller_Profile_Control_Assignment_Condition {
	if c.Momentary != nil {
		return c.Momentary.Conditions
	}
	if c.Linear != nil {
		return c.Linear.Conditions
	}
	if c.Toggle != nil {
		return c.Toggle.Conditions
	}
	if c.DirectControl != nil {
		return c.DirectControl.Conditions
	}
	if c.ApiControl != nil {
		return c.ApiControl.Conditions
	}
	if c.SyncControl != nil {
		return c.SyncControl.Conditions
	}
	return nil
}

func (c *Config_Controller_Profile_Control_Assignment) RailClassInformation() *[]Config_Controller_Profile_RailClassInformationEntry {
	if c.Momentary != nil {
		return c.Momentary.RailClassInformation
	}
	if c.Linear != nil {
		return c.Linear.RailClassInformation
	}
	if c.Toggle != nil {
		return c.Toggle.RailClassInformation
	}
	if c.DirectControl != nil {
		return c.DirectControl.RailClassInformation
	}
	if c.ApiControl != nil {
		return c.ApiControl.RailClassInformation
	}
	if c.SyncControl != nil {
		return c.SyncControl.RailClassInformation
	}
	return nil
}

func (c *Config_Controller_Profile_Control_Assignment) UnmarshalJSON(data []byte) error {
	v := validator.New()

	var peek struct {
		Type string `type:"type"`
	}
	if err := json.Unmarshal(data, &peek); err != nil {
		return err
	}
	if err := v.Struct(peek); err != nil {
		return err
	}

	switch peek.Type {
	case "momentary":
		var momentary Config_Controller_Profile_Control_Assignment_Momentary
		if err := json.Unmarshal(data, &momentary); err != nil {
			return err
		}
		if err := v.Struct(momentary); err != nil {
			return err
		}
		c.Momentary = &momentary
		return nil
	case "linear":
		var linear Config_Controller_Profile_Control_Assignment_Linear
		if err := json.Unmarshal(data, &linear); err != nil {
			return err
		}
		if err := v.Struct(linear); err != nil {
			return err
		}
		c.Linear = &linear
		return nil
	case "toggle":
		var toggle Config_Controller_Profile_Control_Assignment_Toggle
		if err := json.Unmarshal(data, &toggle); err != nil {
			return err
		}
		if err := v.Struct(toggle); err != nil {
			return err
		}
		c.Toggle = &toggle
		return nil
	case "api_control":
		var ac Config_Controller_Profile_Control_Assignment_ApiControl
		if err := json.Unmarshal(data, &ac); err != nil {
			return err
		}
		if err := v.Struct(ac); err != nil {
			return err
		}
		c.ApiControl = &ac
		return nil
	case "direct_control":
		var dc Config_Controller_Profile_Control_Assignment_DirectControl
		if err := json.Unmarshal(data, &dc); err != nil {
			return err
		}
		if err := v.Struct(dc); err != nil {
			return err
		}
		c.DirectControl = &dc
		return nil
	case "sync_control":
		var sc Config_Controller_Profile_Control_Assignment_SyncControl
		if err := json.Unmarshal(data, &sc); err != nil {
			return err
		}
		if err := v.Struct(sc); err != nil {
			return err
		}
		c.SyncControl = &sc
		return nil
	}
	return fmt.Errorf("invalid assignment type (%s)", peek.Type)
}

func (c Config_Controller_Profile_Control_Assignment) MarshalJSON() ([]byte, error) {
	if c.Momentary != nil {
		return json.Marshal(c.Momentary)
	}
	if c.Linear != nil {
		return json.Marshal(c.Linear)
	}
	if c.Toggle != nil {
		return json.Marshal(c.Toggle)
	}
	if c.DirectControl != nil {
		return json.Marshal(c.DirectControl)
	}
	if c.SyncControl != nil {
		return json.Marshal(c.SyncControl)
	}
	if c.ApiControl != nil {
		return json.Marshal(c.ApiControl)
	}
	return nil, fmt.Errorf("unable to marshal control assignment; no valid assignment found")
}
