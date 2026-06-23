package config

import (
	"encoding/json"
	"math"
	"tsw_controller_app/math_utils"

	"github.com/creasty/go-easing"
	"github.com/go-playground/validator/v10"
)

type Config_Controller_Calibration_Threshold struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type Config_Controller_CalibrationData struct {
	/** the ID of the controller button or trigger as named in the controller mapping config (see other file - eg: "throttle1", "throttle2", "button1") */
	Id           string     `json:"id" validate:"required"`
	IsCalibrated bool       `json:"-"`
	Deadzone     *float64   `json:"deadzone,omitempty"`
	Invert       *bool      `json:"invert,omitempty"`
	Min          float64    `json:"min" validate:"required"`
	Max          float64    `json:"max" validate:"required"`
	Idle         *float64   `json:"idle,omitempty"`
	EasingCurve  *[]float64 `json:"easing_curve,omitempty"`
	AntiJitter   *float64   `json:"anti_jitter,omitempty"`
	/* these are named thresholds which can be used in step thresholds and easily remapped */
	Thresholds []Config_Controller_Calibration_Threshold `json:"thresholds,omitempty"`
}

type Config_Controller_Calibration struct {
	/* the USBID here is equivalent to the DeviceID - which for SDL devices is always the VID:PID combination */
	UsbID string `json:"usb_id" validate:"required" example:"{0xVENDOR_ID}:{0xPRODUCT_ID}"`
	/* the unique ID can be provided to override the default calibration map for a device based on their reported UniqueID */
	UniqueID string                              `json:"unique_id,omitempty"`
	Data     []Config_Controller_CalibrationData `json:"data" validate:"required"`
}

type NormalizedValue struct {
	Value            float64
	IsWithinDeadzone bool
}

func ControllerCalibrationFromJSON(json_str string) (*Config_Controller_Calibration, error) {
	var c Config_Controller_Calibration
	if err := json.Unmarshal([]byte(json_str), &c); err != nil {
		return nil, err
	}

	v := validator.New()
	if err := v.Struct(c); err != nil {
		return nil, err
	}

	return &c, nil
}

/*
Normalizes the raw input value to a [-1,1] range.
Will return IsWithinDeadzone true when within deadzone
*/
func (calibration *Config_Controller_CalibrationData) NormalizeRawValue(raw_value float64) NormalizedValue {
	if !calibration.IsCalibrated {
		return NormalizedValue{
			Value:            raw_value,
			IsWithinDeadzone: false,
		}
	}

	/* get optional values */
	idle_value := calibration.Min
	if calibration.Idle != nil {
		idle_value = *calibration.Idle
	}

	deadzone_value := float64(0)
	if calibration.Deadzone != nil {
		deadzone_value = *calibration.Deadzone
	}

	invert_value := false
	if calibration.Invert != nil {
		invert_value = *calibration.Invert
	}

	/* actual normalization logic starts here */
	idle_range := []float64{
		/* deadzone is optional so it will be initialized as 0 (no deadzone) */
		math.Max(idle_value-deadzone_value, calibration.Min),
		math.Min(idle_value+deadzone_value, calibration.Max),
	}

	easing_curve_value := []float64{0.0, 0.0, 1.0, 1.0}
	if calibration.EasingCurve != nil && len(*calibration.EasingCurve) == 4 {
		easing_curve_value = *calibration.EasingCurve
	}

	value := float64(raw_value)
	if invert_value {
		value = calibration.Max - value + calibration.Min
	}

	if value >= idle_range[0] && value <= idle_range[1] {
		return NormalizedValue{Value: 0, IsWithinDeadzone: true}
	}

	ease_func := easing.NewCustomEasing(easing_curve_value[0], easing_curve_value[1], easing_curve_value[2], easing_curve_value[3])

	/**
	Value will only be normalized to a negative value if the idle value is not the same as the minimum value
	AND the value is less than the lower idle range
	*/
	if calibration.Min != idle_value && value < idle_range[0] {
		abs_value := math.Abs(math_utils.Clamp((value-idle_range[0])/(calibration.Min-idle_range[0]), 0.0, 1.0))
		normal := NormalizedValue{Value: ease_func(abs_value) * -1.0, IsWithinDeadzone: false}
		return normal
	}

	abs_value := math.Abs(math_utils.Clamp((value-idle_range[1])/(calibration.Max-idle_range[1]), 0.0, 1.0))
	normal := NormalizedValue{Value: ease_func(abs_value), IsWithinDeadzone: false}
	return normal
}

func (c *Config_Controller_Calibration) GetUsbID() string {
	return c.UsbID
}

func (c *Config_Controller_Calibration) GetUniqueID() string {
	return c.UniqueID
}

func (c *Config_Controller_Calibration) Matches(config DeviceConfiguration) bool {
	unique_id := c.GetUniqueID()
	if unique_id != "" {
		return unique_id == config.GetUniqueID()
	}
	return c.GetUsbID() == config.GetUsbID()
}
