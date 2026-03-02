package config

import (
	"fmt"
	"strings"
)

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
