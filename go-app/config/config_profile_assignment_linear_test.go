package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigProfile_LinearThreshold_IsExceedingThreshold(t *testing.T) {

	t.Run("simple value", func(t *testing.T) {
		threshold := Config_Controller_Profile_Control_Assignment_Linear_Threshold{
			Value: Config_Threshold_Value{Value: 0.5},
		}
		// 0.51 should exceed 0.5
		assert.True(t, threshold.IsExceedingThreshold(0.51, map[string]float64{}))

		// 0.49 should not exceed 0.49
		assert.False(t, threshold.IsExceedingThreshold(0.49, map[string]float64{}))
	})

	t.Run("with named thresholds", func(t *testing.T) {
		named_thresholds := map[string]float64{"Center": 0.75}
		threshold := Config_Controller_Profile_Control_Assignment_Linear_Threshold{
			Value: Config_Threshold_Value{Value: 0.5, Reference: stringPtr("Center")},
		}
		// 0.76 should exceed 0.75
		assert.True(t, threshold.IsExceedingThreshold(0.76, named_thresholds))

		// 0.74 should not exceed 0.75
		assert.False(t, threshold.IsExceedingThreshold(0.74, named_thresholds))
	})
}

func TestConfigProfile_Linear_GenerateThresholds(t *testing.T) {
	t.Run("simple generate", func(t *testing.T) {
		linear := Config_Controller_Profile_Control_Assignment_Linear{
			Type: "linear",
			Thresholds: []Config_Controller_Profile_Control_Assignment_Linear_Threshold{
				// simple thresholds
				{Value: Config_Threshold_Value{Value: 0.0}},
				{Value: Config_Threshold_Value{Value: 0.1}},
				{Value: Config_Threshold_Value{Value: 0.2}},
				// auto step threshold - value end is exclusive
				{Value: Config_Threshold_Value{Value: 0.3}, ValueEnd: &Config_Threshold_Value{Value: 0.4}, ValueStep: floatPtr(0.03)},
				{Value: Config_Threshold_Value{Value: 0.6}},
			},
		}

		thresholds := linear.GenerateThresholds(map[string]float64{})

		// should have generated 8 thresholds
		assert.Len(t, thresholds, 8)
		assert.Equal(t, thresholds[0].Value.GetValue(map[string]float64{}, false), 0.0)
		assert.Equal(t, thresholds[1].Value.GetValue(map[string]float64{}, false), 0.1)
		assert.Equal(t, thresholds[2].Value.GetValue(map[string]float64{}, false), 0.2)
		assert.Equal(t, thresholds[3].Value.GetValue(map[string]float64{}, false), 0.3)
		assert.Equal(t, thresholds[4].Value.GetValue(map[string]float64{}, false), 0.33)
		assert.Equal(t, thresholds[5].Value.GetValue(map[string]float64{}, false), 0.36)
		assert.Equal(t, thresholds[6].Value.GetValue(map[string]float64{}, false), 0.39)
		assert.Equal(t, thresholds[7].Value.GetValue(map[string]float64{}, false), 0.6)
	})

	t.Run("with named thresholds", func(t *testing.T) {
		named_thresholds := map[string]float64{
			"Min":    0.1,
			"Middle": 0.6,
			"Max":    0.9,
		}
		linear := Config_Controller_Profile_Control_Assignment_Linear{
			Type: "linear",
			Thresholds: []Config_Controller_Profile_Control_Assignment_Linear_Threshold{
				{Value: Config_Threshold_Value{Value: 0.0, Reference: stringPtr("Min")}},
				{Value: Config_Threshold_Value{Value: 0.5, Reference: stringPtr("Middle")}},
				{Value: Config_Threshold_Value{Value: 1.0, Reference: stringPtr("Max")}},
			},
		}

		thresholds := linear.GenerateThresholds(named_thresholds)

		// should have generated 8 thresholds
		assert.Len(t, thresholds, 3)
		assert.Equal(t, thresholds[0].Value.Value, 0.1)
		assert.Equal(t, thresholds[1].Value.Value, 0.6)
		assert.Equal(t, thresholds[2].Value.Value, 0.9)
	})
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
