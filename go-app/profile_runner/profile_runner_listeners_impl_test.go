package profile_runner

import (
	"sync"
	"testing"

	"tsw_controller_app/cabdebugger"
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
			Type:      "hid_output_report",
			ReportID:  1,
			Mask:      []byte{0b00001111},
			Operation: "and",
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
			Type:      "hid_output_report",
			ReportID:  1,
			Mask:      []byte{0b00001111},
			Operation: "or",
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
			Type:      "hid_output_report",
			ReportID:  1,
			Mask:      []byte{0b00001111, 0b10000000},
			Operation: "and",
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

	mockHID.On("ReadFeatureReport", uint8(1), uint8(1)).Return([]byte{0b00011101}, nil)
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

	mockHID.On("ReadFeatureReport", uint8(1), uint8(1)).Return([]byte{0b00011101}, nil)
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

	mockHID.On("ReadFeatureReport", uint8(1), uint8(2)).Return([]byte{0b00011101, 0b10000000}, nil)
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

func TestFilterMatchingListenerActions_APIValueConditions_APIValueNotFound(t *testing.T) {
	runner := &ProfileRunner{}

	mockAPI := new(mockAPI)
	mockAPI.On("CanConnect").Return(true)
	mockAPI.On("GetByPath", "CurrentDrivableActor.Function.IS_TractionLocked").Return(map[string]any{}, nil)
	runner.API = mockAPI

	listener := &config.Config_Controller_Profile_Listener{
		Actions: []config.Config_Controller_Profile_Listener_Action{
			{
				HIDOutputReport: &config.Config_Controller_Profile_Listener_Action_HIDOutputReport{
					Type: "hid_output_report",
					Config_Controller_Profile_Listener_SharedAction: config.Config_Controller_Profile_Listener_SharedAction{
						Conditions: []config.Config_Controller_Profile_Listener_Action_Condition{
							{Type: "api_value", Name: "CurrentDrivableActor.Function.IS_TractionLocked.IsLocked", Operator: "eq", Value: true},
						},
					},
				},
			},
		},
	}

	actions, _ := runner.filterMatchingListenerActions(listener, nil)
	assert.Len(t, actions, 0)
	mockAPI.AssertExpectations(t)
}

func TestFilterMatchingListenerActions_APIValueConditions_AllConditionsMatch(t *testing.T) {
	runner := &ProfileRunner{}

	mockAPI := new(mockAPI)
	mockAPI.On("CanConnect").Return(true)
	mockAPI.On("GetByPath", "CurrentDrivableActor.Function.IS_TractionLocked").Return(map[string]any{"IsLocked": true}, nil)
	runner.API = mockAPI

	listener := &config.Config_Controller_Profile_Listener{
		Actions: []config.Config_Controller_Profile_Listener_Action{
			{
				HIDOutputReport: &config.Config_Controller_Profile_Listener_Action_HIDOutputReport{
					Type: "hid_output_report",
					Config_Controller_Profile_Listener_SharedAction: config.Config_Controller_Profile_Listener_SharedAction{
						Conditions: []config.Config_Controller_Profile_Listener_Action_Condition{
							{Type: "api_value", Name: "CurrentDrivableActor.Function.IS_TractionLocked.IsLocked", Operator: "eq", Value: true},
						},
					},
				},
			},
		},
	}

	actions, err := runner.filterMatchingListenerActions(listener, nil)
	assert.NoError(t, err)
	assert.Len(t, actions, 1)
	mockAPI.AssertExpectations(t)
}

func TestFilterMatchingAPIListenerActions_APIValueCondition_OneConditionFails(t *testing.T) {
	runner := &ProfileRunner{}

	mockAPI := new(mockAPI)
	mockAPI.On("CanConnect").Return(true)
	mockAPI.On("GetByPath", "CurrentDrivableActor.Function.Get_Speed").Return(map[string]any{"Speed": 5.0}, nil)
	runner.API = mockAPI

	listener := &config.Config_Controller_Profile_Listener{
		Actions: []config.Config_Controller_Profile_Listener_Action{
			{
				HIDOutputReport: &config.Config_Controller_Profile_Listener_Action_HIDOutputReport{
					Type: "hid_output_report",
					Config_Controller_Profile_Listener_SharedAction: config.Config_Controller_Profile_Listener_SharedAction{
						Conditions: []config.Config_Controller_Profile_Listener_Action_Condition{
							{Type: "api_value", Name: "CurrentDrivableActor.Function.Get_Speed.Speed", Operator: "eq", Value: 5.0},
							{Type: "api_value", Name: "CurrentDrivableActor.Function.Get_Speed.Speed", Operator: "eq", Value: 10.0},
						},
					},
				},
			},
		},
	}

	actions, err := runner.filterMatchingListenerActions(listener, nil)
	assert.NoError(t, err)
	assert.Len(t, actions, 0)
	mockAPI.AssertExpectations(t)
}

func TestFilterMatchingAPIListenerActions_APIValueCondition_AnyConditionMatches(t *testing.T) {
	runner := &ProfileRunner{}

	mockAPI := new(mockAPI)
	mockAPI.On("CanConnect").Return(true)
	mockAPI.On("GetByPath", "CurrentDrivableActor.Function.Get_Speed").Return(map[string]any{"Speed": 5.0}, nil)
	runner.API = mockAPI

	listener := &config.Config_Controller_Profile_Listener{
		Actions: []config.Config_Controller_Profile_Listener_Action{
			{
				HIDOutputReport: &config.Config_Controller_Profile_Listener_Action_HIDOutputReport{
					Type: "hid_output_report",
					Config_Controller_Profile_Listener_SharedAction: config.Config_Controller_Profile_Listener_SharedAction{
						ConditionsEvaluationStrategy: "any",
						Conditions: []config.Config_Controller_Profile_Listener_Action_Condition{
							{Type: "api_value", Name: "CurrentDrivableActor.Function.Get_Speed.Speed", Operator: "eq", Value: 5.0},
							{Type: "api_value", Name: "CurrentDrivableActor.Function.Get_Speed.Speed", Operator: "eq", Value: 10.0},
						},
					},
				},
			},
		},
	}

	actions, err := runner.filterMatchingListenerActions(listener, nil)
	assert.NoError(t, err)
	assert.Len(t, actions, 1)
	mockAPI.AssertExpectations(t)
}

func TestFilterMatchingControlListenerActions_ControlValueCondition_ControlValueMatches(t *testing.T) {
	runner := &ProfileRunner{}

	mockController := new(mockController)
	mockControl := new(mockControl)
	mockControl.On("GetState").Return(controller_mgr.ControllerManager_Controller_ControlState{NormalizedValues: controller_mgr.ControllerManager_Controller_ControlStateValues{Value: 0.75}})
	mockControls := map_utils.NewLockMap[string, controller_mgr.IControllerManager_Controller_Control]()
	mockControls.Set("axis0", mockControl)
	mockController.On("Controls").Return(mockControls)
	runner.VirtualControllerManager = &controller_mgr.VirtualControllerManager{}

	listener := &config.Config_Controller_Profile_Listener{
		Actions: []config.Config_Controller_Profile_Listener_Action{
			{
				HIDOutputReport: &config.Config_Controller_Profile_Listener_Action_HIDOutputReport{
					Type: "hid_output_report",
					Config_Controller_Profile_Listener_SharedAction: config.Config_Controller_Profile_Listener_SharedAction{
						Conditions: []config.Config_Controller_Profile_Listener_Action_Condition{
							{Type: "control_value", Name: "axis0", Operator: "gte", Value: 0.5},
						},
					},
				},
			},
		},
	}

	actions, err := runner.filterMatchingListenerActions(listener, mockController)
	assert.NoError(t, err)
	assert.Len(t, actions, 1)
	mockController.AssertExpectations(t)
	mockControl.AssertExpectations(t)
}

func TestFilterMatchingControlListenerActions_ControlValueCondition_ControlValueDoesNotMatch(t *testing.T) {
	runner := &ProfileRunner{}

	mockController := new(mockController)
	mockControl := new(mockControl)
	mockControl.On("GetState").Return(controller_mgr.ControllerManager_Controller_ControlState{NormalizedValues: controller_mgr.ControllerManager_Controller_ControlStateValues{Value: 0.75}})
	mockControls := map_utils.NewLockMap[string, controller_mgr.IControllerManager_Controller_Control]()
	mockControls.Set("axis0", mockControl)
	mockController.On("Controls").Return(mockControls)
	runner.VirtualControllerManager = &controller_mgr.VirtualControllerManager{}

	listener := &config.Config_Controller_Profile_Listener{
		Actions: []config.Config_Controller_Profile_Listener_Action{
			{
				HIDOutputReport: &config.Config_Controller_Profile_Listener_Action_HIDOutputReport{
					Type: "hid_output_report",
					Config_Controller_Profile_Listener_SharedAction: config.Config_Controller_Profile_Listener_SharedAction{
						Conditions: []config.Config_Controller_Profile_Listener_Action_Condition{
							{Type: "control_value", Name: "axis0", Operator: "lt", Value: 0.5},
						},
					},
				},
			},
		},
	}

	actions, err := runner.filterMatchingListenerActions(listener, mockController)
	assert.NoError(t, err)
	assert.Len(t, actions, 0)
	mockController.AssertExpectations(t)
	mockControl.AssertExpectations(t)
}

func TestFilterMatchingControlListenerActions_ControlValueCondition_AnyConditionMatches(t *testing.T) {
	runner := &ProfileRunner{}

	mockController := new(mockController)
	mockControl := new(mockControl)
	mockControl.On("GetState").Return(controller_mgr.ControllerManager_Controller_ControlState{NormalizedValues: controller_mgr.ControllerManager_Controller_ControlStateValues{Value: 0.75}})
	mockControls := map_utils.NewLockMap[string, controller_mgr.IControllerManager_Controller_Control]()
	mockControls.Set("axis0", mockControl)
	mockController.On("Controls").Return(mockControls)
	runner.VirtualControllerManager = &controller_mgr.VirtualControllerManager{}

	listener := &config.Config_Controller_Profile_Listener{
		Actions: []config.Config_Controller_Profile_Listener_Action{
			{
				HIDOutputReport: &config.Config_Controller_Profile_Listener_Action_HIDOutputReport{
					Type: "hid_output_report",
					Config_Controller_Profile_Listener_SharedAction: config.Config_Controller_Profile_Listener_SharedAction{
						ConditionsEvaluationStrategy: "any",
						Conditions: []config.Config_Controller_Profile_Listener_Action_Condition{
							{Type: "control_value", Name: "axis0", Operator: "gte", Value: 0.5},
							{Type: "control_value", Name: "axis0", Operator: "lt", Value: 0.5},
						},
					},
				},
			},
		},
	}

	actions, err := runner.filterMatchingListenerActions(listener, mockController)
	assert.NoError(t, err)
	assert.Len(t, actions, 1)
	mockController.AssertExpectations(t)
	mockControl.AssertExpectations(t)
}

func TestFilterMatchingCabStateListenerActions_CabStateValueCondition_CabStateValueMatches(t *testing.T) {
	runner := &ProfileRunner{}

	controls := map_utils.NewLockMap[cabdebugger.PropertyName, cabdebugger.CabDebugger_ControlState_Control]()
	runner.CabDebugger = &cabdebugger.CabDebugger{
		State: cabdebugger.CabDebugger_ControlState{
			Mutex:             sync.Mutex{},
			DrivableActorName: "",
			Controls:          controls,
		},
	}
	controls.Set("Throttle", cabdebugger.CabDebugger_ControlState_Control{
		CurrentValue: 0.5,
	})

	listener := &config.Config_Controller_Profile_Listener{
		Actions: []config.Config_Controller_Profile_Listener_Action{
			{
				HIDOutputReport: &config.Config_Controller_Profile_Listener_Action_HIDOutputReport{
					Type: "hid_output_report",
					Config_Controller_Profile_Listener_SharedAction: config.Config_Controller_Profile_Listener_SharedAction{
						Conditions: []config.Config_Controller_Profile_Listener_Action_Condition{
							{Type: "cab_state_value", Name: "Throttle", Operator: "gte", Value: 0.5},
						},
					},
				},
			},
		},
	}

	actions, err := runner.filterMatchingListenerActions(listener, nil)
	assert.NoError(t, err)
	assert.Len(t, actions, 1)
}

func TestFilterMatchingCabStateListenerActions_CabStateValueCondition_CabStateValueDoesNotMatch(t *testing.T) {
	runner := &ProfileRunner{}

	controls := map_utils.NewLockMap[cabdebugger.PropertyName, cabdebugger.CabDebugger_ControlState_Control]()
	runner.CabDebugger = &cabdebugger.CabDebugger{
		State: cabdebugger.CabDebugger_ControlState{
			Mutex:             sync.Mutex{},
			DrivableActorName: "",
			Controls:          controls,
		},
	}
	controls.Set("Throttle", cabdebugger.CabDebugger_ControlState_Control{
		CurrentValue: 0.4,
	})

	listener := &config.Config_Controller_Profile_Listener{
		Actions: []config.Config_Controller_Profile_Listener_Action{
			{
				HIDOutputReport: &config.Config_Controller_Profile_Listener_Action_HIDOutputReport{
					Type: "hid_output_report",
					Config_Controller_Profile_Listener_SharedAction: config.Config_Controller_Profile_Listener_SharedAction{
						Conditions: []config.Config_Controller_Profile_Listener_Action_Condition{
							{Type: "cab_state_value", Name: "Throttle", Operator: "gte", Value: 0.5},
						},
					},
				},
			},
		},
	}

	actions, err := runner.filterMatchingListenerActions(listener, nil)
	assert.NoError(t, err)
	assert.Len(t, actions, 0)
}

func TestFilterMatchingCabStateListenerActions_CabStateValueCondition_AnyConditionMatches(t *testing.T) {
	runner := &ProfileRunner{}

	controls := map_utils.NewLockMap[cabdebugger.PropertyName, cabdebugger.CabDebugger_ControlState_Control]()
	runner.CabDebugger = &cabdebugger.CabDebugger{
		State: cabdebugger.CabDebugger_ControlState{
			Mutex:             sync.Mutex{},
			DrivableActorName: "",
			Controls:          controls,
		},
	}
	controls.Set("Throttle", cabdebugger.CabDebugger_ControlState_Control{
		CurrentValue: 0.5,
	})

	listener := &config.Config_Controller_Profile_Listener{
		Actions: []config.Config_Controller_Profile_Listener_Action{
			{
				HIDOutputReport: &config.Config_Controller_Profile_Listener_Action_HIDOutputReport{
					Type: "hid_output_report",
					Config_Controller_Profile_Listener_SharedAction: config.Config_Controller_Profile_Listener_SharedAction{
						ConditionsEvaluationStrategy: "any",
						Conditions: []config.Config_Controller_Profile_Listener_Action_Condition{
							{Type: "cab_state_value", Name: "Throttle", Operator: "gte", Value: 0.5},
							{Type: "cab_state_value", Name: "Throttle", Operator: "lt", Value: 0.5},
						},
					},
				},
			},
		},
	}

	actions, err := runner.filterMatchingListenerActions(listener, nil)
	assert.NoError(t, err)
	assert.Len(t, actions, 1)
}
