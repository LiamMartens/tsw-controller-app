package profile_runner

import (
	"testing"
	"tsw_controller_app/controller_mgr"
	"tsw_controller_app/sdl_mgr"

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
