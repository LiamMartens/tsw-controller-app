package config

import "tsw_controller_app/math_utils"

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
