package config

import (
	"tsw_controller_app/math_utils"
)

func (condition *Config_Controller_Profile_Listener_Action_Condition) Matches(value any) bool {
	if target_value, is_float := condition.Value.(float64); is_float {
		incoming_value, incoming_value_is_float := value.(float64)
		if !incoming_value_is_float {
			return false
		}
		if condition.Operator == "eq" {
			return math_utils.IsWithinMarginOfError(target_value, incoming_value)
		}
		if condition.Operator == "gt" {
			return incoming_value > target_value
		}
		if condition.Operator == "gte" {
			return incoming_value >= target_value
		}
		if condition.Operator == "lt" {
			return incoming_value < target_value
		}
		if condition.Operator == "lte" {
			return incoming_value <= target_value
		}
		return false
	}

	if target_value, is_string := condition.Value.(string); is_string {
		incoming_value, incoming_value_is_string := value.(string)
		if !incoming_value_is_string {
			return false
		}
		if condition.Operator == "eq" {
			return target_value == incoming_value
		}
		return false
	}

	if target_value, is_bool := condition.Value.(bool); is_bool {
		incoming_value, incoming_value_is_bool := value.(bool)
		if !incoming_value_is_bool {
			return false
		}
		if condition.Operator == "eq" {
			return target_value == incoming_value
		}
		return false
	}

	return false
}
