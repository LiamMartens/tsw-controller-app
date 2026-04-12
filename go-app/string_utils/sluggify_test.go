package string_utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSluggify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"Hello_World", "hello-world"},
		{"Hello! World?", "hello-world"},
		{"  Hello World  ", "hello-world"},
		{"---Hello---World---", "hello-world"},
		{"123 Hello 456", "123-hello-456"},
		{"Special@#$%^&*Characters", "special-characters"},
		{"lowercase", "lowercase"},
		{"UPPERCASE", "uppercase"},
		{"MixedCase", "mixedcase"},
		{"", ""},
		{"---", ""},
		{"a b c", "a-b-c"},
		{"a...b...c", "a-b-c"},
		{"a___b___c", "a-b-c"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			actual := Sluggify(tt.input)
			assert.Equalf(t, tt.expected, actual, "Sluggify(%q) = %q; want %q", tt.input, actual, tt.expected)
		})
	}
}
