package math_utils

import (
	"testing"
)

func TestRoundToMarginOfError(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{
			name:     "Already rounded to 4 decimal places",
			input:    1.2345,
			expected: 1.2345,
		},
		{
			name:     "Needs rounding up",
			input:    1.234567,
			expected: 1.2346,
		},
		{
			name:     "Needs rounding down",
			input:    1.234543,
			expected: 1.2345,
		},
		{
			name:     "Many decimal places",
			input:    1.23456789012345,
			expected: 1.2346,
		},
		{
			name:     "Zero",
			input:    0.0,
			expected: 0.0,
		},
		{
			name:     "Negative value rounding up",
			input:    -1.234567,
			expected: -1.2346,
		},
		{
			name:     "Negative value rounding down",
			input:    -1.234543,
			expected: -1.2345,
		},
		{
			name:     "Very small value",
			input:    0.00001,
			expected: 0.0,
		},
		{
			name:     "Very large value",
			input:    123456789.1234567,
			expected: 123456789.1235,
		},
		{
			name:     "Boundary rounding up",
			input:    1.23455,
			expected: 1.2346,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RoundToMarginOfError(tt.input)
			// Using a small epsilon for float comparison to avoid precision issues in tests,
			// although since we are rounding to 4 decimal places, direct comparison should work.
			if got != tt.expected {
				t.Errorf("RoundToMarginOfError(%v) = %v; want %v", tt.input, got, tt.expected)
			}
		})
	}
}
