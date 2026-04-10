package sdl_mgr

import (
	"context"
	"crypto/sha1"
	"fmt"
	"sync"
	"time"
	"tsw_controller_app/chan_utils"
	"tsw_controller_app/logger"

	"github.com/veandco/go-sdl2/sdl"
)

/* the SDL control kind like Button, Hat, Axis */
type SDLMgr_Control_Kind = string
type SDLMgr_Guid_Str = string

const SDL_BUFFER_SIZE = 32
const SDL_RATE = 16

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

	InternalJoystick *sdl.Joystick
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

func (mgr *SDLMgr) joyDeviceAdded(event *sdl.JoyDeviceAddedEvent) (*SDLMgr_Joystick, error) {
	mgr.joydevices_mutex.Lock()
	defer mgr.joydevices_mutex.Unlock()

	joy_index := int(event.Which)
	instance_id := sdl.JoystickGetDeviceInstanceID(joy_index)
	name := sdl.JoystickNameForIndex(joy_index)
	usb_vendor := sdl.JoystickGetDeviceVendor(joy_index)
	usb_product := sdl.JoystickGetDeviceProduct(joy_index)
	joystick := SDLMgr_Joystick{
		InstanceID:       instance_id,
		name:             name,
		vendorID:         usb_vendor,
		productID:        usb_product,
		InternalJoystick: nil,
	}

	logger.Logger.Debug("[SDLMgr_Joystick::Open] opening joystick", "joystick", joystick.DeviceID(), "name", joystick.Name)
	joystick.InternalJoystick = sdl.JoystickOpen(joy_index)
	if joystick.InternalJoystick == nil {
		return nil, fmt.Errorf("could not open joystick for use: %w", sdl.GetError())
	}

	mgr.joydevices[instance_id] = &joystick
	return &joystick, nil
}

func (mgr *SDLMgr) joyDeviceRemoved(event *sdl.JoyDeviceRemovedEvent) {
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
		sdl.JoystickEventState(1)
		if err := sdl.Init(sdl.INIT_GAMECONTROLLER | sdl.INIT_JOYSTICK | sdl.INIT_EVENTS); err != nil {
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
func (mgr *SDLMgr) StartPolling(ctx context.Context) (chan sdl.Event, context.CancelFunc) {
	ctx_with_cancel, cancel := context.WithCancel(ctx)
	event_channel := make(chan sdl.Event, SDL_BUFFER_SIZE)

	go func() {
		for {
			/* stop if context has been cancelled */
			if ctx_with_cancel.Err() != nil {
				return
			}

			if event := sdl.WaitEventTimeout(SDL_RATE); event != nil {
				switch e := event.(type) {
				case *sdl.JoyDeviceAddedEvent:
					if joystick, err := mgr.joyDeviceAdded(&sdl.JoyDeviceAddedEvent{
						Type:      e.Type,
						Timestamp: e.Timestamp,
						Which:     e.Which,
					}); err == nil {
						chan_utils.SendTimeout[sdl.Event](event_channel, time.Second, &sdl.JoyDeviceAddedEvent{
							Type:      e.Type,
							Timestamp: e.Timestamp,
							/* switching to instance ID once we're out of the SDL internals */
							Which: joystick.InstanceID,
						})
					}
				case *sdl.JoyDeviceRemovedEvent:
					removed_event := &sdl.JoyDeviceRemovedEvent{
						Type:      e.Type,
						Timestamp: e.Timestamp,
						Which:     e.Which,
					}
					mgr.joyDeviceRemoved(removed_event)
					chan_utils.SendTimeout[sdl.Event](event_channel, time.Second, removed_event)
				case *sdl.JoyButtonEvent:
					chan_utils.SendTimeout[sdl.Event](event_channel, time.Second, &sdl.JoyButtonEvent{
						Type:      e.Type,
						Timestamp: e.Timestamp,
						Which:     e.Which,
						Button:    e.Button,
						State:     e.State,
					})
				case *sdl.JoyHatEvent:
					chan_utils.SendTimeout[sdl.Event](event_channel, time.Second, &sdl.JoyHatEvent{
						Type:      e.Type,
						Timestamp: e.Timestamp,
						Which:     e.Which,
						Hat:       e.Hat,
						Value:     e.Value,
					})
				case *sdl.JoyAxisEvent:
					chan_utils.SendTimeout[sdl.Event](event_channel, time.Second, &sdl.JoyAxisEvent{
						Type:      e.Type,
						Timestamp: e.Timestamp,
						Which:     e.Which,
						Axis:      e.Axis,
						Value:     e.Value,
					})
				}
			}
		}
	}()
	return event_channel, cancel
}

func (joystick *SDLMgr_Joystick) UniqueID() string {
	unique_id := fmt.Sprintf("usb_id=%s,instance_id=%s", joystick.DeviceID(), string(joystick.InstanceID))
	hash := sha1.Sum([]byte(unique_id))
	return fmt.Sprintf("%x", hash)
}

func (joystick *SDLMgr_Joystick) DeviceID() string {
	return fmt.Sprintf("%04X:%04X", joystick.vendorID, joystick.productID)
}

func (joystick *SDLMgr_Joystick) Name() string {
	return joystick.name
}
