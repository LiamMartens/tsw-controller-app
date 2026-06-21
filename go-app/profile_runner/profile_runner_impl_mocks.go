package profile_runner

import (
	"testing"
	"tsw_controller_app/controller_mgr"
	"tsw_controller_app/map_utils"
	"tsw_controller_app/sdl_mgr"
	"tsw_controller_app/tswapi"

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

type mockAPI struct {
	mock.Mock
}

func (m *mockAPI) ListCurrentDrivableActor() (tswapi.TSWAPI_ListResponse, error) {
	args := m.Called()
	return args.Get(0).(tswapi.TSWAPI_ListResponse), args.Error(1)
}
func (m *mockAPI) GetCurrentDrivableActorObjectClass() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}
func (m *mockAPI) DeleteSubscription(id int) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *mockAPI) GetSubscription(id int) (map[string]any, error) {
	args := m.Called(id)
	return args.Get(0).(map[string]any), args.Error(1)
}
func (m *mockAPI) SetInteracting(control string, value float64) error {
	args := m.Called(control, value)
	return args.Error(0)
}
func (m *mockAPI) SetInputValue(control string, value float64) error {
	args := m.Called(control, value)
	return args.Error(0)
}
func (m *mockAPI) GetByPath(path string) (map[string]any, error) {
	args := m.Called(path)
	return args.Get(0).(map[string]any), args.Error(1)
}
func (m *mockAPI) GetInputValue(control string) (float64, error) {
	args := m.Called(control)
	return args.Get(0).(float64), args.Error(1)
}
func (m *mockAPI) GetIsButton(control string) (bool, error) {
	args := m.Called(control)
	return args.Get(0).(bool), args.Error(1)
}
func (m *mockAPI) GetActiveCab() (tswapi.TSWAPIActiveCab, error) {
	args := m.Called()
	return args.Get(0).(tswapi.TSWAPIActiveCab), args.Error(1)
}
func (m *mockAPI) CreateCurrentDrivableActorSubscription(id int) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *mockAPI) GetCurrentDrivableActorSubscription(id int) (tswapi.TSWAPI_GetCurrentDrivableActorSubscriptionResponse, error) {
	args := m.Called(id)
	return args.Get(0).(tswapi.TSWAPI_GetCurrentDrivableActorSubscriptionResponse), args.Error(1)
}
func (m *mockAPI) LoadAPIKey(path string) error {
	args := m.Called(path)
	return args.Error(0)
}
func (m *mockAPI) CanConnect() bool {
	args := m.Called()
	return args.Get(0).(bool)
}
func (m *mockAPI) Enabled() bool {
	args := m.Called()
	return args.Get(0).(bool)
}

type mockController struct {
	mock.Mock
}

func (m *mockController) Device() controller_mgr.IControllerManager_Device {
	args := m.Called()
	return args.Get(0).(controller_mgr.IControllerManager_Device)
}
func (m *mockController) Controls() *map_utils.LockMap[string, controller_mgr.IControllerManager_Controller_Control] {
	args := m.Called()
	return args.Get(0).(*map_utils.LockMap[string, controller_mgr.IControllerManager_Controller_Control])
}
func (m *mockController) VirtualControls() *map_utils.LockMap[string, controller_mgr.IControllerManager_Controller_Control] {
	args := m.Called()
	return args.Get(0).(*map_utils.LockMap[string, controller_mgr.IControllerManager_Controller_Control])
}
func (m *mockController) RegisterVirtualControl(name string, initialvalue float64) {
	m.Called(name, initialvalue)
}

type mockControl struct {
	mock.Mock
}

func (m *mockControl) Manager() controller_mgr.IControllerManager {
	args := m.Called()
	return args.Get(0).(controller_mgr.IControllerManager)
}
func (m *mockControl) Controller() controller_mgr.IControllerManager_Controller {
	args := m.Called()
	return args.Get(0).(controller_mgr.IControllerManager_Controller)
}
func (m *mockControl) Device() controller_mgr.IControllerManager_Device {
	args := m.Called()
	return args.Get(0).(controller_mgr.IControllerManager_Device)
}
func (m *mockControl) Name() string {
	args := m.Called()
	return args.String(0)
}
func (m *mockControl) UpdateValue(value float64, is_reset bool) {
	m.Called(value, is_reset)
}
func (m *mockControl) GetState() controller_mgr.ControllerManager_Controller_ControlState {
	args := m.Called()
	return args.Get(0).(controller_mgr.ControllerManager_Controller_ControlState)
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
