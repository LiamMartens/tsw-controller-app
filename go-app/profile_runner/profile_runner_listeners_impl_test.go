package profile_runner

import (
	"testing"

	"tsw_controller_app/config"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestExecuteProfileListenerAction_HIDOutputReport_AND(t *testing.T) {
	runner := &ProfileRunner{}

	mockHID := new(mockHIDDevice)
	sdlDev := newSDLDeviceWithMockHID(t, mockHID)

	mockHID.On("GetOutputReport", uint8(1), uint8(1)).Return([]byte{0b00011101}, nil)
	mockHID.On("SetOutputReport", uint8(1), []byte{0b00001101}).Return(nil)
	action := config.Config_Controller_Profile_Listener_Action{
		HIDOutputReport: &config.Config_Controller_Profile_Listener_Action_HIDOutputReport{
			Type:         "hid_output_report",
			ReportID:     1,
			ReportLength: 1,
			Mask:         0b00001111,
			Operation:    "and",
		},
	}
	err := runner.executeProfileListenerAction(sdlDev, action)
	assert.NoError(t, err)
	mockHID.AssertExpectations(t)
}

func TestExecuteProfileListenerAction_HIDOutputReport_OR(t *testing.T) {
	runner := &ProfileRunner{}

	mockHID := new(mockHIDDevice)
	sdlDev := newSDLDeviceWithMockHID(t, mockHID)

	mockHID.On("GetOutputReport", uint8(1), uint8(1)).Return([]byte{0b00011101}, nil)
	mockHID.On("SetOutputReport", uint8(1), []byte{0b00011111}).Return(nil)
	action := config.Config_Controller_Profile_Listener_Action{
		HIDOutputReport: &config.Config_Controller_Profile_Listener_Action_HIDOutputReport{
			Type:         "hid_output_report",
			ReportID:     1,
			ReportLength: 1,
			Mask:         0b00001111,
			Operation:    "or",
		},
	}
	err := runner.executeProfileListenerAction(sdlDev, action)
	assert.NoError(t, err)
	mockHID.AssertExpectations(t)
}

func TestExecuteProfileListenerAction_HIDOutputReport_AND_MultiByte(t *testing.T) {
	runner := &ProfileRunner{}

	mockHID := new(mockHIDDevice)
	sdlDev := newSDLDeviceWithMockHID(t, mockHID)

	mockHID.On("GetOutputReport", uint8(1), uint8(2)).Return([]byte{0b00011101, 0b10000001}, nil)
	mockHID.On("SetOutputReport", uint8(1), []byte{0b00001101, 0b10000000}).Return(nil)
	action := config.Config_Controller_Profile_Listener_Action{
		HIDOutputReport: &config.Config_Controller_Profile_Listener_Action_HIDOutputReport{
			Type:         "hid_output_report",
			ReportID:     1,
			ReportLength: 2,
			Mask:         0b10000000_00001111,
			Operation:    "and",
		},
	}
	err := runner.executeProfileListenerAction(sdlDev, action)
	assert.NoError(t, err)
	mockHID.AssertExpectations(t)
}

func TestExecuteProfileListenerAction_HIDFeatureReport_AND(t *testing.T) {
	runner := &ProfileRunner{}

	mockHID := new(mockHIDDevice)
	sdlDev := newSDLDeviceWithMockHID(t, mockHID)

	mockHID.On("ReadFeatureReport", uint8(1)).Return([]byte{0b00011101}, nil)
	mockHID.On("SendFeatureReport", uint8(1), []byte{0b00001101}).Return(nil)
	action := config.Config_Controller_Profile_Listener_Action{
		HIDFeatureReport: &config.Config_Controller_Profile_Listener_Action_HIDFeatureReport{
			Type:      "hid_feature_report",
			ReportID:  1,
			Mask:      0b00001111,
			Operation: "and",
		},
	}
	err := runner.executeProfileListenerAction(sdlDev, action)
	assert.NoError(t, err)
	mockHID.AssertExpectations(t)
}

func TestExecuteProfileListenerAction_HIDFeatureReport_OR(t *testing.T) {
	runner := &ProfileRunner{}

	mockHID := new(mockHIDDevice)
	sdlDev := newSDLDeviceWithMockHID(t, mockHID)

	mockHID.On("ReadFeatureReport", uint8(1)).Return([]byte{0b00011101}, nil)
	mockHID.On("SendFeatureReport", uint8(1), []byte{0b00011111}).Return(nil)
	action := config.Config_Controller_Profile_Listener_Action{
		HIDFeatureReport: &config.Config_Controller_Profile_Listener_Action_HIDFeatureReport{
			Type:      "hid_feature_report",
			ReportID:  1,
			Mask:      0b00001111,
			Operation: "or",
		},
	}
	err := runner.executeProfileListenerAction(sdlDev, action)
	assert.NoError(t, err)
	mockHID.AssertExpectations(t)
}

func TestExecuteProfileListenerAction_HIDFeatureReport_AND_MultiByte(t *testing.T) {
	runner := &ProfileRunner{}

	mockHID := new(mockHIDDevice)
	sdlDev := newSDLDeviceWithMockHID(t, mockHID)

	mockHID.On("ReadFeatureReport", uint8(1)).Return([]byte{0b00011101, 0b10000000}, nil)
	mockHID.On("SendFeatureReport", uint8(1), []byte{0b00001101, 0b10000000}).Return(nil)
	action := config.Config_Controller_Profile_Listener_Action{
		HIDFeatureReport: &config.Config_Controller_Profile_Listener_Action_HIDFeatureReport{
			Type:      "hid_feature_report",
			ReportID:  1,
			Mask:      0b10000001_00001111,
			Operation: "and",
		},
	}
	err := runner.executeProfileListenerAction(sdlDev, action)
	assert.NoError(t, err)
	mockHID.AssertExpectations(t)
}
