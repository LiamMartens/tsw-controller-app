package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigProfile_DirectLike_InputValue_GetStepsList(t *testing.T) {
	input_value := Config_Controller_Profile_Control_Assignment_DirectLike_InputValue{
		Min:   0.0,
		Max:   1.0,
		Steps: &[]*float64{floatPtr(0.0), nil, nil, floatPtr(0.5), floatPtr(0.5), nil, nil, floatPtr(1.0)},
	}
	steps_list := input_value.GetStepsList()
	assert.Len(t, steps_list, 5)
	assert.Equal(t, steps_list[0], StepList_Item{
		IsFreeRange:   false,
		Value:         floatPtr(0.0),
		PreviousValue: nil,
		NextValue:     nil,
	})
	assert.Equal(t, steps_list[1], StepList_Item{
		IsFreeRange:   true,
		Value:         nil,
		PreviousValue: floatPtr(0.0),
		NextValue:     floatPtr(0.5),
	})
	assert.Equal(t, steps_list[2], StepList_Item{
		IsFreeRange:   false,
		Value:         floatPtr(0.5),
		PreviousValue: nil,
		NextValue:     nil,
	})
	assert.Equal(t, steps_list[3], StepList_Item{
		IsFreeRange:   true,
		Value:         nil,
		PreviousValue: floatPtr(0.5),
		NextValue:     floatPtr(1.0),
	})
	assert.Equal(t, steps_list[4], StepList_Item{
		IsFreeRange:   false,
		Value:         floatPtr(1.0),
		PreviousValue: nil,
		NextValue:     nil,
	})
}

func TestConfigProfile_DirectLike_InputValue_GetSteps_AutomticFromStep(t *testing.T) {
	input_value := Config_Controller_Profile_Control_Assignment_DirectLike_InputValue{
		Min:  0.0,
		Max:  1.0,
		Step: floatPtr(0.2),
	}

	steps := input_value.GetSteps(map[string]float64{})
	assert.Len(t, steps, 6)
	assert.Equal(t, 0.0, steps[0].ValueStart)
	assert.Equal(t, 0.2, steps[1].ValueStart)
	assert.Equal(t, 0.4, steps[2].ValueStart)
	assert.Equal(t, 0.6, steps[3].ValueStart)
	assert.Equal(t, 0.8, steps[4].ValueStart)
	assert.Equal(t, 1.0, steps[5].ValueStart)
}

func TestConfigProfile_DirectLike_InputValue_GetSteps_PreDefinedStepsWithFreeRange(t *testing.T) {
	input_value := Config_Controller_Profile_Control_Assignment_DirectLike_InputValue{
		Min:   0.0,
		Max:   1.0,
		Steps: &[]*float64{floatPtr(0.0), nil, floatPtr(0.5), floatPtr(0.75), nil},
	}

	steps := input_value.GetSteps(map[string]float64{})
	assert.Len(t, steps, 5)
	assert.Equal(t, ControlStepDefinition{IsFreeRange: false, ValueStart: 0.0, ValueEnd: 0.0, Threshold: ControlStepDefinition_Threshold{ValueStart: 0.0, ValueEnd: 0.0, Tolerance: 0.125}}, steps[0])
	assert.Equal(t, ControlStepDefinition{IsFreeRange: true, ValueStart: 0.0, ValueEnd: 0.5, Threshold: ControlStepDefinition_Threshold{ValueStart: 0.0, ValueEnd: 0.5, Tolerance: 0}}, steps[1])
	assert.Equal(t, ControlStepDefinition{IsFreeRange: false, ValueStart: 0.5, ValueEnd: 0.5, Threshold: ControlStepDefinition_Threshold{ValueStart: 0.5, ValueEnd: 0.5, Tolerance: 0.125}}, steps[2])
	assert.Equal(t, ControlStepDefinition{IsFreeRange: false, ValueStart: 0.75, ValueEnd: 0.75, Threshold: ControlStepDefinition_Threshold{ValueStart: 0.75, ValueEnd: 0.75, Tolerance: 0.125}}, steps[3])
	assert.Equal(t, ControlStepDefinition{IsFreeRange: true, ValueStart: 0.75, ValueEnd: 1.0, Threshold: ControlStepDefinition_Threshold{ValueStart: 0.75, ValueEnd: 1.0, Tolerance: 0}}, steps[4])
}

