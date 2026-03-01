package config

import (
	"math"
	"tsw_controller_app/math_utils"
)

type StepList_Item struct {
	IsFreeRange   bool
	Value         *float64
	PreviousValue *float64
	NextValue     *float64
}

type ControlStepDefinition_Threshold struct {
	ValueStart float64
	ValueEnd   float64
	Tolerance  float64
}

type ControlStepDefinition struct {
	IsFreeRange bool
	ValueStart  float64
	ValueEnd    float64
	Threshold   ControlStepDefinition_Threshold
}

func (t *ControlStepDefinition_Threshold) Delta() float64 {
	return math.Abs(t.ValueEnd - t.ValueStart)
}

func (t *ControlStepDefinition) Delta() float64 {
	return math.Abs(t.ValueEnd - t.ValueStart)
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

/*
Gets the actual step list as derived from step/steps
*/
func (c *Config_Controller_Profile_Control_Assignment_DirectLike_InputValue) GetStepsList() []StepList_Item {
	if c.Steps == nil && c.Step == nil {
		return []StepList_Item{}
	}

	/* gather raw numerical step values based on steps or step; identical sequential values */
	stepvalues := []*float64{}
	if c.Steps != nil {
		for ix, step := range *c.Steps {
			num_values := len(stepvalues)
			if num_values == 0 || stepvalues[ix-1] != step {
				stepvalues = append(stepvalues, step)
			}
		}
	} else if c.Step != nil {
		/* auto-generate steps based on the step param */
		current_value := c.Min
		for {
			step_value := current_value
			stepvalues = append(stepvalues, &step_value)
			if current_value >= c.Max {
				break
			}
			current_value = math.Min(math_utils.RoundToMarginOfError(current_value+*c.Step), c.Max)
		}
	}

	num_values := len(stepvalues)
	if num_values == 0 {
		return []StepList_Item{}
	}

	steplist := []StepList_Item{}
	for ix, step := range stepvalues {
		var previous_value *float64 = nil
		var next_value *float64 = nil
		if ix > 0 && stepvalues[ix-1] != nil {
			value := *stepvalues[ix-1]
			previous_value = &value
		}
		if ix < num_values-1 && stepvalues[ix+1] != nil {
			value := *stepvalues[ix+1]
			next_value = &value
		}
		steplist = append(steplist, StepList_Item{
			IsFreeRange:   step == nil,
			Value:         step,
			PreviousValue: previous_value,
			NextValue:     next_value,
		})
	}
	return steplist
}

/**
* Returns a normalized steps definition list which contains both free range zones and normal steps
* and additionally. handles the threshold definitions
 */
func (c *Config_Controller_Profile_Control_Assignment_DirectLike_InputValue) GetSteps() []ControlStepDefinition {
	stepslist := c.GetStepsList()
	if len(stepslist) == 0 {
		return []ControlStepDefinition{}
	}

	step_thresholds := []Config_Controller_Profile_Control_Assignment_DirectLike_InputValue_StepThreshold{}
	if c.StepThresholds != nil {
		step_thresholds = *c.StepThresholds
	}

	defs := []ControlStepDefinition{}
	for ix, step := range stepslist {
		num_steps := len(stepslist)
		default_controlvalue_step_tolerance := math_utils.RoundToMarginOfError((c.Max - c.Min) / (float64(num_steps) - 1.0) / 2.0)
		step_threshold := ControlStepDefinition_Threshold{}
		if len(step_thresholds) > ix {
			step_threshold.ValueStart = step_thresholds[ix].Threshold
			step_threshold.ValueEnd = step_thresholds[ix].Threshold
			if step_thresholds[ix].ThresholdEnd != nil {
				step_threshold.ValueEnd = *step_thresholds[ix].ThresholdEnd
			}
			if step_thresholds[ix].ThresholdTolerance != nil {
				step_threshold.Tolerance = *step_thresholds[ix].ThresholdTolerance
			} else if !step.IsFreeRange {
				step_threshold.Tolerance = default_controlvalue_step_tolerance
			}
		} else if !step.IsFreeRange {
			step_threshold.ValueStart = *step.Value
			step_threshold.ValueEnd = *step.Value
			step_threshold.Tolerance = default_controlvalue_step_tolerance
		} else if step.IsFreeRange {
			step_threshold.ValueStart = c.Min
			step_threshold.ValueEnd = c.Max
			step_threshold.Tolerance = 0.0 /* free range zones get no tolerance by default */
			if step.PreviousValue != nil {
				step_threshold.ValueStart = *step.PreviousValue
			}
			if step.NextValue != nil {
				step_threshold.ValueEnd = *step.NextValue
			}
		}

		if !step.IsFreeRange {
			defs = append(defs, ControlStepDefinition{
				IsFreeRange: false,
				ValueStart:  *step.Value,
				ValueEnd:    *step.Value,
				Threshold:   step_threshold,
			})
		} else if step.IsFreeRange {
			value_start := c.Min
			value_end := c.Max
			if step.PreviousValue != nil {
				value_start = *step.PreviousValue
			}
			if step.NextValue != nil {
				value_end = *step.NextValue
			}
			defs = append(defs, ControlStepDefinition{
				IsFreeRange: true,
				ValueStart:  value_start,
				ValueEnd:    value_end,
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

	if len(steps) == 0 {
		/* if no steps are defined - send value directly */
		value := math_utils.Clamp(normal, c.Min, c.Max)
		return &value
	}

	/* free range zones will get prioritized */
	for _, step := range steps {
		if !step.IsFreeRange {
			continue
		}

		threshold_start := step.Threshold.ValueStart - step.Threshold.Tolerance
		threshold_end := step.Threshold.ValueEnd + step.Threshold.Tolerance
		is_within_threshold := normal >= threshold_start && normal <= threshold_end
		if is_within_threshold {
			/* normal depends on the threshold */
			incoming_value := step.ValueStart + step.Delta()*((normal-step.Threshold.ValueStart)/step.Threshold.Delta())
			value := math_utils.Clamp(incoming_value, c.Min, c.Max)
			return &value
		}
	}

	for _, step := range steps {
		if step.IsFreeRange {
			continue
		}

		threshold_start := step.Threshold.ValueStart - step.Threshold.Tolerance
		threshold_end := step.Threshold.ValueEnd + step.Threshold.Tolerance
		is_within_threshold := normal >= threshold_start && normal <= threshold_end
		if is_within_threshold {
			value := math_utils.Clamp(step.ValueStart, c.Min, c.Max)
			return &value
		}
	}

	return nil
}
