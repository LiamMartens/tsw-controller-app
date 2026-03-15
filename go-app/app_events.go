package main

import (
	"context"
	"fmt"
	"tsw_controller_app/controller_mgr"
	"tsw_controller_app/logger"
	"tsw_controller_app/sdl_mgr"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) UnsubscribeRaw() {
	if a.raw_subscriber != nil {
		a.raw_subscriber.Cancel()
		a.raw_subscriber = nil
	}
}

func (a *App) SubscribeRaw(unique_id controller_mgr.DeviceUniqueID) error {
	if a.raw_subscriber != nil {
		logger.Logger.Error("already listening")
		return fmt.Errorf("already listening")
	}

	var joystick *sdl_mgr.SDLMgr_Joystick
	if j, has_unconfigured_joystick := a.sdl_controller_manager.UnconfiguredControllers.Get(unique_id); has_unconfigured_joystick {
		joystick = j.Joystick
	} else if j, has_configured_joystick := a.sdl_controller_manager.ConfiguredControllers.Get(unique_id); has_configured_joystick {
		joystick = j.Joystick
	}

	if joystick == nil {
		logger.Logger.Error("joystick not found")
		return fmt.Errorf("joystick not found")
	}

	childctx, cancel := context.WithCancel(a.ctx)
	sdl_channel, sdl_cancel := a.sdl_controller_manager.SubscribeRaw()
	raw_subscriber := AppRawSubscriber{
		Channel: sdl_channel,
		Cancel:  cancel,
	}
	go func() {
		defer sdl_cancel()
		for {
			select {
			case <-childctx.Done():
				return
			case e := <-sdl_channel:
				if e.Device().UniqueID == joystick.UniqueID() {
					raw_event := Interop_RawEvent{
						UniqueID:  joystick.UniqueID(),
						DeviceID:  joystick.DeviceID(),
						Timestamp: e.Timestamp(),
					}
					switch event := e.(type) {
					case *controller_mgr.ControllerManager_RawEvent_Axis:
						raw_event.Kind = sdl_mgr.SDLMgr_Control_Kind_Axis
						raw_event.Index = event.Axis()
						raw_event.Value = event.Value()
					case *controller_mgr.ControllerManager_RawEvent_Button:
						raw_event.Kind = sdl_mgr.SDLMgr_Control_Kind_Button
						raw_event.Index = event.Button()
						raw_event.Value = event.Value()
					case *controller_mgr.ControllerManager_RawEvent_Hat:
						raw_event.Kind = sdl_mgr.SDLMgr_Control_Kind_Hat
						raw_event.Index = event.Hat()
						raw_event.Value = event.Value()
					}
					go runtime.EventsEmit(a.ctx, AppEventType_RawEvent, raw_event)
				}
			}
		}
	}()
	a.raw_subscriber = &raw_subscriber

	return nil
}

func (a *App) UnsubscribeChangeEvent() {
	if a.change_event_subscriber != nil {
		a.change_event_subscriber.Cancel()
		a.change_event_subscriber = nil
	}
}

func (a *App) SubscribeChangeEvent() error {
	if a.change_event_subscriber != nil {
		logger.Logger.Error("already subscribed")
		return fmt.Errorf("already subscribed")
	}

	childctx, cancel := context.WithCancel(a.ctx)
	sdl_channel, sdl_cancel := a.sdl_controller_manager.SubscribeChangeEvent()
	virt_channel, virt_cancel := a.virtual_controller_manager.SubscribeChangeEvent()

	change_event_subscriber := AppChangeEventSubscriber{
		Channel: make(chan controller_mgr.ControllerManager_Control_ChangeEvent),
		Cancel:  cancel,
	}
	a.change_event_subscriber = &change_event_subscriber

	go func() {
		defer sdl_cancel()
		defer virt_cancel()
		for {
			select {
			case <-childctx.Done():
				return
			case event := <-sdl_channel:
				change_event := Interop_ChangeEvent{
					UniqueID:    event.Device.UniqueID,
					DeviceID:    event.Device.DeviceID,
					ControlName: event.ControlName,
					Value:       event.ControlState.NormalizedValues.Value,
				}
				go runtime.EventsEmit(a.ctx, AppEventType_ChangeEvent, change_event)
			case event := <-virt_channel:
				change_event := Interop_ChangeEvent{
					UniqueID:    event.Device.UniqueID,
					DeviceID:    event.Device.DeviceID,
					ControlName: event.ControlName,
					Value:       event.ControlState.NormalizedValues.Value,
				}
				go runtime.EventsEmit(a.ctx, AppEventType_ChangeEvent, change_event)
			}
		}
	}()

	return nil
}
