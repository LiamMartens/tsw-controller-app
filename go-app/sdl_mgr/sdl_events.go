package sdl_mgr

import "github.com/Zyko0/go-sdl3/sdl"

type SDL_Event interface {
	GetType() sdl.EventType
	GetTimestamp() uint64
}

/** SDL_JoyDeviceAddedEvent */
type SDL_JoyDeviceAddedEvent struct {
	Timestamp uint64
	Which     sdl.JoystickID
}

func (e *SDL_JoyDeviceAddedEvent) GetType() sdl.EventType {
	return sdl.EVENT_JOYSTICK_ADDED
}

func (e *SDL_JoyDeviceAddedEvent) GetTimestamp() uint64 {
	return e.Timestamp
}

var _ SDL_Event = &SDL_JoyDeviceAddedEvent{}

/** SDL_JoyDeviceRemovedEvent */
type SDL_JoyDeviceRemovedEvent struct {
	Timestamp uint64
	Which     sdl.JoystickID
	Button    uint8
	State     bool
}

func (e *SDL_JoyDeviceRemovedEvent) GetType() sdl.EventType {
	return sdl.EVENT_JOYSTICK_REMOVED
}

func (e *SDL_JoyDeviceRemovedEvent) GetTimestamp() uint64 {
	return e.Timestamp
}

var _ SDL_Event = &SDL_JoyDeviceRemovedEvent{}

/** SDL_JoyButtonDownEvent */
type SDL_JoyButtonDownEvent struct {
	Timestamp uint64
	Which     sdl.JoystickID
	Button    uint8
	Down      bool
}

func (e *SDL_JoyButtonDownEvent) GetType() sdl.EventType {
	return sdl.EVENT_PEN_BUTTON_DOWN
}

func (e *SDL_JoyButtonDownEvent) GetTimestamp() uint64 {
	return e.Timestamp
}

var _ SDL_Event = &SDL_JoyButtonDownEvent{}

/** SDL_JoyButtonUpEvent */
type SDL_JoyButtonUpEvent struct {
	Timestamp uint64
	Which     sdl.JoystickID
	Button    uint8
	Down      bool
}

func (e *SDL_JoyButtonUpEvent) GetType() sdl.EventType {
	return sdl.EVENT_PEN_BUTTON_UP
}

func (e *SDL_JoyButtonUpEvent) GetTimestamp() uint64 {
	return e.Timestamp
}

var _ SDL_Event = &SDL_JoyButtonUpEvent{}

/** SDL_JoyHatEvent */
type SDL_JoyHatEvent struct {
	Timestamp uint64
	Which     sdl.JoystickID
	Hat       uint8
	Value     uint8
}

func (e *SDL_JoyHatEvent) GetType() sdl.EventType {
	return sdl.EVENT_JOYSTICK_HAT_MOTION
}

func (e *SDL_JoyHatEvent) GetTimestamp() uint64 {
	return e.Timestamp
}

var _ SDL_Event = &SDL_JoyHatEvent{}

/** SDL_JoyAxisEvent */
type SDL_JoyAxisEvent struct {
	Timestamp uint64
	Which     sdl.JoystickID
	Axis      uint8
	Value     int16
}

func (e *SDL_JoyAxisEvent) GetType() sdl.EventType {
	return sdl.EVENT_JOYSTICK_AXIS_MOTION
}

func (e *SDL_JoyAxisEvent) GetTimestamp() uint64 {
	return e.Timestamp
}

var _ SDL_Event = &SDL_JoyAxisEvent{}

/** SDL_QuitEvent */
type SDL_QuitEvent struct {
	Timestamp uint64
}

func (e *SDL_QuitEvent) GetType() sdl.EventType {
	return sdl.EVENT_QUIT
}

func (e *SDL_QuitEvent) GetTimestamp() uint64 {
	return e.Timestamp
}

var _ SDL_Event = &SDL_QuitEvent{}
