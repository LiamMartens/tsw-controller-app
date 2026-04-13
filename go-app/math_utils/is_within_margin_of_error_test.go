package math_utils

import (
	"testing"
)

func TestIsWithinMarginOfError(t *testing.T) {
	tests := []struct {
		name     string
		value1   float64
		value2   float64
		expected bool
	}{
		{
			name:     "Exactly equal",
			value1:   1.0,
			value2:   1.0,
			expected: true,
		},
		{
			name:     "Within margin of error (0.0001)",
			value1:   1.0,
			value2:   1.00005,
			expected: true,
		},
		{
			name:     "Just inside the margin of error",
			value1:   1.0,
			value2:   1.0001,
			expected: true,
		},
		{
			name:     "Just outside the margin of error",
			value1:   1.0,
			value2:   1.00011,
			expected: false,
		},
		{
			name:     "Significantly different",
			value1:   1.0,
			value2:   2.0,
			expected: false,
		},
		{
			name:     "Negative values exactly equal",
			value1:   -1.0,
			value2:   -1.0,
			expected: true,
		},
		{
			name:     "Negative values within margin",
			value1:   -1.0,
			value2:   -1.00005,
			expected: true,
		},
		{
			name:     "Negative values just outside margin",
			value1:   -1.0,
			value2:   -1.00011,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsWithinMarginOfError(tt.value1, tt.value2)
			if got != tt.expected {
				t.Errorf("IsWithinMarginOfError(%v, %v) = %v; want %v", tt.value1, tt.value2, got, tt.expected)
			}
		})
	}
}