func TestConfigProfile_DirectLike_InputValue_GetSteps_PreDefinedStepsWithFreeRangeAndThresholdDefs(t *testing.T) {
	input_value := Config_Controller_Profile_Control_Assignment_DirectLike_InputValue{
		Min:   0.0,
		Max:   1.0,
		Steps: &[]*float64{floatPtr(0.0), nil, floatPtr(0.5), floatPtr(0.75), nil},
		StepThresholds: &[]Config_Controller_Profile_Control_Assignment_DirectLike_InputValue_StepThreshold{
			{Threshold: Config_Threshold_Value{Value: 0.0}},
			{Threshold: Config_Threshold_Value{Value: 0.1}, ThresholdEnd: &Config_Threshold_Value{Value: 0.4}},
			{Threshold: Config_Threshold_Value{Value: 0.5}, ThresholdTolerance: floatPtr(0.05)},
			{Threshold: Config_Threshold_Value{Value: 0.75}, ThresholdTolerance: floatPtr(0.05)},
			{Threshold: Config_Threshold_Value{Value: 0.8}, ThresholdEnd: &Config_Threshold_Value{Value: 1.0}, ThresholdTolerance: floatPtr(0.05)},
		},
	}

	steps := input_value.GetSteps(map[string]float64{})
	assert.Len(t, steps, 5)
	assert.Equal(t, ControlStepDefinition{IsFreeRange: false, ValueStart: 0.0, ValueEnd: 0.0, Threshold: ControlStepDefinition_Threshold{ValueStart: 0.0, ValueEnd: 0.0, Tolerance: 0.125}}, steps[0])
	assert.Equal(t, ControlStepDefinition{IsFreeRange: true, ValueStart: 0.0, ValueEnd: 0.5, Threshold: ControlStepDefinition_Threshold{ValueStart: 0.1, ValueEnd: 0.4, Tolerance: 0.0}}, steps[1])
	assert.Equal(t, ControlStepDefinition{IsFreeRange: false, ValueStart: 0.5, ValueEnd: 0.5, Threshold: ControlStepDefinition_Threshold{ValueStart: 0.5, ValueEnd: 0.5, Tolerance: 0.05}}, steps[2])
	assert.Equal(t, ControlStepDefinition{IsFreeRange: false, ValueStart: 0.75, ValueEnd: 0.75, Threshold: ControlStepDefinition_Threshold{ValueStart: 0.75, ValueEnd: 0.75, Tolerance: 0.05}}, steps[3])
	assert.Equal(t, ControlStepDefinition{IsFreeRange: true, ValueStart: 0.75, ValueEnd: 1.0, Threshold: ControlStepDefinition_Threshold{ValueStart: 0.8, ValueEnd: 1.0, Tolerance: 0.05}}, steps[4])
}

func TestConfigProfile_DirectLike_InputValue_GetSteps_PreDefinedStepsWithFreeRange_WithNegativeValues(t *testing.T) {
	input_value := Config_Controller_Profile_Control_Assignment_DirectLike_InputValue{
		Min:   -1.0,
		Max:   1.0,
		Steps: &[]*float64{floatPtr(-1.0), nil, floatPtr(0.0), nil, floatPtr(1.0)},
	}

	steps := input_value.GetSteps(map[string]float64{})
	assert.Len(t, steps, 5)
	assert.Equal(t, ControlStepDefinition{IsFreeRange: false, ValueStart: -1.0, ValueEnd: -1.0, Threshold: ControlStepDefinition_Threshold{ValueStart: 0.0, ValueEnd: 0.0, Tolerance: 0.125}}, steps[0])
	assert.Equal(t, ControlStepDefinition{IsFreeRange: true, ValueStart: -1.0, ValueEnd: 0.0, Threshold: ControlStepDefinition_Threshold{ValueStart: 0.0, ValueEnd: 0.5, Tolerance: 0}}, steps[1])
	assert.Equal(t, ControlStepDefinition{IsFreeRange: false, ValueStart: 0.0, ValueEnd: 0.0, Threshold: ControlStepDefinition_Threshold{ValueStart: 0.5, ValueEnd: 0.5, Tolerance: 0.125}}, steps[2])
	assert.Equal(t, ControlStepDefinition{IsFreeRange: true, ValueStart: 0.0, ValueEnd: 1.0, Threshold: ControlStepDefinition_Threshold{ValueStart: 0.5, ValueEnd: 1.0, Tolerance: 0}}, steps[3])
	assert.Equal(t, ControlStepDefinition{IsFreeRange: false, ValueStart: 1.0, ValueEnd: 1.0, Threshold: ControlStepDefinition_Threshold{ValueStart: 1.0, ValueEnd: 1.0, Tolerance: 0.125}}, steps[4])
}

func TestConfigProfile_DirectLike_InputValue_GetSteps_WithNamedThresholds(t *testing.T) {
	input_value := Config_Controller_Profile_Control_Assignment_DirectLike_InputValue{
		Min:   0.0,
		Max:   1.0,
		Steps: &[]*float64{floatPtr(0.0), floatPtr(0.5), floatPtr(1.0)},
		StepThresholds: &[]Config_Controller_Profile_Control_Assignment_DirectLike_InputValue_StepThreshold{
			{Threshold: Config_Threshold_Value{Value: 0.0}},
			{Threshold: Config_Threshold_Value{Value: 0.5, Reference: stringPtr("middle")}},
			{Threshold: Config_Threshold_Value{Value: 1.0}},
		},
	}

	steps := input_value.GetSteps(map[string]float64{
		"middle": 0.75,
	})
	assert.Len(t, steps, 3)
	assert.Equal(t, ControlStepDefinition{IsFreeRange: false, ValueStart: 0.0, ValueEnd: 0.0, Threshold: ControlStepDefinition_Threshold{ValueStart: 0.0, ValueEnd: 0.0, Tolerance: 0.25}}, steps[0])
	assert.Equal(t, ControlStepDefinition{IsFreeRange: false, ValueStart: 0.5, ValueEnd: 0.5, Threshold: ControlStepDefinition_Threshold{ValueStart: 0.75, ValueEnd: 0.75, Tolerance: 0.25}}, steps[1])
	assert.Equal(t, ControlStepDefinition{IsFreeRange: false, ValueStart: 1.0, ValueEnd: 1.0, Threshold: ControlStepDefinition_Threshold{ValueStart: 1.0, ValueEnd: 1.0, Tolerance: 0.25}}, steps[2])
}
