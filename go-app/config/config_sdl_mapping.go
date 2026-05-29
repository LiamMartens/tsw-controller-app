package config

import (
	"encoding/json"
	"fmt"
	"tsw_controller_app/sdl_mgr"

	"github.com/go-playground/validator/v10"
)

type Config_Controller_SDLMap_Control struct {
	Kind  sdl_mgr.SDLMgr_Control_Kind `json:"kind" validate:"required" example:"button"`
	Index int                         `json:"index" validate:"required"`
	Name  string                      `json:"name" validate:"required" example:"Lever1"`
}

type Config_Controller_SDLMap struct {
	Name string `json:"name" example:"Thrustmaster Quadrant" validate:"required"`
	/* the USBID here is equivalent to the DeviceID - which for SDL devices is always the VID:PID combination */
	UsbID string `json:"usb_id" example:"{0xVENDOR_ID}:{0xPRODUCT_ID}" validate:"required"`
	/* the unique ID can be provided to override the default calibration map for a device based on their reported UniqueID */
	UniqueID string                             `json:"unique_id,omitempty"`
	Data     []Config_Controller_SDLMap_Control `json:"data" validate:"required"`
}

func ControllerSDLMapFromJSON(json_str string) (*Config_Controller_SDLMap, error) {
	var c Config_Controller_SDLMap
	if err := json.Unmarshal([]byte(json_str), &c); err != nil {
		return nil, err
	}

	v := validator.New()
	if err := v.Struct(c); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Config_Controller_SDLMap) GetUsbID() string {
	return c.UsbID
}

func (c *Config_Controller_SDLMap) GetUniqueID() string {
	return c.UniqueID
}

func (c *Config_Controller_SDLMap) Matches(config DeviceConfiguration) bool {
	unique_id := c.GetUniqueID()
	if unique_id != "" {
		return unique_id == config.GetUniqueID()
	}
	return c.GetUsbID() == config.GetUsbID()
}

func (c *Config_Controller_SDLMap) FindByKindAndIndex(kind sdl_mgr.SDLMgr_Control_Kind, index int) (Config_Controller_SDLMap_Control, error) {
	for _, control := range c.Data {
		if control.Kind == kind && control.Index == index {
			return control, nil
		}
	}

	return Config_Controller_SDLMap_Control{}, fmt.Errorf("could not find control")
}
