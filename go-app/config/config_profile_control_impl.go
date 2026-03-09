package config

func (c *Config_Controller_Profile_Control) GetAssignments() []Config_Controller_Profile_Control_Assignment {
	var assignments []Config_Controller_Profile_Control_Assignment
	if c.Assignment != nil {
		assignments = append(assignments, *c.Assignment)
	} else if c.Assignments != nil {
		/* copy by value clone */
		assignments = append(assignments, *c.Assignments...)
	}
	return assignments
}
