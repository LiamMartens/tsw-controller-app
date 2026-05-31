package main

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
	"tsw_controller_app/action_sequencer"
	"tsw_controller_app/cabdebugger"
	"tsw_controller_app/config"
	"tsw_controller_app/config_loader"
	"tsw_controller_app/controller_mgr"
	"tsw_controller_app/logger"
	"tsw_controller_app/profile_runner"
	"tsw_controller_app/tswapi"
	"tsw_controller_app/tswconnector"

	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) startupInitialize() {
	var connector tswconnector.TSWConnector
	var tsw_api *tswapi.TSWAPI
	switch a.config.Mode {
	case AppConfig_Mode_Default:
		connector = tswconnector.NewSocketConnection(a.ctx, VERSION)
		tsw_api = tswapi.NewTSWAPI(tswapi.TSWAPIConfig{
			BaseURL: "http://localhost:31270",
		})
	case AppConfig_Mode_Proxy:
		connector = tswconnector.NewSocketProxyConnection(a.ctx, a.config.ProxySettings.Addr)
		tsw_api = tswapi.NewTSWAPI(tswapi.TSWAPIConfig{
			BaseURL: fmt.Sprintf("http://%s:31270", a.config.ProxySettings.Addr),
		})
	}

	sdl_controller_manager := controller_mgr.NewSDLControllerManager(a.sdl_manager)
	virtual_controller_manager := controller_mgr.NewVirtualControllerManager(connector)
	action_sequencer := action_sequencer.New(connector)

	cab_debugger := cabdebugger.NewCabDebugger(tsw_api, connector, cabdebugger.CabDebugger_Config{})
	api_controller := profile_runner.NewAPIController(tsw_api)
	direct_controller := profile_runner.NewDirectController(connector)
	sync_controller := profile_runner.NewSyncController(connector)
	profile_runner := profile_runner.New(
		action_sequencer,
		sdl_controller_manager,
		virtual_controller_manager,
		direct_controller,
		sync_controller,
		api_controller,
		cab_debugger,
	)

	a.sdl_controller_manager = sdl_controller_manager
	a.virtual_controller_manager = virtual_controller_manager
	a.action_sequencer = action_sequencer
	a.connector = connector
	a.tswapi = tsw_api
	a.cab_debugger = cab_debugger
	a.direct_controller = direct_controller
	a.sync_controller = sync_controller
	a.api_controller = api_controller
	a.profile_runner = profile_runner
}

func (a *App) startupLoad() {
	a.LoadConfiguration()

	if a.program_config.TSWAPIKeyLocation != "" {
		a.tswapi.LoadAPIKey(a.program_config.TSWAPIKeyLocation)
		a.cab_debugger.UpdateConfig(cabdebugger.CabDebugger_Config{
			TSWAPISubscriptionIDStart: a.program_config.TSWAPISubscriptionIDStart,
		})
	}

	if a.program_config.PreferredControlMode == config.PreferredControlMode_DirectControl ||
		a.program_config.PreferredControlMode == config.PreferredControlMode_SyncControl ||
		a.program_config.PreferredControlMode == config.PreferredControlMode_ApiControl {
		a.profile_runner.Settings.SetPreferredControlMode(a.program_config.PreferredControlMode)
	}

	if a.program_config.AlwaysOnTop {
		runtime.WindowSetAlwaysOnTop(a.ctx, true)
	}
}

func (a *App) startupRun() {
	go func() {
		channel, unsubscribe := logger.Logger.Listen()
		defer unsubscribe()
		for {
			select {
			case <-a.ctx.Done():
				return
			case msg := <-channel:
				switch msg.LogLevel {
				case "debug":
					runtime.EventsEmit(a.ctx, AppEventType_Log_Debug, msg.Message)
				case "info":
					runtime.EventsEmit(a.ctx, AppEventType_Log_Info, msg.Message)
				case "error":
					runtime.EventsEmit(a.ctx, AppEventType_Log_Error, msg.Message)
				}
			}
		}
	}()

	go func() {
		for {
			if err := a.connector.Start(); err != nil {
				logger.Logger.Error("[app] could not start direct control server", "error", err)
			}

			/* if app context has been canceled -> return */
			if a.ctx.Err() != nil {
				return
			}

			/* otherwise wait and restart connector */
			time.Sleep(3 * time.Second)
		}
	}()

	go func() {
		a.cab_debugger.Start(a.ctx)
	}()

	go func() {
		cancel := a.sdl_controller_manager.Attach(a.ctx)
		defer cancel()
		<-a.ctx.Done()
	}()

	go func() {
		cancel := a.virtual_controller_manager.Attach(a.ctx)
		defer cancel()
		<-a.ctx.Done()
	}()

	go func() {
		cancel := a.profile_runner.Run(a.ctx)
		defer cancel()
		<-a.ctx.Done()
	}()

	go func() {
		cancel := a.action_sequencer.Run(a.ctx)
		defer cancel()
		<-a.ctx.Done()
	}()

	go func() {
		cancel := a.direct_controller.Run(a.ctx)
		defer cancel()
		<-a.ctx.Done()
	}()

	go func() {
		cancel := a.api_controller.Run(a.ctx)
		defer cancel()
		<-a.ctx.Done()
	}()

	go func() {
		cancel := a.sync_controller.Run(a.ctx)
		defer cancel()

		<-a.ctx.Done()
	}()

	go func() {
		sdl_channel, sdl_unsubsribe := a.sdl_controller_manager.SubscribeDevicesUpdated()
		virtual_channel, virtual_unsubscribe := a.virtual_controller_manager.SubscribeDevicesUpdated()
		defer sdl_unsubsribe()
		defer virtual_unsubscribe()
		for {
			select {
			case <-a.ctx.Done():
				return
			case <-sdl_channel:
				runtime.EventsEmit(a.ctx, AppEventType_JoyDevicesUpdated)
			case <-virtual_channel:
				runtime.EventsEmit(a.ctx, AppEventType_JoyDevicesUpdated)
			}
		}
	}()

	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			logger.Logger.Error("unable to watch files for auto-reloading", "error", err)
			return
		}
		defer watcher.Close()

		subdirectories := []string{config_loader.DIR_PROFILES_NAME, config_loader.DIR_CALIBRATION_NAME, config_loader.DIR_SDL_MAPPINGS_NAME}
		for _, dir := range subdirectories {
			watcher.Add(filepath.Join(a.config.GlobalConfigDir, dir))
			watcher.Add(filepath.Join(a.config.LocalConfigDir, dir))
		}

		for {
			select {
			case <-a.ctx.Done():
				return
			case _, ok := <-watcher.Events:
				if !ok {
					return
				}
				logger.Logger.Debug("[App::watcher] Automatically reloading configuration")
				a.LoadConfiguration()
			case err, ok := <-watcher.Errors:
				logger.Logger.Error("[App::watcher] error received while watching config directories", "error", err)
				if !ok {
					return
				}
			}
		}
	}()
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.sdllib = binsdl.Load()
	a.sdl_manager.PanicInit()
	a.startupInitialize()
	a.startupLoad()
	a.startupRun()
}

func (a *App) shutdown(ctx context.Context) {
	a.sdl_manager.Quit()
	defer a.sdllib.Unload()
}
