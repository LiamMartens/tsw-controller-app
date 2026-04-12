package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Controller_Profile_Control_Assignment_Linear_Threshold_Resolve(t *testing.T) {
	tests := []struct {
		name         string
		threshold    Config_Controller_Profile_Control_Assignment_Linear_Threshold
		thresholds   map[string]float64
		wantValue    float64
		wantValueEnd *float64
	}{
		{
			name: "ValueEnd is nil, Value is direct",
			threshold: Config_Controller_Profile_Control_Assignment_Linear_Threshold{
				Value: Config_Threshold_Value{Value: 0.5},
			},
			thresholds:   map[string]float64{},
			wantValue:    0.5,
			wantValueEnd: nil,
		},
		{
			name: "ValueEnd is nil, Value is reference",
			threshold: Config_Controller_Profile_Control_Assignment_Linear_Threshold{
				Value: Config_Threshold_Value{Value: 0.0, Reference: stringPtr("ref1")},
			},
			thresholds:   map[string]float64{"ref1": 0.7},
			wantValue:    0.7,
			wantValueEnd: nil,
		},
		{
			name: "ValueEnd is not nil, ValueEnd is direct",
			threshold: Config_Controller_Profile_Control_Assignment_Linear_Threshold{
				Value:    Config_Threshold_Value{Value: 0.5},
				ValueEnd: &Config_Threshold_Value{Value: 0.8},
			},
			thresholds:   map[string]float64{},
			wantValue:    0.5,
			wantValueEnd: floatPtr(0.8),
		},
		{
			name: "ValueEnd is not nil, ValueEnd is reference",
			threshold: Config_Controller_Profile_Control_Assignment_Linear_Threshold{
				Value:    Config_Threshold_Value{Value: 0.5},
				ValueEnd: &Config_Threshold_Value{Value: 0.0, Reference: stringPtr("ref2")},
			},
			thresholds:   map[string]float64{"ref2": 0.9},
			wantValue:    0.5,
			wantValueEnd: floatPtr(0.9),
		},
		{
			name: "ValueEnd is not nil, Value is reference, ValueEnd is direct",
			threshold: Config_Controller_Profile_Control_Assignment_Linear_Threshold{
				Value:    Config_Threshold_Value{Value: 0.0, Reference: stringPtr("ref1")},
				ValueEnd: &Config_Threshold_Value{Value: 0.8},
			},
			thresholds:   map[string]float64{"ref1": 0.3},
			wantValue:    0.3,
			wantValueEnd: floatPtr(0.8),
		},
		{
			name: "ValueEnd is not nil, Value is direct, ValueEnd is reference",
			threshold: Config_Controller_Profile_Control_Assignment_Linear_Threshold{
				Value:    Config_Threshold_Value{Value: 0.5},
				ValueEnd: &Config_Threshold_Value{Value: 0.0, Reference: stringPtr("ref2")},
			},
			thresholds:   map[string]float64{"ref2": 0.9},
			wantValue:    0.5,
			wantValueEnd: floatPtr(0.9),
		},
		{
			name: "ValueEnd is not nil, both are references",
			threshold: Config_Controller_Profile_Control_Assignment_Linear_Threshold{
				Value:    Config_Threshold_Value{Value: 0.0, Reference: stringPtr("ref1")},
				ValueEnd: &Config_Threshold_Value{Value: 0.0, Reference: stringPtr("ref2")},
			},
			thresholds:   map[string]float64{"ref1": 0.3, "ref2": 0.9},
			wantValue:    0.3,
			wantValueEnd: floatPtr(0.9),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved := tt.threshold.Resolve(tt.thresholds)
			assert.Equal(t, tt.wantValue, resolved.Value)
			if tt.wantValueEnd == nil {
				assert.Nil(t, resolved.ValueEnd)
			} else {
				assert.NotNil(t, resolved.ValueEnd)
				assert.Equal(t, *tt.wantValueEnd, *resolved.ValueEnd)
			}
		})
	}
}
