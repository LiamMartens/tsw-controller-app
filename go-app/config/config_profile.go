package config

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
	"tsw_controller_app/math_utils"

	"github.com/go-playground/validator/v10"
)

type PreferredControlMode = string

const (
	PreferredControlMode_DirectControl PreferredControlMode = "direct_control"
	PreferredControlMode_SyncControl   PreferredControlMode = "sync_control"
	PreferredControlMode_ApiControl    PreferredControlMode = "api_control"
)

type ControlStepDefinition_Threshold struct {
	Value     float64
	ValueEnd  float64
	Tolerance float64
}

type ControlStepDefinition struct {
	IsFreeRange bool
	Value       float64
	Threshold   ControlStepDefinition_Threshold
}

type FreeRangeZone struct {
	Start float64
	End   float64
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
	Threshold          float64  `json:"threshold,omitempty"`           /* The actual threshold of this corresponding step. Can be combined with threshold tolerance */
	ThresholdEnd       *float64 `json:"threshold_end,omitempty"`       /* Defines the end value of the corresponding step */
	ThresholdTolerance *float64 `json:"threshold_tolerance,omitempty"` /* Defines the tolerance of the threshold and threshold_end; defaults to 0 */
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
	Notify        *bool `json:"notify,omitempty"`
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

func (c *Config_Controller_Profile_Control_Assignment_Action_DirectControl) ToString() string {
	flags := []string{}
	if c.Hold != nil && *c.Hold {
		flags = append(flags, "hold")
	}
	if c.Relative != nil && *c.Relative {
		flags = append(flags, "relative")
	}
	if c.UseNormalized != nil && *c.UseNormalized {
		flags = append(flags, "normalized")
	}

	return fmt.Sprintf("%s,%f,%s", c.Controls, c.Value, strings.Join(flags, "|"))
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

func (c *Config_Controller_Profile_Control_Assignment_Momentary) IsMatch(value float64) bool {
	if c.Match != nil {
		if *c.Match == "equals" {
			return value == c.Threshold
		}
	}
	return value >= c.Threshold
}

func (c *Config_Controller_Profile_Control_Assignment_Toggle) IsMatch(value float64) bool {
	if c.Match != nil {
		if *c.Match == "equals" {
			return value == c.Threshold
		}
	}
	return value >= c.Threshold
}

func (c *Config_Controller_Profile_Control_Assignment_Linear_Threshold) IsExceedingThreshold(value float64) bool {
	if c.Value < 0.0 {
		return value < c.Value
	}
	return value >= c.Value
}

func (c *Config_Controller_Profile_Control_Assignment_Linear) GenerateThresholds() []Config_Controller_Profile_Control_Assignment_Linear_Threshold {
	var thresholds []Config_Controller_Profile_Control_Assignment_Linear_Threshold
	for _, threshold := range c.Thresholds {
		if threshold.ValueEnd == nil || threshold.ValueStep == nil {
			thresholds = append(thresholds, threshold)
		} else {
			current_value := threshold.Value
			for current_value <= *threshold.ValueEnd {
				thresholds = append(thresholds, Config_Controller_Profile_Control_Assignment_Linear_Threshold{
					Value: current_value,
					/* generated thresholds don't need these anymore */
					ValueEnd:         nil,
					ValueStep:        nil,
					ActionActivate:   threshold.ActionActivate,
					ActionDeactivate: threshold.ActionDeactivate,
				})
				current_value = math_utils.RoundToMarginOfError(current_value + *threshold.ValueStep)
			}
		}
	}
	return thresholds
}

/*
Normalizes the input value according to the neutral value
*/
func (c *Config_Controller_Profile_Control_Assignment_Linear) CalculateNeutralizedValue(value float64) float64 {
	if c.Neutral != nil && *c.Neutral > 0 {
		return (value - *c.Neutral) * (1.0 / *c.Neutral)
	}
	return value
}

func (c *Config_Controller_Profile_Control_Assignment_DirectLike_ControlRange) Clamp(value float64) float64 {
	// Positive side
	if value >= 0.0 {

		start := math.Max(c.Start, 0.0)
		end := math.Max(c.End, start)
		if end == start {
			return 1.0
		}

		return (value - start) / (end - start)
	}

	// Negative side (may be inverted)
	start := math.Min(c.Start, 0.0)
	end := math.Min(c.End, 0.0)

	// If inverted (e.g. -0.5 → -0.8), swap
	if start < end {
		start, end = end, start
	}

	if start == end {
		return 1.0
	}

	// Normalize by magnitude toward 1.0
	return (end - value) / (end - start)
}

/**
* Returns a normalized steps definition list which contains both free range zones and normal steps
* and additionally. handles the threshold definitions
 */
func (c *Config_Controller_Profile_Control_Assignment_DirectLike_InputValue) GetSteps() []ControlStepDefinition {
	if c.Steps == nil && c.Step == nil {
		return []ControlStepDefinition{}
	}

	/* normalize into a single deduplicated steps list */
	steps := []*float64{}
	if c.Steps != nil {
		for _, step := range *c.Steps {
			num_steps := len(steps)
			/* dedup */
			if num_steps == 0 || steps[num_steps-1] != step {
				steps = append(steps, step)
			}
		}
	} else if c.Step != nil {
		current_value := c.Min
		for {
			step_value := current_value
			steps = append(steps, &step_value)
			current_value = math.Min(math_utils.RoundToMarginOfError(current_value+*c.Step), c.Max)
			if current_value >= c.Max {
				steps = append(steps, &c.Max)
				break
			}
		}
	}

	step_thresholds := []Config_Controller_Profile_Control_Assignment_DirectLike_InputValue_StepThreshold{}
	if c.StepThresholds != nil {
		step_thresholds = *c.StepThresholds
	}

	defs := []ControlStepDefinition{}
	for ix, step := range steps {
		num_steps := len(steps)
		default_controlvalue_step_tolerance := math_utils.RoundToMarginOfError((c.Max - c.Min) / (float64(num_steps) - 1.0) / 2.0)
		step_threshold := ControlStepDefinition_Threshold{}
		if len(step_thresholds) > ix {
			step_threshold.Value = step_thresholds[ix].Threshold
			step_threshold.ValueEnd = step_thresholds[ix].Threshold
			if step_thresholds[ix].ThresholdEnd != nil {
				step_threshold.ValueEnd = *step_thresholds[ix].ThresholdEnd
			}
			if step_thresholds[ix].ThresholdTolerance != nil {
				step_threshold.Tolerance = *step_thresholds[ix].ThresholdTolerance
			}
		} else if step != nil {
			step_threshold.Value = *step
			step_threshold.ValueEnd = *step
			step_threshold.Tolerance = default_controlvalue_step_tolerance
		} else {
			step_threshold.Value = c.Min
			step_threshold.ValueEnd = c.Max
			step_threshold.Tolerance = 0.0 /* free range zones get no tolerance by default */
			/* the previous and next values should always be non-nil because the deduplicated list can never contain 2 nil values in sequence */
			if ix > 0 {
				step_threshold.Value = *steps[ix-1]
			}
			if ix < num_steps-1 {
				step_threshold.ValueEnd = *steps[ix+1]
			}
		}

		if step != nil {
			defs = append(defs, ControlStepDefinition{
				IsFreeRange: false,
				Value:       step_threshold.Value,
				Threshold:   step_threshold,
			})
		} else {
			defs = append(defs, ControlStepDefinition{
				IsFreeRange: true,
				Value:       step_threshold.Value,
				Threshold:   step_threshold,
			})
		}
	}

	return defs
}

/*
*
The incoming value here can only be [-1, 1]
This calculates the actual value which would be sent to the game
*/
func (c *Config_Controller_Profile_Control_Assignment_DirectLike_InputValue) CalculateOutputValue(value float64) *float64 {
	input_value := value

	if c.Invert != nil && *c.Invert {
		if value < 0.0 {
			input_value = -1.0 - value
		} else {
			input_value = 1.0 - value
		}
	}

	total_distance := math.Abs(c.Max - c.Min)
	normal := (input_value * total_distance) + c.Min
	steps := c.GetSteps()

	/* free range zones will get prioritized */
	for _, step := range steps {
		if !step.IsFreeRange {
			continue
		}

		threshold_start := step.Threshold.Value - step.Threshold.Tolerance
		threshold_end := step.Threshold.ValueEnd + step.Threshold.Tolerance
		is_within_threshold := normal >= threshold_start && normal <= threshold_end
		if is_within_threshold {
			value := math_utils.Clamp(normal, c.Min, c.Max)
			return &value
		}
	}

	for _, step := range steps {
		if step.IsFreeRange {
			continue
		}

		threshold_start := step.Threshold.Value - step.Threshold.Tolerance
		threshold_end := step.Threshold.ValueEnd + step.Threshold.Tolerance
		is_within_threshold := normal >= threshold_start && normal <= threshold_end
		if is_within_threshold {
			value := math_utils.Clamp(step.Value, c.Min, c.Max)
			return &value
		}
	}

	return nil
}

func (c *Config_Controller_Profile) Id() string {
	id_str := fmt.Sprintf("%s-%s", c.Metadata.Path, c.Name)
	hash := sha1.Sum([]byte(id_str))
	return fmt.Sprintf("%x", hash)
}

func (c *Config_Controller_Profile) FindControlByName(name string) *Config_Controller_Profile_Control {
	for _, control := range c.Controls {
		if control.Name == name {
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
