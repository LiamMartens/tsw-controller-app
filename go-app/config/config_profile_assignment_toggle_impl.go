package config

func (c *Config_Controller_Profile_Control_Assignment_Toggle) IsMatch(value float64) bool {
	if c.Match != nil {
		if *c.Match == "equals" {
			return value == c.Threshold
		}
	}
	return value >= c.Threshold
}
