package config

import (
	"tsw_controller_app/math_utils"
)

type Config_Controller_Profile_Control_Assignment_Linear_Threshold_Resolved struct {
	Value    float64
	ValueEnd *float64
}

func (c *Config_Controller_Profile_Control_Assignment_Linear_Threshold) Resolve(thresholds map[string]float64) Config_Controller_Profile_Control_Assignment_Linear_Threshold_Resolved {
	var value_end *float64 = nil
	if c.ValueEnd != nil {
		value := c.ValueEnd.GetValue(thresholds)
		value_end = &value
	}

	return Config_Controller_Profile_Control_Assignment_Linear_Threshold_Resolved{
		Value:    c.Value.GetValue(thresholds),
		ValueEnd: value_end,
	}
}

func (c *Config_Controller_Profile_Control_Assignment_Linear_Threshold) IsExceedingThreshold(value float64, thresholds map[string]float64) bool {
	resolved := c.Resolve(thresholds)

	if resolved.Value < 0.0 {
		return value < resolved.Value
	}
	return value >= resolved.Value
}

func (c *Config_Controller_Profile_Control_Assignment_Linear) GenerateThresholds(namedthresholds map[string]float64) []Config_Controller_Profile_Control_Assignment_Linear_Threshold {
	var thresholds []Config_Controller_Profile_Control_Assignment_Linear_Threshold
	for _, threshold := range c.Thresholds {
		resolved := threshold.Resolve(namedthresholds)
		if resolved.ValueEnd == nil || threshold.ValueStep == nil {
			thresholds = append(thresholds, Config_Controller_Profile_Control_Assignment_Linear_Threshold{
				Value:            Config_Threshold_Value{Value: resolved.Value},
				ActionActivate:   threshold.ActionActivate,
				ActionDeactivate: threshold.ActionDeactivate,
			})
		} else {
			current_value := resolved.Value
			for current_value <= *resolved.ValueEnd {
				thresholds = append(thresholds, Config_Controller_Profile_Control_Assignment_Linear_Threshold{
					Value: Config_Threshold_Value{
						Value: current_value,
					},
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
