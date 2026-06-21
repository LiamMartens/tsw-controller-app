package profile_runner

import (
	"testing"

	"tsw_controller_app/config"
	"tsw_controller_app/controller_mgr"
	"tsw_controller_app/map_utils"

	"github.com/stretchr/testify/assert"
)

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
			Mask:         []byte{0b00001111},
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
			Mask:         []byte{0b00001111},
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
			Mask:         []byte{0b00001111, 0b10000000},
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
			Mask:      []byte{0b00001111},
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
			Mask:      []byte{0b00001111},
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
			Mask:      []byte{0b00001111, 0b10000001},
			Operation: "and",
		},
	}
	err := runner.executeProfileListenerAction(sdlDev, action)
	assert.NoError(t, err)
	mockHID.AssertExpectations(t)
}

// --- filterMatchingAPIListenerActions ---

func TestFilterMatchingAPIListenerActions_APIValueNotFound(t *testing.T) {
	runner := &ProfileRunner{}

	mockAPI := new(mockAPI)
	mockAPI.On("GetByPath", "CurrentDrivableActor.Function.IS_TractionLocked").Return(map[string]any{}, nil)
	runner.API = mockAPI

	listener := &config.Config_Controller_Profile_Listener{
		API: &config.Config_Controller_Profile_Listener_Type_APIValue{
			Path:      "CurrentDrivableActor.Function.IS_TractionLocked",
			ValuesKey: "IsLocked",
			Actions: []config.Config_Controller_Profile_Listener_Action{
				{HIDOutputReport: &config.Config_Controller_Profile_Listener_Action_HIDOutputReport{Type: "hid_output_report"}},
			},
		},
	}

	_, err := runner.filterMatchingAPIListenerActions(listener)
	assert.ErrorIs(t, err, ErrNoListenerValue)
	mockAPI.AssertExpectations(t)
}

func TestFilterMatchingAPIListenerActions_AllConditionsMatch(t *testing.T) {
	runner := &ProfileRunner{}

	mockAPI := new(mockAPI)
	mockAPI.On("GetByPath", "CurrentDrivableActor.Function.IS_TractionLocked").Return(map[string]any{"IsLocked": true}, nil)
	runner.API = mockAPI

	listener := &config.Config_Controller_Profile_Listener{
		API: &config.Config_Controller_Profile_Listener_Type_APIValue{
			Path:      "CurrentDrivableActor.Function.IS_TractionLocked",
			ValuesKey: "IsLocked",
			Actions: []config.Config_Controller_Profile_Listener_Action{
				{
					HIDOutputReport: &config.Config_Controller_Profile_Listener_Action_HIDOutputReport{
						Config_Controller_Profile_Listener_SharedAction: config.Config_Controller_Profile_Listener_SharedAction{
							Conditions: []config.Config_Controller_Profile_Listener_Action_Condition{
								{Operator: "eq", Value: true},
							},
						},
						Type: "hid_output_report",
					},
				},
			},
		},
	}

	actions, err := runner.filterMatchingAPIListenerActions(listener)
	assert.NoError(t, err)
	assert.Len(t, actions, 1)
	mockAPI.AssertExpectations(t)
}

func TestFilterMatchingAPIListenerActions_OneConditionFails(t *testing.T) {
	runner := &ProfileRunner{}

	mockAPI := new(mockAPI)
	mockAPI.On("GetByPath", "CurrentDrivableActor.Function.Get_Speed").Return(map[string]any{"Speed": 5.0}, nil)
	runner.API = mockAPI

	listener := &config.Config_Controller_Profile_Listener{
		API: &config.Config_Controller_Profile_Listener_Type_APIValue{
			Path:      "CurrentDrivableActor.Function.Get_Speed",
			ValuesKey: "Speed",
			Actions: []config.Config_Controller_Profile_Listener_Action{
				{
					HIDOutputReport: &config.Config_Controller_Profile_Listener_Action_HIDOutputReport{
						Config_Controller_Profile_Listener_SharedAction: config.Config_Controller_Profile_Listener_SharedAction{
							Conditions: []config.Config_Controller_Profile_Listener_Action_Condition{
								{Operator: "eq", Value: 5.0},
								{Operator: "eq", Value: 10.0},
							},
						},
						Type: "hid_output_report",
					},
				},
			},
		},
	}

	actions, err := runner.filterMatchingAPIListenerActions(listener)
	assert.NoError(t, err)
	assert.Len(t, actions, 0)
	mockAPI.AssertExpectations(t)
}

