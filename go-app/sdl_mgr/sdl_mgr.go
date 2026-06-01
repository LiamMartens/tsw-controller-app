package sdl_mgr

import (
	"context"
	"crypto/sha1"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
	"tsw_controller_app/chan_utils"
	"tsw_controller_app/logger"

	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/Zyko0/go-sdl3/sdl"
	"rafaelmartins.com/p/usbhid"
)

/* the SDL control kind like Button, Hat, Axis */
type SDLMgr_Control_Kind = string
type SDLMgr_Guid_Str = string

const SDL_BUFFER_SIZE = 32

const (
	SDLMgr_Control_Kind_Button SDLMgr_Control_Kind = "button"
	SDLMgr_Control_Kind_Hat    SDLMgr_Control_Kind = "hat"
	SDLMgr_Control_Kind_Axis   SDLMgr_Control_Kind = "axis"
)

type SDLMgr_Joystick struct {
	InstanceID sdl.JoystickID
	name       string
	vendorID   int
	productID  int
	devicePath string

	InternalJoystick *sdl.Joystick
	HIDDevice        *usbhid.Device
}

type SDLMgr struct {
	Initialized bool
	Timestamp   time.Time

	joydevices_mutex sync.Mutex
	joydevices       map[sdl.JoystickID]*SDLMgr_Joystick
}

func New() *SDLMgr {
	return &SDLMgr{
		Initialized:      false,
		Timestamp:        time.Now(),
		joydevices_mutex: sync.Mutex{},
		joydevices:       map[sdl.JoystickID]*SDLMgr_Joystick{},
	}
}

func (mgr *SDLMgr) hidDeviceFromPath(path string) (*usbhid.Device, error) {
	path_lower := strings.ToLower(path)
	devices, err := usbhid.Enumerate(func(d *usbhid.Device) bool {
		return strings.ToLower(d.Path()) == path_lower
	})
	if err != nil {
		return nil, fmt.Errorf("could not find HID device from path due to an error: %s: %w", path, err)
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("could not find HID device from path: %s", path)
	}
	return devices[0], nil
}

func (mgr *SDLMgr) joyDeviceAdded(event *SDL_JoyDeviceAddedEvent) (*SDLMgr_Joystick, error) {
	mgr.joydevices_mutex.Lock()
	defer mgr.joydevices_mutex.Unlock()

	if joydevice, already_registered := mgr.joydevices[event.Which]; already_registered {
		return joydevice, nil
	}

	sdl_joystick, err := event.Which.OpenJoystick()
	if err != nil {
		return nil, fmt.Errorf("could not open joystick for use: %w", err)
	}

	name, _ := sdl_joystick.Name()
	usb_vendor := sdl_joystick.Vendor()
	usb_product := sdl_joystick.Product()
	device_path, _ := sdl_joystick.Path()
	var hid_device *usbhid.Device
	if device_path != "" {
		hid_device, err = mgr.hidDeviceFromPath(device_path)
		if err != nil {
			logger.Logger.Info("[SDLMgr] could not match HID device from path", "error", err)
		}
	}

	joystick := SDLMgr_Joystick{
		InstanceID:       event.Which,
		name:             name,
		vendorID:         int(usb_vendor),
		productID:        int(usb_product),
		devicePath:       device_path,
		InternalJoystick: sdl_joystick,
		HIDDevice:        hid_device,
	}

	mgr.joydevices[event.Which] = &joystick
	return &joystick, nil
}

func (mgr *SDLMgr) joyDeviceRemoved(event *SDL_JoyDeviceRemovedEvent) {
	mgr.joydevices_mutex.Lock()
	defer mgr.joydevices_mutex.Unlock()
	if joystick, has_device := mgr.joydevices[event.Which]; has_device {
		joystick.InternalJoystick.Close()
		delete(mgr.joydevices, event.Which)
	}
}

/*
Initializes the SDL library for the app
sdl.Init is guarded to only be ran once per app
*/
func (mgr *SDLMgr) PanicInit() bool {
	if !mgr.Initialized {
		init_ts := time.Now()

		/* try to initialize if not already initialized */
		sdl.SetJoystickEventsEnabled(true)
		sdl.SetHint(sdl.HINT_JOYSTICK_HIDAPI, "1")
		sdl.SetHint(sdl.HINT_JOYSTICK_RAWINPUT, "1")
		sdl.SetHint(sdl.HINT_JOYSTICK_WGI, "1")
		sdl.SetHint(sdl.HINT_XINPUT_ENABLED, "0")
		sdl.SetHint(sdl.HINT_HIDAPI_ENUMERATE_ONLY_CONTROLLERS, "0")
		if err := sdl.Init(sdl.INIT_GAMEPAD | sdl.INIT_JOYSTICK | sdl.INIT_EVENTS); err != nil {
			panic(err)
		}

		mgr.Initialized = true
		mgr.Timestamp = init_ts
	}

	return true
}

/* Just a passthrough for the sdl quit method */
func (mgr *SDLMgr) Quit() {
	sdl.Quit()
}

func (mgr *SDLMgr) GetJoystickByInstanceID(instance_id sdl.JoystickID) (*SDLMgr_Joystick, error) {
	mgr.joydevices_mutex.Lock()
	defer mgr.joydevices_mutex.Unlock()
	if joydevice, has_joydevice := mgr.joydevices[instance_id]; has_joydevice {
		return joydevice, nil
	}
	return nil, fmt.Errorf("could not find joystick by instance ID")
}

