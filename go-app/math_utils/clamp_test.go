package math_utils

import (
	"testing"
)

func TestClamp(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		tests := []struct {
			name     string
			value    int
			min      int
			max      int
			expected int
		}{
			{"Value within range", 5, 0, 10, 5},
			{"Value below minimum", -5, 0, 10, 0},
			{"Value above maximum", 15, 0, 10, 10},
			{"Value exactly at minimum", 0, 0, 10, 0},
			{"Value exactly at maximum", 10, 0, 10, 10},
			{"Generic (Int) - Large values", 100, 50, 150, 100},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := Clamp(tt.value, tt.min, tt.max)
				if result != tt.expected {
					t.Errorf("Clamp(%d, %d, %d) = %d; want %d", tt.value, tt.min, tt.max, result, tt.expected)
				}
			})
		}
	})

	t.Run("float64", func(t *testing.T) {
		tests := []struct {
			name     string
			value    float64
			min      float64
			max      float64
			expected float64
		}{
			{"Value within range", 5.5, 1.1, 9.9, 5.5},
			{"Value below minimum", 0.5, 1.1, 9.9, 1.1},
			{"Value above maximum", 10.5, 1.1, 9.9, 9.9},
			{"Value exactly at minimum", 1.1, 1.1, 9.9, 1.1},
			{"Value exactly at maximum", 9.9, 1.1, 9.9, 9.9},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := Clamp(tt.value, tt.min, tt.max)
				if result != tt.expected {
					t.Errorf("Clamp(%f, %f, %f) = %f; want %f", tt.value, tt.min, tt.max, result, tt.expected)
				}
			})
		}
	})
}
