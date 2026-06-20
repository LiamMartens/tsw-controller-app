package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAction_UnmarshalJSON_HIDOutputReport(t *testing.T) {
	input := `{"type":"hid_output_report","report_id":1,"mask":255,"operation":"or","conditions":[{"operator":"eq","value":1.0}]}`
	var a Config_Controller_Profile_Listener_Action
	err := json.Unmarshal([]byte(input), &a)
	assert.NoError(t, err)
	assert.NotNil(t, a.HIDOutputReport)
	assert.Equal(t, "eq", a.HIDOutputReport.Conditions[0].Operator)
	assert.Equal(t, 1.0, a.HIDOutputReport.Conditions[0].Value)
	assert.Equal(t, uint8(1), a.HIDOutputReport.ReportID)
	assert.Equal(t, uint64(255), a.HIDOutputReport.Mask)
}

func TestAction_UnmarshalJSON_InvalidType(t *testing.T) {
	input := `{"type":"unknown","operator":"eq","value":1.0}`
	var a Config_Controller_Profile_Listener_Action
	err := json.Unmarshal([]byte(input), &a)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid action type")
}

func TestAction_MarshalJSON_HIDReport(t *testing.T) {
	a := Config_Controller_Profile_Listener_Action{
		HIDOutputReport: &Config_Controller_Profile_Listener_Action_HIDOutputReport{
			Type: "hid_output_report",
			Config_Controller_Profile_Listener_SharedAction: Config_Controller_Profile_Listener_SharedAction{
				Conditions: []Config_Controller_Profile_Listener_Action_Condition{
					{Operator: "gt", Value: 0.75},
				},
			},
			ReportID:  42,
			Mask:      128,
			Operation: "or",
		},
	}
	data, err := json.Marshal(a)
	assert.NoError(t, err)

	var out Config_Controller_Profile_Listener_Action
	err = json.Unmarshal(data, &out)
	assert.NoError(t, err)
	assert.Equal(t, a, out)
}

func TestAction_MarshalJSON_NoType(t *testing.T) {
	a := Config_Controller_Profile_Listener_Action{}
	data, err := json.Marshal(a)
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "no valid action type found")
}

func TestListener_UnmarshalJSON_APIValue(t *testing.T) {
	input := `{"type":"api_value","path":"/api/status","values_key":"Status","actions":[{"type":"hid_output_report","report_id":1,"mask":255,"operation":"or","conditions":[{"operator":"eq","value":1.0}]}]}`
	var l Config_Controller_Profile_Listener
	err := json.Unmarshal([]byte(input), &l)
	assert.NoError(t, err)
	assert.NotNil(t, l.API)
	assert.Equal(t, "api_value", l.API.Type)
	assert.Equal(t, "/api/status", l.API.Path)
	assert.Equal(t, "Status", l.API.ValuesKey)
	assert.Equal(t, 1, len(l.API.Actions))
	assert.NotNil(t, l.API.Actions[0].HIDOutputReport)
	assert.Equal(t, "eq", l.API.Actions[0].HIDOutputReport.Conditions[0].Operator)
	assert.Equal(t, 1.0, l.API.Actions[0].HIDOutputReport.Conditions[0].Value)
	assert.Equal(t, uint8(1), l.API.Actions[0].HIDOutputReport.ReportID)
	assert.Equal(t, uint64(255), l.API.Actions[0].HIDOutputReport.Mask)
}

func TestListener_UnmarshalJSON_InvalidType(t *testing.T) {
	input := `{"type":"unknown","path":"/api/status","actions":[]}`
	var l Config_Controller_Profile_Listener
	err := json.Unmarshal([]byte(input), &l)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid listener type")
}

func TestListener_MarshalJSON_APIValue(t *testing.T) {
	l := Config_Controller_Profile_Listener{
		API: &Config_Controller_Profile_Listener_Type_APIValue{
			Type:      "api_value",
			Path:      "/api/status",
			ValuesKey: "items",
			Actions: []Config_Controller_Profile_Listener_Action{
				{
					HIDOutputReport: &Config_Controller_Profile_Listener_Action_HIDOutputReport{
						Type: "hid_output_report",
						Config_Controller_Profile_Listener_SharedAction: Config_Controller_Profile_Listener_SharedAction{
							Conditions: []Config_Controller_Profile_Listener_Action_Condition{
								{Operator: "eq", Value: 1.0},
							},
						},
						ReportID:  1,
						Mask:      255,
						Operation: "or",
					},
				},
			},
		},
	}
	data, err := json.Marshal(l)
	assert.NoError(t, err)

	var out Config_Controller_Profile_Listener
	err = json.Unmarshal(data, &out)
	assert.NoError(t, err)
	assert.Equal(t, out, l)
}

func TestListener_MarshalJSON_NoType(t *testing.T) {
	l := Config_Controller_Profile_Listener{}
	data, err := l.MarshalJSON()
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "no valid listener type found")
}
