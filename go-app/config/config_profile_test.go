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

func TestConfigProfile_LinearThreshold_IsExceedingThreshold(t *testing.T) {
	threshold := Config_Controller_Profile_Control_Assignment_Linear_Threshold{
		Value: 0.5,
	}

	// 0.51 should exceed 0.5
	assert.True(t, threshold.IsExceedingThreshold(0.51))

	// 0.49 should not exceed 0.49
	assert.False(t, threshold.IsExceedingThreshold(0.49))
}

func TestConfigProfile_Linear_GenerateThresholds(t *testing.T) {
	var auto_step_threshold_value_end float64 = 0.4
	var auto_step_threshold_value_step float64 = 0.03

	linear := Config_Controller_Profile_Control_Assignment_Linear{
		Type: "linear",
		Thresholds: []Config_Controller_Profile_Control_Assignment_Linear_Threshold{
			// simple thresholds
			{Value: 0.0},
			{Value: 0.1},
			{Value: 0.2},
			// auto step threshold - value end is exclusive
			{Value: 0.3, ValueEnd: &auto_step_threshold_value_end, ValueStep: &auto_step_threshold_value_step},
			{Value: 0.6},
		},
	}

	thresholds := linear.GenerateThresholds()

	// should have generated 9 thresholds
	assert.Equal(t, thresholds[0].Value, 0.0)
	assert.Equal(t, thresholds[1].Value, 0.1)
	assert.Equal(t, thresholds[2].Value, 0.2)
	assert.Equal(t, thresholds[3].Value, 0.3)
	assert.Equal(t, thresholds[4].Value, 0.33)
	assert.Equal(t, thresholds[5].Value, 0.36)
	assert.Equal(t, thresholds[6].Value, 0.39)
	assert.Equal(t, thresholds[7].Value, 0.6)
}

func TestConfigProfile_Linear_CalculateNeutralizedValue(t *testing.T) {
	var neutral_value float64 = 0.5
	linear := Config_Controller_Profile_Control_Assignment_Linear{
		Type:       "linear",
		Neutral:    &neutral_value,
		Thresholds: []Config_Controller_Profile_Control_Assignment_Linear_Threshold{},
	}

	assert.Equal(t, 0.0, linear.CalculateNeutralizedValue(0.5))
	assert.Equal(t, -1.0, linear.CalculateNeutralizedValue(0))
	assert.Equal(t, 1.0, linear.CalculateNeutralizedValue(1))
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
