package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatches_Float64_Equal(t *testing.T) {
	cond := Config_Controller_Profile_Listener_Action_Condition{
		Operator: "eq",
		Value:    5.0,
	}

	assert.True(t, cond.Matches(5.0))
	assert.True(t, cond.Matches(5.00005)) // within margin of error
	assert.True(t, cond.Matches(4.99995)) // within margin of error
	assert.False(t, cond.Matches(5.0002))  // outside margin of error
	assert.False(t, cond.Matches(4.9998))  // outside margin of error
}

func TestMatches_Float64_GreaterThan(t *testing.T) {
	cond := Config_Controller_Profile_Listener_Action_Condition{
		Operator: "gt",
		Value:    5.0,
	}

	assert.True(t, cond.Matches(6.0))
	assert.True(t, cond.Matches(5.0001))
	assert.False(t, cond.Matches(5.0))
	assert.False(t, cond.Matches(4.0))
}

func TestMatches_Float64_GreaterThanOrEqual(t *testing.T) {
	cond := Config_Controller_Profile_Listener_Action_Condition{
		Operator: "gte",
		Value:    5.0,
	}

	assert.True(t, cond.Matches(6.0))
	assert.True(t, cond.Matches(5.0))
	assert.False(t, cond.Matches(4.0))
}

func TestMatches_Float64_LessThan(t *testing.T) {
	cond := Config_Controller_Profile_Listener_Action_Condition{
		Operator: "lt",
		Value:    5.0,
	}

	assert.True(t, cond.Matches(4.0))
	assert.True(t, cond.Matches(4.9999))
	assert.False(t, cond.Matches(5.0))
	assert.False(t, cond.Matches(6.0))
}

func TestMatches_Float64_LessThanOrEqual(t *testing.T) {
	cond := Config_Controller_Profile_Listener_Action_Condition{
		Operator: "lte",
		Value:    5.0,
	}

	assert.True(t, cond.Matches(4.0))
	assert.True(t, cond.Matches(5.0))
	assert.False(t, cond.Matches(6.0))
}

func TestMatches_Float64_UnknownOperator(t *testing.T) {
	cond := Config_Controller_Profile_Listener_Action_Condition{
		Operator: "neq",
		Value:    5.0,
	}

	assert.False(t, cond.Matches(5.0))
	assert.False(t, cond.Matches(6.0))
}

func TestMatches_Float64_TypeMismatch(t *testing.T) {
	cond := Config_Controller_Profile_Listener_Action_Condition{
		Operator: "eq",
		Value:    5.0,
	}

	assert.False(t, cond.Matches("5"))
	assert.False(t, cond.Matches(true))
	assert.False(t, cond.Matches(nil))
}

func TestMatches_String_Equal(t *testing.T) {
	cond := Config_Controller_Profile_Listener_Action_Condition{
		Operator: "eq",
		Value:    "hello",
	}

	assert.True(t, cond.Matches("hello"))
	assert.False(t, cond.Matches("world"))
	assert.False(t, cond.Matches("Hello")) // case-sensitive
}

func TestMatches_String_UnknownOperator(t *testing.T) {
	cond := Config_Controller_Profile_Listener_Action_Condition{
		Operator: "gt",
		Value:    "hello",
	}

	assert.False(t, cond.Matches("world"))
	assert.False(t, cond.Matches("hello"))
}

func TestMatches_String_TypeMismatch(t *testing.T) {
	cond := Config_Controller_Profile_Listener_Action_Condition{
		Operator: "eq",
		Value:    "hello",
	}

	assert.False(t, cond.Matches(5))
	assert.False(t, cond.Matches(true))
	assert.False(t, cond.Matches(nil))
}

func TestMatches_Bool_Equal(t *testing.T) {
	cond := Config_Controller_Profile_Listener_Action_Condition{
		Operator: "eq",
		Value:    true,
	}

	assert.True(t, cond.Matches(true))
	assert.False(t, cond.Matches(false))

	cond2 := Config_Controller_Profile_Listener_Action_Condition{
		Operator: "eq",
		Value:    false,
	}

	assert.True(t, cond2.Matches(false))
	assert.False(t, cond2.Matches(true))
}

func TestMatches_Bool_UnknownOperator(t *testing.T) {
	cond := Config_Controller_Profile_Listener_Action_Condition{
		Operator: "gt",
		Value:    true,
	}

	assert.False(t, cond.Matches(true))
	assert.False(t, cond.Matches(false))
}

func TestMatches_Bool_TypeMismatch(t *testing.T) {
	cond := Config_Controller_Profile_Listener_Action_Condition{
		Operator: "eq",
		Value:    true,
	}

	assert.False(t, cond.Matches(1))
	assert.False(t, cond.Matches("true"))
	assert.False(t, cond.Matches(nil))
}

func TestMatches_UndefinedValue(t *testing.T) {
	cond := Config_Controller_Profile_Listener_Action_Condition{
		Operator: "eq",
		Value:    nil,
	}

	assert.False(t, cond.Matches(nil))
	assert.False(t, cond.Matches("anything"))
	assert.False(t, cond.Matches(1))
}
