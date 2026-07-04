package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAction_UnmarshalJSON_HIDOutputReport(t *testing.T) {
	input := `{"type":"hid_output_report","report_id":1,"mask":[64],"operation":"or","conditions":[{"operator":"eq","value":1.0}]}`
	var a Config_Controller_Profile_Listener_Action
	err := json.Unmarshal([]byte(input), &a)
	assert.NoError(t, err)
	assert.NotNil(t, a.HIDOutputReport)
	assert.Equal(t, "eq", a.HIDOutputReport.Conditions[0].Operator)
	assert.Equal(t, 1.0, a.HIDOutputReport.Conditions[0].Value)
	assert.Equal(t, uint8(1), a.HIDOutputReport.ReportID)
	assert.Equal(t, []byte{64}, a.HIDOutputReport.Mask)
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
			Mask:      []byte{64},
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
