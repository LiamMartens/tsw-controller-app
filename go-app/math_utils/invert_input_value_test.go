package math_utils

import (
	"testing"
)

func TestInvertInputValue(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		// Positive values in [0:1] range
		{"0.0 -> 1.0", 0.0, 1.0},
		{"0.5 -> 0.5", 0.5, 0.5},
		{"1.0 -> 0.0", 1.0, 0.0},

		// Negative values in [-1:0] range
		// Based on implementation: if value < 0.0 { return -1.0 - value }
		{"-0.5 -> -0.5", -0.5, -0.5},
		{"-1.0 -> 0.0", -1.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InvertInputValue(tt.input)
			if got != tt.expected {
				t.Errorf("InvertInputValue(%v) = %v; want %v", tt.input, got, tt.expected)
			}
		})
	}
}
