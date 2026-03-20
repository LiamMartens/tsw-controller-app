package config

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func floatPtr(f float64) *float64 {
	return &f
}

func stringPtr(v string) *string {
	return &v
}

func TestConfigProfile_SyncControlValidation(t *testing.T) {
	value := Config_Controller_Profile_Control_Assignment_SyncControl{
		Type:       "sync_control",
		Identifier: "Throttle",
		InputValue: Config_Controller_Profile_Control_Assignment_DirectLike_InputValue{
			Min: 0.0,
			Max: 1.0,
		},
		ActionIncrease: Config_Controller_Profile_Control_Assignment_Action_Keys{
			Keys: "a",
		},
		ActionDecrease: Config_Controller_Profile_Control_Assignment_Action_Keys{
			Keys: "d",
		},
	}

	v := validator.New()
	assert.NoError(t, v.Struct(value))
}

func TestConfigProfile_ControlStepDefinition_Threshold(t *testing.T) {
	threshold_positive := ControlStepDefinition_Threshold{
		ValueStart: 0.0,
		ValueEnd:   0.5,
		Tolerance:  0.1,
	}
	assert.False(t, threshold_positive.IsWithinThreshold(-0.1))
	assert.True(t, threshold_positive.IsWithinThreshold(0.0))
	assert.True(t, threshold_positive.IsWithinThreshold(0.1))
	assert.True(t, threshold_positive.IsWithinThreshold(0.5))
	assert.True(t, threshold_positive.IsWithinThreshold(0.6))
	assert.False(t, threshold_positive.IsWithinThreshold(0.61))

	threshold_negative := ControlStepDefinition_Threshold{
		ValueStart: 0.0,
		ValueEnd:   -0.5,
		Tolerance:  0.1,
	}
	assert.False(t, threshold_negative.IsWithinThreshold(0.1))
	assert.True(t, threshold_negative.IsWithinThreshold(0.0))
	assert.True(t, threshold_negative.IsWithinThreshold(-0.1))
	assert.True(t, threshold_negative.IsWithinThreshold(-0.5))
	assert.True(t, threshold_negative.IsWithinThreshold(-0.6))
	assert.False(t, threshold_negative.IsWithinThreshold(-0.61))
}

func TestConfigProfile_DirectOrSyncControl_InputValue_CalculateOutputValue_Simple(t *testing.T) {
	input_value := Config_Controller_Profile_Control_Assignment_DirectLike_InputValue{
		Min:  0.0,
		Max:  1.0,
		Step: floatPtr(0.2),
	}
	assert.Equal(t, 0.0, *input_value.CalculateOutputValue(0.0, nil))
	assert.Equal(t, 0.0, *input_value.CalculateOutputValue(0.1, nil))
	assert.Equal(t, 0.2, *input_value.CalculateOutputValue(0.15, nil))
	assert.Equal(t, 0.2, *input_value.CalculateOutputValue(0.2, nil))
}

func TestConfigProfile_DirectOrSyncControl_InputValue_CalculateOutputValue_SimpleFreeRange(t *testing.T) {
	input_value := Config_Controller_Profile_Control_Assignment_DirectLike_InputValue{
		Min: 0.0,
		Max: 1.0,
		Steps: &[]*float64{
			floatPtr(0.0), floatPtr(0.5), nil, floatPtr(1.0),
		},
	}
	assert.Equal(t, 0.0, *input_value.CalculateOutputValue(0.0, nil))
	assert.Equal(t, 0.0, *input_value.CalculateOutputValue(0.1, nil))
	assert.Equal(t, 0.5, *input_value.CalculateOutputValue(0.2, nil))
	assert.Equal(t, 0.5, *input_value.CalculateOutputValue(0.3, nil))
	assert.Equal(t, 0.5, *input_value.CalculateOutputValue(0.3333, nil))
	assert.Equal(t, 0.7, *input_value.CalculateOutputValue(0.6, nil))
	assert.Equal(t, 0.7751, *input_value.CalculateOutputValue(0.7, nil))
	assert.Equal(t, 0.8501, *input_value.CalculateOutputValue(0.8, nil))
	assert.Equal(t, 0.9251, *input_value.CalculateOutputValue(0.9, nil))
	assert.Equal(t, 1.0, *input_value.CalculateOutputValue(1.0, nil))
}

