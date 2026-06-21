package profile_runner

import (
	"testing"

	"tsw_controller_app/config"
	"tsw_controller_app/controller_mgr"
	"tsw_controller_app/sdl_mgr"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ---------------------------------------------------------------------------
// Mock: ISDLMgr_HIDDevice
// ---------------------------------------------------------------------------

type mockHIDDevice struct {
	mock.Mock
}

func (m *mockHIDDevice) Version() uint16 { return 0 }
func (m *mockHIDDevice) Serial() string  { return "" }
func (m *mockHIDDevice) Open() error     { return nil }
func (m *mockHIDDevice) Close() error    { return nil }

func (m *mockHIDDevice) GetOutputReport(id byte, length uint8) ([]byte, error) {
	args := m.Called(id, length)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockHIDDevice) SetOutputReport(id byte, data []byte) error {
	args := m.Called(id, data)
	return args.Error(0)
}

func (m *mockHIDDevice) ReadFeatureReport(id byte) ([]byte, error) {
	args := m.Called(id)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockHIDDevice) SendFeatureReport(id byte, data []byte) error {
	args := m.Called(id, data)
	return args.Error(0)
}

// ---------------------------------------------------------------------------
// Mock: IControllerManager_Device (the device interface)
// ---------------------------------------------------------------------------

type mockDevice struct {
	mock.Mock
}

func (m *mockDevice) UniqueID() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockDevice) DeviceID() controller_mgr.DeviceUniqueID {
	args := m.Called()
	return args.Get(0).(controller_mgr.DeviceUniqueID)
}

func (m *mockDevice) Name() string {
	args := m.Called()
	return args.String(0)
}

// ---------------------------------------------------------------------------
// Helper: build a *SDLMgr_Joystick with a mock HID device
// ---------------------------------------------------------------------------

func newSDLDeviceWithMockHID(t *testing.T, hid *mockHIDDevice) *sdl_mgr.SDLMgr_Joystick {
	t.Helper()
	dev := &sdl_mgr.SDLMgr_Joystick{
		InstanceID: 1,
		HIDDevice:  hid,
	}
	return dev
}

// ---------------------------------------------------------------------------
// Helper: build a non-SDL device (plain mockDevice)
// ---------------------------------------------------------------------------

func newNonSDLDevice(t *testing.T) controller_mgr.IControllerManager_Device {
	t.Helper()
	dev := new(mockDevice)
	dev.On("DeviceID").Return(controller_mgr.DeviceUniqueID("non-sdl"))
	dev.On("Name").Return("Non-SDL Device")
	dev.On("VendorID").Return(uint16(0))
	dev.On("ProductID").Return(uint16(0))
	dev.On("GUID").Return("")
	dev.On("IsConnected").Return(true)
	dev.On("Controls").Return(nil)
	dev.On("VirtualControls").Return(nil)
	return dev
}

// ---------------------------------------------------------------------------
// Helper: build a SDL device with nil HIDDevice
// ---------------------------------------------------------------------------

func newSDLDeviceWithoutHID(t *testing.T) *sdl_mgr.SDLMgr_Joystick {
	t.Helper()
	return &sdl_mgr.SDLMgr_Joystick{
		InstanceID: 1,
		HIDDevice:  nil,
	}
}

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