/*
Starts polling for events within a go-routine every 60ms
Can be cancelled using the context
Returns a channel to listen to events
*/
func (mgr *SDLMgr) StartPolling(ctx context.Context) (chan SDL_Event, context.CancelFunc) {
	ctx_with_cancel, cancel := context.WithCancel(ctx)
	event_channel := make(chan SDL_Event, SDL_BUFFER_SIZE)

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		defer binsdl.Load().Unload()
		mgr.PanicInit()

		for {
			/* stop if context has been cancelled */
			if ctx_with_cancel.Err() != nil {
				return
			}

			var event sdl.Event
			if sdl.PollEvent(&event) {
				switch event.Type {
				case sdl.EVENT_JOYSTICK_ADDED:
					e := event.JoyDeviceEvent()
					added_event := &SDL_JoyDeviceAddedEvent{
						Timestamp: e.Timestamp,
						Which:     e.Which,
					}

					joystick, err := mgr.joyDeviceAdded(added_event)
					if err != nil {
						logger.Logger.Error("[SDLMgr] could not open joydevice", "error", err)
						continue
					}
					logger.Logger.Info(
						"[SDLMgr] registered joy device",
						"name", joystick.Name(),
						"device_id", joystick.DeviceID(),
						"product_version", joystick.version(),
						"serial", joystick.serial(),
						"path", joystick.path(),
					)
					chan_utils.SendTimeout[SDL_Event](event_channel, time.Second, added_event)
				case sdl.EVENT_JOYSTICK_REMOVED:
					e := event.JoyDeviceEvent()
					removed_event := &SDL_JoyDeviceRemovedEvent{
						Timestamp: e.Timestamp,
						Which:     e.Which,
					}
					mgr.joyDeviceRemoved(removed_event)
					chan_utils.SendTimeout[SDL_Event](event_channel, time.Second, removed_event)
				case sdl.EVENT_JOYSTICK_BUTTON_DOWN:
					e := event.JoyButtonEvent()
					chan_utils.SendTimeout[SDL_Event](event_channel, time.Second, &SDL_JoyButtonDownEvent{
						Timestamp: e.Timestamp,
						Which:     e.Which,
						Button:    e.Button,
						Down:      e.Down,
					})
				case sdl.EVENT_JOYSTICK_BUTTON_UP:
					e := event.JoyButtonEvent()
					chan_utils.SendTimeout[SDL_Event](event_channel, time.Second, &SDL_JoyButtonUpEvent{
						Timestamp: e.Timestamp,
						Which:     e.Which,
						Button:    e.Button,
						Down:      e.Down,
					})
				case sdl.EVENT_JOYSTICK_HAT_MOTION:
					e := event.JoyHatEvent()
					chan_utils.SendTimeout[SDL_Event](event_channel, time.Second, &SDL_JoyHatEvent{
						Timestamp: e.Timestamp,
						Which:     e.Which,
						Hat:       e.Hat,
						Value:     e.Value,
					})
				case sdl.EVENT_JOYSTICK_AXIS_MOTION:
					e := event.JoyAxisEvent()
					chan_utils.SendTimeout[SDL_Event](event_channel, time.Second, &SDL_JoyAxisEvent{
						Timestamp: e.Timestamp,
						Which:     e.Which,
						Axis:      e.Axis,
						Value:     e.Value,
					})
				case sdl.EVENT_QUIT:
					e := event.QuitEvent()
					chan_utils.SendTimeout[SDL_Event](event_channel, time.Second, &SDL_QuitEvent{
						Timestamp: e.Timestamp,
					})
				}
			}
		}
	}()
	return event_channel, cancel
}

func (joystick *SDLMgr_Joystick) version() uint16 {
	product_version := joystick.InternalJoystick.ProductVersion()
	if product_version == 0 && joystick.HIDDevice != nil {
		product_version = joystick.HIDDevice.Version()
	}
	return product_version
}

func (joystick *SDLMgr_Joystick) serial() string {
	device_serial := joystick.InternalJoystick.Serial()
	if device_serial == "" && joystick.HIDDevice != nil {
		device_serial = joystick.HIDDevice.SerialNumber()
	}
	return device_serial
}

func (joystick *SDLMgr_Joystick) path() string {
	device_path, _ := joystick.InternalJoystick.Path()
	return device_path
}

func (joystick *SDLMgr_Joystick) UniqueID() string {
	product_version := joystick.version()
	device_serial := joystick.serial()
	unique_id := fmt.Sprintf("usb_id=%s,version=%d,serial=%s", joystick.DeviceID(), product_version, device_serial)

	/*
		add device path or instance ID if serial wasn't available; from a session perspective this is
		the most unique ID we have even though it may not be very stable across sessions
	*/
	if device_serial == "" {
		device_path := joystick.path()
		if device_path != "" {
			unique_id = fmt.Sprintf("%s,device_path=%s", unique_id, device_path)
		} else {
			unique_id = fmt.Sprintf("%s,instance_id=%d", unique_id, joystick.InstanceID)
		}
	}

	hash := sha1.Sum([]byte(unique_id))

	return fmt.Sprintf("%x", hash)
}

func (joystick *SDLMgr_Joystick) DeviceID() string {
	return fmt.Sprintf("%04X:%04X", joystick.vendorID, joystick.productID)
}

func (joystick *SDLMgr_Joystick) Name() string {
	return joystick.name
}