func TestConfigProfile_DirectOrSyncControl_InputValue_CalculateOutputValue_SimpleFreeRange_CustomThresholds(t *testing.T) {
	input_value := Config_Controller_Profile_Control_Assignment_DirectLike_InputValue{
		Min: 0.0,
		Max: 1.0,
		Steps: &[]*float64{
			floatPtr(0.0), floatPtr(0.5), nil, floatPtr(1.0),
		},
		StepThresholds: &[]Config_Controller_Profile_Control_Assignment_DirectLike_InputValue_StepThreshold{
			{Threshold: Config_Threshold_Value{Value: 0.0}, ThresholdTolerance: floatPtr(0.05)},
			{Threshold: Config_Threshold_Value{Value: 0.5}, ThresholdTolerance: floatPtr(0.05)},
			{Threshold: Config_Threshold_Value{Value: 0.55}, ThresholdEnd: &Config_Threshold_Value{Value: 0.95}},
			{Threshold: Config_Threshold_Value{Value: 1.0}, ThresholdTolerance: floatPtr(0.05)},
		},
	}
	assert.Equal(t, 0.0, *input_value.CalculateOutputValue(0.0, nil))
	assert.Equal(t, 0.0, *input_value.CalculateOutputValue(0.05, nil))

	assert.Nil(t, input_value.CalculateOutputValue(0.1, nil))
	assert.Equal(t, 0.5, *input_value.CalculateOutputValue(0.5, nil))

	assert.Equal(t, 0.5, *input_value.CalculateOutputValue(0.55, nil))
	assert.Equal(t, 0.75, *input_value.CalculateOutputValue(0.75, nil))
	assert.Equal(t, 0.9375, *input_value.CalculateOutputValue(0.9, nil))
	assert.Equal(t, 1.0, *input_value.CalculateOutputValue(0.95, nil))

	assert.Equal(t, 1.0, *input_value.CalculateOutputValue(1.0, nil))
}

func TestConfigProfile_DirectOrSyncControl_InputValue_CalculateOutputValue_SimpleFreeRange_WithNegativeValues(t *testing.T) {
	input_value := Config_Controller_Profile_Control_Assignment_DirectLike_InputValue{
		Min: -1.0,
		Max: 1.0,
		Steps: &[]*float64{
			floatPtr(-1.0), nil, floatPtr(0), nil, floatPtr(1.0),
		},
	}
	assert.Equal(t, -1.0, *input_value.CalculateOutputValue(0.0, nil))
	assert.Equal(t, -0.5, *input_value.CalculateOutputValue(0.25, nil))
	assert.Equal(t, 0.0, *input_value.CalculateOutputValue(0.5, nil))
	assert.Equal(t, 0.5, *input_value.CalculateOutputValue(0.75, nil))
	assert.Equal(t, 1.0, *input_value.CalculateOutputValue(1.0, nil))
}

func TestConfigProfile_DirectOrSyncControl_InputValue_CalculateOutputValue_SimpleFreeRange_WithNegativeValues_CustomThresholdss(t *testing.T) {
	input_value := Config_Controller_Profile_Control_Assignment_DirectLike_InputValue{
		Min: -1.0,
		Max: 1.0,
		Steps: &[]*float64{
			floatPtr(-1.0), nil, floatPtr(0), nil, floatPtr(1.0),
		},
		StepThresholds: &[]Config_Controller_Profile_Control_Assignment_DirectLike_InputValue_StepThreshold{
			{Threshold: Config_Threshold_Value{Value: 0.0}},
			{Threshold: Config_Threshold_Value{Value: 0.0}, ThresholdEnd: &Config_Threshold_Value{Value: 0.5}},
			{Threshold: Config_Threshold_Value{Value: 0.5}},
			{Threshold: Config_Threshold_Value{Value: 0.5}, ThresholdEnd: &Config_Threshold_Value{Value: 1.0}},
			{Threshold: Config_Threshold_Value{Value: 1.0}},
		},
	}
	assert.Nil(t, input_value.CalculateOutputValue(-1.0, nil))
	assert.Equal(t, -1.0, *input_value.CalculateOutputValue(0.0, nil))
	assert.Equal(t, -0.5, *input_value.CalculateOutputValue(0.25, nil))
	assert.Equal(t, 0.0, *input_value.CalculateOutputValue(0.5, nil))
	assert.Equal(t, 1.0, *input_value.CalculateOutputValue(1.0, nil))
}
