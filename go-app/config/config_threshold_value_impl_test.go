package config

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Threshold_Value_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantValue     float64
		wantReference *string
		wantErr       bool
	}{
		{
			name:          "unmarshal object with value and reference",
			input:         `{"value": 0.5, "reference": "test_ref"}`,
			wantValue:     0.5,
			wantReference: func() *string { s := "test_ref"; return &s }(),
			wantErr:       false,
		},
		{
			name:          "unmarshal object with value only",
			input:         `{"value": 0.75}`,
			wantValue:     0.75,
			wantReference: nil,
			wantErr:       false,
		},
		{
			name:          "unmarshal number format",
			input:         `0.25`,
			wantValue:     0.25,
			wantReference: nil,
			wantErr:       false,
		},
		{
			name:          "unmarshal integer as number",
			input:         `100`,
			wantValue:     100,
			wantReference: nil,
			wantErr:       false,
		},
		{
			name:          "unmarshal invalid JSON string",
			input:         `"invalid"`,
			wantValue:     0,
			wantReference: nil,
			wantErr:       true,
		},
		{
			name:          "unmarshal empty object",
			input:         `{}`,
			wantValue:     0,
			wantReference: nil,
			wantErr:       true,
		},
		{
			name:          "unmarshal null",
			input:         `null`,
			wantValue:     0,
			wantReference: nil,
			wantErr:       true,
		},
		{
			name:          "unmarshal invalid JSON syntax",
			input:         `{invalid}`,
			wantValue:     0,
			wantReference: nil,
			wantErr:       true,
		},
		{
			name:          "unmarshal object with empty reference",
			input:         `{"value": 0.5, "reference": ""}`,
			wantValue:     0.5,
			wantReference: func() *string { s := ""; return &s }(),
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v Config_Threshold_Value
			err := v.UnmarshalJSON([]byte(tt.input))

			if tt.wantErr {
				assert.Error(t, err, "Expected to receive error for JSON: %s", tt.input)
				return
			}

			assert.Equal(t, tt.wantValue, v.Value, "Expected %v, received %v", tt.wantValue, v.Value)
			assert.Equal(t, tt.wantReference, v.Reference, "Expected %v, received %v", tt.wantReference, v.Reference)
		})
	}
}

func TestConfig_Threshold_Value_MarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		reference *string
		wantJSON  string
	}{
		{
			name:      "marshal number format (no reference)",
			value:     0.5,
			reference: nil,
			wantJSON:  `0.5`,
		},
		{
			name:      "marshal object format (with reference)",
			value:     0.75,
			reference: func() *string { s := "test_ref"; return &s }(),
			wantJSON:  `{"value":0.75,"reference":"test_ref"}`,
		},
		{
			name:      "marshal zero value (no reference)",
			value:     0,
			reference: nil,
			wantJSON:  `0`,
		},
		{
			name:      "marshal object with zero value and reference",
			value:     0,
			reference: func() *string { s := "ref"; return &s }(),
			wantJSON:  `{"value":0,"reference":"ref"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Config_Threshold_Value{
				Value:     tt.value,
				Reference: tt.reference,
			}

			gotJSON, _ := v.MarshalJSON()

			wantJSONdata := map[string]any{}
			gotJSONdata := map[string]any{}
			json.Unmarshal([]byte(tt.wantJSON), &wantJSONdata)
			json.Unmarshal(gotJSON, &gotJSONdata)

			assert.Equal(t, wantJSONdata, gotJSONdata, fmt.Sprintf("Expected %s, received %s", tt.wantJSON, string(gotJSON)))
		})
	}
}