// --- filterMatchingControlListenerActions ---

func TestFilterMatchingControlListenerActions_ControlValueMatches(t *testing.T) {
	runner := &ProfileRunner{}

	mockController := new(mockController)
	mockControl := new(mockControl)
	mockControl.On("GetState").Return(controller_mgr.ControllerManager_Controller_ControlState{NormalizedValues: controller_mgr.ControllerManager_Controller_ControlStateValues{Value: 0.75}})
	mockControls := map_utils.NewLockMap[string, controller_mgr.IControllerManager_Controller_Control]()
	mockControls.Set("axis0", mockControl)
	mockController.On("Controls").Return(mockControls)
	runner.VirtualControllerManager = &controller_mgr.VirtualControllerManager{}

	listener := &config.Config_Controller_Profile_Listener{
		Control: &config.Config_Controller_Profile_Listener_Type_ControlValue{
			Name: "axis0",
			Actions: []config.Config_Controller_Profile_Listener_Action{
				{
					HIDOutputReport: &config.Config_Controller_Profile_Listener_Action_HIDOutputReport{
						Config_Controller_Profile_Listener_SharedAction: config.Config_Controller_Profile_Listener_SharedAction{
							Conditions: []config.Config_Controller_Profile_Listener_Action_Condition{
								{Operator: "gte", Value: 0.5},
							},
						},
						Type: "hid_output_report",
					},
				},
			},
		},
	}

	actions, err := runner.filterMatchingControlListenerActions(listener, mockController)
	assert.NoError(t, err)
	assert.Len(t, actions, 1)
	mockController.AssertExpectations(t)
	mockControl.AssertExpectations(t)
}

func TestFilterMatchingControlListenerActions_ControlValueDoesNotMatch(t *testing.T) {
	runner := &ProfileRunner{}

	mockController := new(mockController)
	mockControl := new(mockControl)
	mockControl.On("GetState").Return(controller_mgr.ControllerManager_Controller_ControlState{NormalizedValues: controller_mgr.ControllerManager_Controller_ControlStateValues{Value: 0.75}})
	mockControls := map_utils.NewLockMap[string, controller_mgr.IControllerManager_Controller_Control]()
	mockControls.Set("axis0", mockControl)
	mockController.On("Controls").Return(mockControls)
	runner.VirtualControllerManager = &controller_mgr.VirtualControllerManager{}

	listener := &config.Config_Controller_Profile_Listener{
		Control: &config.Config_Controller_Profile_Listener_Type_ControlValue{
			Name: "axis0",
			Actions: []config.Config_Controller_Profile_Listener_Action{
				{
					HIDOutputReport: &config.Config_Controller_Profile_Listener_Action_HIDOutputReport{
						Config_Controller_Profile_Listener_SharedAction: config.Config_Controller_Profile_Listener_SharedAction{
							Conditions: []config.Config_Controller_Profile_Listener_Action_Condition{
								{Operator: "lt", Value: 0.5},
							},
						},
						Type: "hid_output_report",
					},
				},
			},
		},
	}

	actions, err := runner.filterMatchingControlListenerActions(listener, mockController)
	assert.NoError(t, err)
	assert.Len(t, actions, 0)
	mockController.AssertExpectations(t)
	mockControl.AssertExpectations(t)
}
