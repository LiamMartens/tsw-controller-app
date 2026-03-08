package main

import (
	"context"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	go_runtime "runtime"
	"sort"
	"strings"
	"time"
	"tsw_controller_app/action_sequencer"
	"tsw_controller_app/cabdebugger"
	"tsw_controller_app/config"
	"tsw_controller_app/config_loader"
	"tsw_controller_app/controller_mgr"
	"tsw_controller_app/logger"
	"tsw_controller_app/profile_runner"
	"tsw_controller_app/sdl_mgr"
	"tsw_controller_app/string_utils"
	"tsw_controller_app/tswapi"
	"tsw_controller_app/tswconnector"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed embed/mod_assets/*
var embed_mod_assets_fs embed.FS

//go:embed embed/tsc_mod_assets/*
var embed_tsc_mod_assets_fs embed.FS

//go:embed embed/config/*
var embed_config_fs embed.FS

type AppEventType = string

const (
	AppEventType_JoyDevicesUpdated AppEventType = "joydevices_updated"
	AppEventType_ProfilesUpdated   AppEventType = "profiles_updated"
	AppEventType_RawEvent          AppEventType = "rawevent"
	AppEventType_ChangeEvent       AppEventType = "changeevent"
	AppEventType_Log_Debug         AppEventType = "log/debug"
	AppEventType_Log_Info          AppEventType = "log/info"
	AppEventType_Log_Error         AppEventType = "log/error"
)

type AppConfig_Mode = string

const (
	AppConfig_Mode_Default AppConfig_Mode = "default"
	AppConfig_Mode_Proxy   AppConfig_Mode = "proxy"
)

type ModAssets_Manifest_Entry_ActionType = string

const (
	ModAssets_Manifest_Entry_ActionType_Copy   ModAssets_Manifest_Entry_ActionType = "copy"
	ModAssets_Manifest_Entry_ActionType_Delete ModAssets_Manifest_Entry_ActionType = "delete"
)

type ModAssets_Manifest_Entry struct {
	Path   string `json:"path"`
	Action string `json:"action" validate:"required,oneof=copy delete"`
}

type ModAssets_Manifest struct {
	Manifest []ModAssets_Manifest_Entry `json:"manifest"`
}

type Remote_SharedProfilesIndex_Profile_Author struct {
	Name string  `json:"name,omitempty"`
	Url  *string `json:"url,omitempty"`
}

type Remote_SharedProfilesIndex_Profile struct {
	File                string                                     `json:"file"`
	Name                string                                     `json:"name"`
	UsbID               string                                     `json:"usb_id"`
	AutoSelect          *bool                                      `json:"auto_select,omitempty"`
	Apps                *[]string                                  `json:"apps,omitempty"`
	ContainsCalibration *bool                                      `json:"contains_calibration,omitempty"`
	Author              *Remote_SharedProfilesIndex_Profile_Author `json:"author,omitempty"`
}

type Remote_SharedProfilesIndex struct {
	Profiles []Remote_SharedProfilesIndex_Profile `json:"profiles"`
}

type AppRawSubscriber struct {
	Channel chan controller_mgr.IControllerManager_RawEvent
	Cancel  func()
}

type AppChangeEventSubscriber struct {
	Channel chan controller_mgr.ControllerManager_Control_ChangeEvent
	Cancel  func()
}

type AppConfig_ProxySettings struct {
	Addr string
}

type AppConfig struct {
	GlobalConfigDir string
	LocalConfigDir  string
	Mode            AppConfig_Mode
	ProxySettings   *AppConfig_ProxySettings
}

type App struct {
	ctx                        context.Context
	config                     AppConfig
	program_config             *config.Config_ProgramConfig
	config_loader              *config_loader.ConfigLoader
	sdl_manager                *sdl_mgr.SDLMgr
	sdl_controller_manager     *controller_mgr.SDLControllerManager
	virtual_controller_manager *controller_mgr.VirtualControllerManager
	action_sequencer           *action_sequencer.ActionSequencer
	connector                  tswconnector.TSWConnector
	tswapi                     *tswapi.TSWAPI
	cab_debugger               *cabdebugger.CabDebugger
	direct_controller          *profile_runner.DirectController
	sync_controller            *profile_runner.SyncController
	api_controller             *profile_runner.ApiController
	profile_runner             *profile_runner.ProfileRunner

	raw_subscriber          *AppRawSubscriber
	change_event_subscriber *AppChangeEventSubscriber
}

/* these are just wails type stubs */
func (a *App) TypestubGetSelectedProfile() *Interop_SelectedProfileInfo {
	return nil
}

func (a *App) TypestubGetRawEvent() *Interop_RawEvent {
	return nil
}

func (a *App) TypestubGetChangeEvent() *Interop_ChangeEvent {
	return nil
}

func NewApp(
	appconfig AppConfig,
) *App {
	sdl_manager := sdl_mgr.New()
	sdl_manager.PanicInit()

	program_config := config.LoadProgramConfigFromFile(filepath.Join(appconfig.GlobalConfigDir, "program.json"))
	if program_config.TSWAPIKeyLocation == "" {
		program_config.TSWAPIKeyLocation = program_config.AutoDetectTSWAPIKeyLocation()
	}

	return &App{
		config:         appconfig,
		program_config: program_config,
		config_loader:  config_loader.New(),
		sdl_manager:    sdl_manager,
	}
}

func (a *App) startupInitialize() {
	var connector tswconnector.TSWConnector
	var tsw_api *tswapi.TSWAPI
	switch a.config.Mode {
	case AppConfig_Mode_Default:
		connector = tswconnector.NewSocketConnection(a.ctx)
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
		a.connector.Start()
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
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.startupInitialize()
	a.startupLoad()
	a.startupRun()
}

func (a *App) shutdown(ctx context.Context) {
}

func (a *App) GetVersion() string {
	return VERSION
}

func (a *App) GetLastInstalledModVersion() string {
	return a.program_config.LastInstalledModVersion
}

func (a *App) SetLastInstalledModVersion(version string) {
	a.program_config.LastInstalledModVersion = version
	a.program_config.Save(filepath.Join(a.config.GlobalConfigDir, "program.json"))
}

func (a *App) GetTSWAPIKeyLocation() string {
	return a.program_config.TSWAPIKeyLocation
}

func (a *App) SetTSWAPIKeyLocation(location string) {
	a.program_config.TSWAPIKeyLocation = location
	a.tswapi.LoadAPIKey(location)
	a.program_config.Save(filepath.Join(a.config.GlobalConfigDir, "program.json"))
}

func (a *App) GetPreferredControlMode() string {
	return a.program_config.PreferredControlMode
}

func (a *App) SetPreferredControlMode(mode config.PreferredControlMode) {
	a.program_config.PreferredControlMode = mode
	a.profile_runner.Settings.SetPreferredControlMode(mode)
	a.program_config.Save(filepath.Join(a.config.GlobalConfigDir, "program.json"))
}

func (a *App) GetAlwaysOnTop() bool {
	return a.program_config.AlwaysOnTop
}

func (a *App) SetAlwaysOnTop(enabled bool) {
	a.program_config.AlwaysOnTop = enabled
	runtime.WindowSetAlwaysOnTop(a.ctx, enabled)
	a.program_config.Save(filepath.Join(a.config.GlobalConfigDir, "program.json"))
}

func (a *App) GetTheme() string {
	return a.program_config.Theme
}

func (a *App) SetTheme(theme string) {
	a.program_config.Theme = theme
	a.program_config.Save(filepath.Join(a.config.GlobalConfigDir, "program.json"))
}

func (a *App) GetDeviceIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func (a *App) LoadConfiguration() {
	/* load config from relative config directory */
	embed_config_fs, _ := fs.Sub(embed_config_fs, "embed/config")

	type loadLocation struct {
		fs       fs.FS
		path     string
		embedded bool
	}

	load_locations := []loadLocation{
		{fs: embed_config_fs, path: "builtin:", embedded: true},
		{fs: os.DirFS(a.config.GlobalConfigDir), path: a.config.GlobalConfigDir},
		{fs: os.DirFS(a.config.LocalConfigDir), path: a.config.LocalConfigDir},
	}

	a.profile_runner.Profiles.Clear()
	for _, loc := range load_locations {
		sdl_mappings, calibrations, profiles, errors := a.config_loader.FromFS(loc.fs, config_loader.ConfigLoader_FromFS_Options{
			Path:     loc.path,
			Embedded: loc.embedded,
		})

		for _, err := range errors {
			logger.Logger.Error("[App] encountered error while reading configuration files", "error", err)
		}

		for _, sdl_mapping := range sdl_mappings {
			var calibration *config.Config_Controller_Calibration
			for _, c := range calibrations {
				if c.UsbID == sdl_mapping.UsbID {
					calibration = &c
					break
				}
			}
			if calibration != nil {
				logger.Logger.Info("[App] registering SDL map and calibration for controller", "name", sdl_mapping.Name, "usb_id", sdl_mapping.UsbID)
				a.sdl_controller_manager.RegisterConfig(sdl_mapping, *calibration)
			}
		}

		for _, profile := range profiles {
			logger.Logger.Info("[App] registering profile", "profile", profile.Id(), profile.Name)
			a.profile_runner.RegisterProfile(profile)
		}
	}

	a.profile_runner.Resolve()
	runtime.EventsEmit(a.ctx, AppEventType_ProfilesUpdated)
}

func (a *App) GetControllers() []Interop_GenericController {
	var controllers []Interop_GenericController = []Interop_GenericController{}
	a.sdl_controller_manager.ConfiguredControllers.ForEach(func(c controller_mgr.SDL_ControllerManager_ConfiguredController, _ controller_mgr.DeviceUniqueID) bool {
		controllers = append(controllers, Interop_GenericController{
			UniqueID:     c.Device().UniqueID(),
			DeviceID:     c.Device().DeviceID(),
			Name:         c.Name,
			IsConfigured: true,
			IsVirtual:    false,
		})
		return true
	})
	a.sdl_controller_manager.UnconfiguredControllers.ForEach(func(c controller_mgr.SDL_ControllerManager_UnconfiguredController, _ controller_mgr.DeviceUniqueID) bool {
		controllers = append(controllers, Interop_GenericController{
			UniqueID:     c.Joystick.UniqueID(),
			DeviceID:     c.Joystick.DeviceID(),
			Name:         c.Joystick.Name(),
			IsConfigured: false,
			IsVirtual:    false,
		})
		return true
	})
	a.virtual_controller_manager.Controllers().ForEach(func(c *controller_mgr.VirtualControllerManager_Controller, key controller_mgr.DeviceUniqueID) bool {
		controllers = append(controllers, Interop_GenericController{
			UniqueID:     c.Device().UniqueID(),
			DeviceID:     c.Device().DeviceID(),
			Name:         c.Device().Name(),
			IsConfigured: true,
			IsVirtual:    true,
		})
		return true
	})
	sort.Slice(controllers, func(i, j int) bool {
		return controllers[i].Name < controllers[j].Name
	})
	return controllers
}

func (a *App) GetProfiles() []Interop_Profile {
	var profiles []Interop_Profile = []Interop_Profile{}

	profile_name_to_ids_map := a.profile_runner.GetProfileNameToIdMap()
	a.profile_runner.Profiles.ForEach(func(profile config.Config_Controller_Profile, key string) bool {
		UsbID := ""
		if profile.Controller != nil && profile.Controller.UsbID != nil {
			UsbID = *profile.Controller.UsbID
		}

		warnings := []string{}
		if profile.Extends != nil && len(*profile.Extends) > 0 {
			extend_from, has_extend_from_ids := profile_name_to_ids_map[*profile.Extends]
			if has_extend_from_ids && len(extend_from) > 1 {
				warnings = append(warnings, fmt.Sprintf("Could not resolve profile, found multiple profiles by name (%s) to resolve from", *profile.Extends))
			} else if !has_extend_from_ids || len(extend_from) == 0 {
				warnings = append(warnings, fmt.Sprintf("Could not find profile name to extend from (%s)", *profile.Extends))
			}
			if *profile.Extends == profile.Name {
				warnings = append(warnings, "This profile extends from itself, which is not a valid use-case")
			}
		}

		profiles = append(profiles, Interop_Profile{
			Id:         profile.Id(),
			Name:       profile.Name,
			DeviceID:   UsbID,
			AutoSelect: profile.AutoSelect,
			Apps:       profile.Apps,
			Metadata: Interop_Profile_Metadata{
				Path:       profile.Metadata.Path,
				IsEmbedded: profile.Metadata.IsEmbedded,
				UpdatedAt:  profile.Metadata.UpdatedAt.Format(time.RFC3339),
				Warnings:   warnings,
			},
		})
		return true
	})
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})
	return profiles
}

func (a *App) GetSelectedProfiles() map[controller_mgr.DeviceUniqueID]Interop_SelectedProfileInfo {
	selected_profiles := map[controller_mgr.DeviceUniqueID]Interop_SelectedProfileInfo{}
	a.profile_runner.Settings.GetSelectedProfiles().ForEach(func(value profile_runner.ProfileRunnerSettings_SelectedProfile, unique_id controller_mgr.DeviceUniqueID) bool {
		selected_profiles[unique_id] = Interop_SelectedProfileInfo{
			Id:   value.Profile.Id(),
			Name: value.Profile.Name,
		}
		return true
	})
	return selected_profiles
}

func (a *App) GetControllerConfiguration(unique_id controller_mgr.DeviceUniqueID) *Interop_ControllerConfiguration {
	if controller, has_controller := a.sdl_controller_manager.ConfiguredControllers.Get(unique_id); has_controller {
		// /* when configured the SDL map and calibration always exist */
		sdl_mapping, _ := controller.Manager.Config().SDLMappingsByDeviceID.Get(controller.Joystick.DeviceID())
		interop_calibration := Interop_ControllerCalibration{
			Name:     sdl_mapping.Name,
			DeviceID: sdl_mapping.UsbID,
			Controls: []Interop_ControllerCalibration_Control{},
		}
		controller.Controls().ForEach(func(c controller_mgr.IControllerManager_Controller_Control, key string) bool {
			if control, ok := c.(*controller_mgr.SDL_ControllerManager_Controller_JoyControl); ok {
				sdl_mapping := control.SDLMapping()
				calibration_data := control.Calibration()
				calibration := Interop_ControllerCalibration_Control{
					Kind:        sdl_mapping.Kind,
					Index:       sdl_mapping.Index,
					Name:        control.Name(),
					Min:         calibration_data.Min,
					Max:         calibration_data.Max,
					Idle:        0,
					Deadzone:    0,
					Invert:      false,
					EasingCurve: []float64{0.0, 0.0, 1.0, 1.0},
				}
				if calibration_data.Idle != nil {
					calibration.Idle = *calibration_data.Idle
				}
				if calibration_data.Deadzone != nil {
					calibration.Deadzone = *calibration_data.Deadzone
				}
				if calibration_data.Invert != nil {
					calibration.Invert = *calibration_data.Invert
				}
				if calibration_data.EasingCurve != nil {
					calibration.EasingCurve = *calibration_data.EasingCurve
				}
				interop_calibration.Controls = append(interop_calibration.Controls, calibration)
			}
			return true
		})
		return &Interop_ControllerConfiguration{
			SDLMapping:  sdl_mapping,
			Calibration: interop_calibration,
		}
	}
	return nil
}

func (a *App) GetCabControlState() (Interop_Cab_ControlState, error) {
	control_state := Interop_Cab_ControlState{
		Name:     a.cab_debugger.State.DrivableActorName,
		Controls: []Interop_Cab_ControlState_Control{},
	}

	a.cab_debugger.State.Controls.ForEach(func(control cabdebugger.CabDebugger_ControlState_Control, key cabdebugger.PropertyName) bool {
		control_state.Controls = append(control_state.Controls, Interop_Cab_ControlState_Control{
			Identifier:             control.Identifier,
			PropertyName:           control.PropertyName,
			CurrentValue:           control.CurrentValue,
			CurrentNormalizedValue: control.CurrentNormalizedValue,
		})
		return true
	})

	return control_state, nil
}

func (a *App) ResetCabControlState() {
	a.cab_debugger.Clear()
}

// https://github.com/LiamMartens/tsw-controller-app/releases/download/v0.2.6/beta.package.zip
func (a *App) GetLatestReleaseVersion() string {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://raw.githubusercontent.com/LiamMartens/tsw-controller-app/refs/heads/main/RELEASE_VERSION")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	return strings.Split(string(body), "\n")[0]
}

func (a *App) GetSharedProfiles() []Interop_SharedProfile {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://raw.githubusercontent.com/LiamMartens/tsw-controller-app/refs/heads/main/shared-profiles/index.json")
	if err != nil {
		return []Interop_SharedProfile{}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []Interop_SharedProfile{}
	}

	var c Remote_SharedProfilesIndex
	json.Unmarshal(body, &c)

	profiles := []Interop_SharedProfile{}
	for _, profile := range c.Profiles {
		var author *Interop_SharedProfile_Author = nil
		if profile.Author != nil {
			author = &Interop_SharedProfile_Author{
				Name: profile.Author.Name,
				Url:  profile.Author.Url,
			}
		}
		profiles = append(profiles, Interop_SharedProfile{
			Name:                profile.Name,
			DeviceID:            profile.UsbID,
			Url:                 fmt.Sprintf("https://raw.githubusercontent.com/LiamMartens/tsw-controller-app/refs/heads/main/shared-profiles/%s", profile.File),
			AutoSelect:          profile.AutoSelect,
			Apps:                profile.Apps,
			ContainsCalibration: profile.ContainsCalibration,
			Author:              author,
		})
	}

	return profiles
}

func (a *App) SelectProfile(unique_id controller_mgr.DeviceUniqueID, id string) error {
	if err := a.profile_runner.SetProfile(unique_id, id); err != nil {
		logger.Logger.Error("failed to select profile by ID", "id", id, "error", err)
		return err
	}
	return nil
}

func (a *App) ClearProfile(unique_id controller_mgr.DeviceUniqueID) {
	a.profile_runner.ClearProfile(unique_id)
}

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

func (a *App) SaveProfileForSharing(id string) error {
	if profile, has_profile := a.profile_runner.Profiles.Get(id); has_profile {
		profile_for_sharing := config.Config_Controller_Profile{
			Name:                 profile.Name,
			AutoSelect:           profile.AutoSelect,
			RailClassInformation: profile.RailClassInformation,
			Controller:           profile.Controller,
			Controls:             profile.Controls,
		}

		profile_for_sharing_filepath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
			Title:           "Select save location for profile",
			DefaultFilename: fmt.Sprintf("%s.tswprofile", string_utils.Sluggify(profile_for_sharing.Name)),
		})
		if err != nil {
			return err
		}

		profile_for_sharing_file, err := os.OpenFile(profile_for_sharing_filepath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer profile_for_sharing_file.Close()

		encoder_sdl_mapping_file := json.NewEncoder(profile_for_sharing_file)
		encoder_sdl_mapping_file.SetIndent("", "  ")
		if err := encoder_sdl_mapping_file.Encode(profile_for_sharing); err != nil {
			return err
		}

		return nil
	} else {
		return fmt.Errorf("could not find profile")
	}
}

func (a *App) SaveProfileForSharingWithControllerInformation(id string, unique_id controller_mgr.DeviceUniqueID) error {
	if profile, has_profile := a.profile_runner.Profiles.Get(id); has_profile {
		controller, has_controller := a.sdl_controller_manager.ConfiguredControllers.Get(unique_id)
		if !has_controller {
			return fmt.Errorf("could not find controller")
		}

		usb_id := controller.Joystick.DeviceID()
		profile_for_sharing := config.Config_Controller_Profile{
			/*
				this copy omits extends and the internal metadata since it's not appropriate for sharing,
			*/
			Name:                 profile.Name,
			AutoSelect:           profile.AutoSelect,
			RailClassInformation: profile.RailClassInformation,
			Controller:           profile.Controller,
			Controls:             profile.Controls,
		}
		if profile_for_sharing.Controller == nil {
			profile_for_sharing.Controller = &config.Config_Controller_Profile_Controller{
				UsbID:   &usb_id,
				Mapping: nil,
			}
		}

		if profile_for_sharing.Controller.Mapping == nil {
			mapping := config.Config_Controller_SDLMap{
				Name:  fmt.Sprintf("%s - %s", controller.Name, profile_for_sharing.Name),
				UsbID: usb_id,
				Data:  []config.Config_Controller_SDLMap_Control{},
			}
			controller.Controls().ForEach(func(c controller_mgr.IControllerManager_Controller_Control, key string) bool {
				if control, ok := c.(*controller_mgr.SDL_ControllerManager_Controller_JoyControl); ok {
					mapping.Data = append(mapping.Data, control.SDLMapping())
				}
				return true
			})
			profile_for_sharing.Controller.Mapping = &mapping
		}

		profile_for_sharing_filepath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
			Title:           "Select save location for profile",
			DefaultFilename: fmt.Sprintf("%s.tswprofile", string_utils.Sluggify(profile_for_sharing.Name)),
		})
		if err != nil {
			return err
		}

		profile_for_sharing_file, err := os.OpenFile(profile_for_sharing_filepath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer profile_for_sharing_file.Close()

		encoder_sdl_mapping_file := json.NewEncoder(profile_for_sharing_file)
		encoder_sdl_mapping_file.SetIndent("", "  ")
		if err := encoder_sdl_mapping_file.Encode(profile_for_sharing); err != nil {
			return err
		}

		return nil
	} else {
		return fmt.Errorf("could not find profile")
	}
}

func (a *App) OpenNewProfileBuilder() {
	empty_profile := config.Config_Controller_Profile{
		Name:     "My new profile",
		Controls: []config.Config_Controller_Profile_Control{},
	}
	profile_json, _ := json.Marshal(empty_profile)
	encoded := base64.StdEncoding.EncodeToString(profile_json)
	runtime.BrowserOpenURL(a.ctx, fmt.Sprintf("https://tsw-controller-app.vercel.app/profile-builder?profile=%s", encoded))
}

func (a *App) OpenNewProfileBuilderForDeviceID(deviceid string) {
	empty_profile := config.Config_Controller_Profile{
		Name: "My new profile",
		Controller: &config.Config_Controller_Profile_Controller{
			UsbID: &deviceid,
		},
		Controls: []config.Config_Controller_Profile_Control{},
	}
	profile_json, _ := json.Marshal(empty_profile)
	encoded := base64.StdEncoding.EncodeToString(profile_json)
	runtime.BrowserOpenURL(a.ctx, fmt.Sprintf("https://tsw-controller-app.vercel.app/profile-builder?profile=%s", encoded))
}

func (a *App) OpenProfileBuilder(id string) {
	if profile, has_profile := a.profile_runner.Profiles.Get(id); has_profile {
		profile_json, _ := json.Marshal(profile)
		encoded := base64.StdEncoding.EncodeToString(profile_json)
		runtime.BrowserOpenURL(a.ctx, fmt.Sprintf("https://tsw-controller-app.vercel.app/profile-builder?profile=%s", encoded))
	}
}

func (a *App) DeleteProfile(id string) error {
	if profile, has_profile := a.profile_runner.Profiles.Get(id); has_profile {
		err := os.Remove(profile.Metadata.Path)
		if err != nil {
			return err
		}
		a.profile_runner.Profiles.Delete(id)
	}
	return nil
}

func (a *App) SaveCalibration(data Interop_ControllerCalibration) error {
	sdl_mapping := config.Config_Controller_SDLMap{
		Name:  data.Name,
		UsbID: data.DeviceID,
		Data:  []config.Config_Controller_SDLMap_Control{},
	}
	calibration := config.Config_Controller_Calibration{
		UsbID: data.DeviceID,
		Data:  []config.Config_Controller_CalibrationData{},
	}
	for _, control := range data.Controls {
		if control.Name != "" {
			sdl_mapping.Data = append(sdl_mapping.Data, config.Config_Controller_SDLMap_Control{
				Kind:  control.Kind,
				Index: control.Index,
				Name:  control.Name,
			})
			if control.Kind == sdl_mgr.SDLMgr_Control_Kind_Axis {
				calibration.Data = append(calibration.Data, config.Config_Controller_CalibrationData{
					Id:          control.Name,
					Min:         control.Min,
					Max:         control.Max,
					Idle:        &control.Idle,
					Deadzone:    &control.Deadzone,
					Invert:      &control.Invert,
					EasingCurve: &control.EasingCurve,
				})
			}
		}
	}

	sdl_mapping_filepath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:            "Select SDL mapping file save location",
		DefaultFilename:  fmt.Sprintf("%s.sdl.json", string_utils.Sluggify(data.Name)),
		DefaultDirectory: filepath.Join(a.config.GlobalConfigDir, config_loader.DIR_SDL_MAPPINGS_NAME),
	})
	if err != nil {
		return err
	}

	calibration_filepath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:            "Select calibration file save location",
		DefaultFilename:  fmt.Sprintf("%s.calibration.json", string_utils.Sluggify(data.Name)),
		DefaultDirectory: filepath.Join(a.config.GlobalConfigDir, config_loader.DIR_CALIBRATION_NAME),
	})
	if err != nil {
		return err
	}

	sdl_mapping_file, err := os.OpenFile(sdl_mapping_filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer sdl_mapping_file.Close()

	calibration_file, err := os.OpenFile(calibration_filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer calibration_file.Close()

	encoder_sdl_mapping_file := json.NewEncoder(sdl_mapping_file)
	encoder_sdl_mapping_file.SetIndent("", "  ")
	if err := encoder_sdl_mapping_file.Encode(sdl_mapping); err != nil {
		return err
	}

	encoder_calibration_file := json.NewEncoder(calibration_file)
	encoder_calibration_file.SetIndent("", "  ")
	if err := encoder_calibration_file.Encode(calibration); err != nil {
		return err
	}

	/* register config */
	a.sdl_controller_manager.RegisterConfig(sdl_mapping, calibration)

	return nil
}

func (a *App) SaveLogs(logs []string) error {
	location, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Select save location for logs",
		DefaultFilename: "output.log",
	})
	if err != nil {
		return err
	}

	output_log_file, err := os.OpenFile(location, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer output_log_file.Close()

	_, err = output_log_file.WriteString(strings.Join(logs, "\n"))
	if err != nil {
		return err
	}

	return nil
}

func (a *App) HasNewerVersion() bool {
	return true
}

func (a *App) UpdateApp() bool {
	return true
}

func (a *App) OpenConfigDirectory() error {
	var cmd *exec.Cmd
	switch go_runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", filepath.Clean(a.config.GlobalConfigDir))
	case "darwin":
		cmd = exec.Command("open", filepath.Clean(a.config.GlobalConfigDir))
	default:
		cmd = exec.Command("xdg-open", filepath.Clean(a.config.GlobalConfigDir))
	}
	if err := cmd.Start(); err != nil {
		logger.Logger.Error("[App::OpenConfigDirectory] could not open config directory", "error", err)
		return err
	}
	return nil
}

func (a *App) SelectCommAPIKeyFile() (string, error) {
	commapikey_path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select the CommAPIKey.txt file",
		Filters: []runtime.FileFilter{
			{DisplayName: "CommAPIKey File", Pattern: "*.txt"},
		},
	})

	if err != nil {
		return "", fmt.Errorf("please select the CommAPIKey.txt file: %w", err)
	}

	if filepath.Base(commapikey_path) != "CommAPIKey.txt" {
		return "", fmt.Errorf("please select the CommAPIKey.txt file")
	}

	return commapikey_path, nil
}

func (a *App) InstallTrainSimWorldMod() error {
	tsw_exe_path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Train Sim World 5/6 executable within TS2Prototype/Binaries/Win64 (TrainSimWorld.exe)",
	})
	if err != nil {
		return err
	}

	if !strings.HasSuffix(filepath.ToSlash(tsw_exe_path), "TS2Prototype/Binaries/Win64/TrainSimWorld.exe") {
		return fmt.Errorf("please select the TrainSimWorld.exe file in the game's TS2Prototype/Binaries/Win64 to install the mod")
	}

	var manifest ModAssets_Manifest
	manifest_json_bytes, err := embed_mod_assets_fs.ReadFile("embed/mod_assets/manifest.json")
	if err != nil {
		logger.Logger.Error("[App::InstallMod] failed to read manfiest file", "error", err)
		return err
	}

	if err := json.Unmarshal(manifest_json_bytes, &manifest); err != nil {
		return err
	}

	install_path := filepath.Dir(tsw_exe_path)
	/* go through files to copy */
	for _, entry := range manifest.Manifest {
		if entry.Action == ModAssets_Manifest_Entry_ActionType_Delete {
			// no action required if remove fails
			os.Remove(filepath.Join(install_path, entry.Path))
		} else if entry.Action == ModAssets_Manifest_Entry_ActionType_Copy {
			file_dir := filepath.Dir(entry.Path)
			if err := os.MkdirAll(filepath.Join(install_path, file_dir), 0755); err != nil {
				logger.Logger.Error("[App::InstallMod] could not create directory", "dir", filepath.Join(install_path, file_dir))
				return err
			}

			fh, err := embed_mod_assets_fs.Open(fmt.Sprintf("embed/mod_assets/%s", entry.Path))
			if err != nil {
				logger.Logger.Error("[App::InstallMod] could open file", "file", entry.Path)
				return fmt.Errorf("could not open file %e", err)
			}
			defer fh.Close()

			out, err := os.Create(filepath.Join(install_path, entry.Path))
			if err != nil {
				logger.Logger.Error("[App::InstallMod] could not create file", "file", filepath.Join(install_path, entry.Path))
				return fmt.Errorf("could not open create %e", err)
			}
			if _, err := io.Copy(out, fh); err != nil {
				logger.Logger.Error("[App::InstallMod] failed to copy file", "file", filepath.Join(install_path, entry.Path))
				return fmt.Errorf("failed to copy file: %w", err)
			}

			defer out.Close()
		}
	}

	/* write version file */
	os.WriteFile(filepath.Join(install_path, "ue4ss_tsw_controller_mod/Mods/TSWControllerMod/version.txt"), []byte(VERSION), 0755)
	a.program_config.LastInstalledModVersion = VERSION
	a.program_config.Save(filepath.Join(a.config.GlobalConfigDir, "program.json"))

	return nil
}

func (a *App) InstallTrainSimClassicMod() error {
	tsc_exe_path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select TS Classic executable (RailWorks64.exe)",
	})
	if err != nil {
		return err
	}

	if !strings.HasSuffix(filepath.ToSlash(tsc_exe_path), "/RailWorks64.exe") {
		return fmt.Errorf("please select the RailWorks64.exe file to install the mod")
	}

	var manifest ModAssets_Manifest
	manifest_json_bytes, err := embed_tsc_mod_assets_fs.ReadFile("embed/tsc_mod_assets/manifest.json")
	if err != nil {
		logger.Logger.Error("[App::InstallMod] failed to read manfiest file", "error", err)
		return err
	}

	if err := json.Unmarshal(manifest_json_bytes, &manifest); err != nil {
		return err
	}

	install_path := filepath.Dir(tsc_exe_path)
	/* go through files to copy */
	for _, entry := range manifest.Manifest {
		if entry.Action == ModAssets_Manifest_Entry_ActionType_Delete {
			// no action required if remove fails
			os.Remove(filepath.Join(install_path, entry.Path))
		} else if entry.Action == ModAssets_Manifest_Entry_ActionType_Copy {
			file_dir := filepath.Dir(entry.Path)
			if err := os.MkdirAll(filepath.Join(install_path, file_dir), 0755); err != nil {
				logger.Logger.Error("[App::InstallMod] could not create directory", "dir", filepath.Join(install_path, file_dir))
				return err
			}

			fh, err := embed_tsc_mod_assets_fs.Open(fmt.Sprintf("embed/tsc_mod_assets/%s", entry.Path))
			if err != nil {
				logger.Logger.Error("[App::InstallMod] could open file", "file", entry.Path)
				return fmt.Errorf("could not open file %e", err)
			}
			defer fh.Close()

			out, err := os.Create(filepath.Join(install_path, entry.Path))
			if err != nil {
				logger.Logger.Error("[App::InstallMod] could not create file", "file", filepath.Join(install_path, entry.Path))
				return fmt.Errorf("could not open create %e", err)
			}
			if _, err := io.Copy(out, fh); err != nil {
				logger.Logger.Error("[App::InstallMod] failed to copy file", "file", filepath.Join(install_path, entry.Path))
				return fmt.Errorf("failed to copy file: %w", err)
			}

			defer out.Close()
		}
	}

	/* write version file */
	os.WriteFile(filepath.Join(install_path, "plugins/version.txt"), []byte(VERSION), 0755)
	a.program_config.LastInstalledModVersion = VERSION
	a.program_config.Save(filepath.Join(a.config.GlobalConfigDir, "program.json"))

	return nil
}

func (a *App) tryWriteStructAsJSON(path string, data any) error {
	data_to_write, err := json.Marshal(data)
	if err != nil {
		return err
	}

	target_file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer target_file.Sync()
	defer target_file.Close()
	if _, err := target_file.Write(data_to_write); err != nil {
		return err
	}

	return nil
}

func (a *App) importProfileJSON(
	json_data []byte,
	metadata config.Config_Controller_Profile_Metadata,
) (*config.Config_Controller_Profile, error) {
	profile, err := config.ControllerProfileFromJSON(string(json_data), metadata)
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(filepath.Dir(metadata.Path), 0o755); err != nil {
		return nil, fmt.Errorf("could not create target location to save profile: %w", err)
	}

	/*
		check if the profile contains complete calibration and mapping information;
		if available and our calibration and mapping is missing; we should load it as well
	*/
	if profile.Controller != nil &&
		profile.Controller.Mapping != nil &&
		profile.Controller.Calibration != nil &&
		!a.sdl_controller_manager.IsConfigured(profile.Controller.Mapping.UsbID) {
		usb_id_slug := string_utils.Sluggify(profile.Controller.Mapping.UsbID)
		sdl_mappings_filepath := filepath.Join(a.config.GlobalConfigDir, config_loader.DIR_SDL_MAPPINGS_NAME, fmt.Sprintf("%s.sdl.json", usb_id_slug))
		calibration_filepath := filepath.Join(a.config.GlobalConfigDir, config_loader.DIR_CALIBRATION_NAME, fmt.Sprintf("%s.calibration.json", usb_id_slug))
		if err := a.tryWriteStructAsJSON(sdl_mappings_filepath, profile.Controller.Mapping); err != nil {
			return nil, fmt.Errorf("failed to import embedded SDL mapping: %w", err)
		}
		if err := a.tryWriteStructAsJSON(calibration_filepath, profile.Controller.Calibration); err != nil {
			return nil, fmt.Errorf("failed to import embedded calibration: %w", err)
		}
	}

	if err := a.tryWriteStructAsJSON(metadata.Path, profile); err != nil {
		return nil, fmt.Errorf("could not save profile at location %s: %w", metadata.Path, err)
	}
	return profile, nil
}

func (a *App) ImportProfile() error {
	import_profile_path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select a profile (.tswprofile)",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "TSW Profiles",
				Pattern:     "*.tswprofile",
			},
		},
	})
	if err != nil {
		return err
	}

	if filepath.Ext(import_profile_path) != ".tswprofile" {
		return fmt.Errorf("selected an invalid profile")
	}

	file_bytes, err := os.ReadFile(import_profile_path)
	if err != nil {
		return fmt.Errorf("could not read profile from location %s: %w", import_profile_path, err)
	}

	original_filename, _ := strings.CutSuffix(filepath.Base(import_profile_path), ".tswprofile")
	target_file_path := filepath.Join(a.config.GlobalConfigDir, "profiles", fmt.Sprintf("%s_%d.json", original_filename, time.Now().Unix()))
	if _, err = a.importProfileJSON(file_bytes, config.Config_Controller_Profile_Metadata{
		Path:      target_file_path,
		UpdatedAt: time.Now(),
	}); err != nil {
		return fmt.Errorf("could not import profile: %w", err)
	}

	return nil
}

func (a *App) ImportSharedProfile(profile Interop_SharedProfile) error {
	resp, err := http.Get(profile.Url)
	if err != nil {
		return fmt.Errorf("could not download profile")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not download profile")
	}

	target_file_path := filepath.Join(a.config.GlobalConfigDir, "profiles", fmt.Sprintf("%s_%d.json", string_utils.Sluggify(profile.Name), time.Now().Unix()))
	if _, err = a.importProfileJSON(body, config.Config_Controller_Profile_Metadata{
		Path:      target_file_path,
		UpdatedAt: time.Now(),
	}); err != nil {
		return fmt.Errorf("could not import profile from repository: %w", err)
	}

	return nil
}

/*
Saves a profile control mapping;
- if the profile already exists by name it will be merged
- if the profile does not exist it will be newly created and saved
*/
func (a *App) SaveControlMapping(
	mapping Interop_SaveControlMapping,
) error {
	profile, err := config.ControllerProfileFromJSON(mapping.ProfileJSON, config.Config_Controller_Profile_Metadata{
		Path: mapping.ExistingPath,
	})
	if err != nil {
		return fmt.Errorf("could not save profile: %w", err)
	}

	/* find profile merge */
	merged_profile := config.Config_Controller_Profile{
		Name: profile.Name,
	}
	a.profile_runner.Profiles.ForEach(func(p config.Config_Controller_Profile, key string) bool {
		if p.Metadata.Path == profile.Metadata.Path {
			merged_profile = p
			return false
		}
		return true
	})

	did_merge_with_existing_control := false
	for _, control := range profile.Controls {
		for index, existing_control := range merged_profile.Controls {
			if control.Name == existing_control.Name {
				assignments := existing_control.GetAssignments()
				assignments = append(assignments, control.GetAssignments()...)
				existing_control.Assignment = nil
				existing_control.Assignments = &assignments
				merged_profile.Controls[index] = existing_control
				did_merge_with_existing_control = true
			}
		}
		if !did_merge_with_existing_control {
			merged_profile.Controls = append(merged_profile.Controls, control)
		}
	}

	/* check if path is set; if not generate filename for saving */
	target_file_path := profile.Metadata.Path
	if target_file_path == "" {
		target_file_path = filepath.Join(a.config.GlobalConfigDir, "profiles", fmt.Sprintf("%s_%d.json", string_utils.Sluggify(profile.Name), time.Now().Unix()))
	}

	json_bytes, err := json.MarshalIndent(merged_profile, "", "  ")
	if err != nil {
		return fmt.Errorf("could not save profile: %w", err)
	}

	if _, err = a.importProfileJSON(json_bytes, config.Config_Controller_Profile_Metadata{
		Path:      target_file_path,
		UpdatedAt: time.Now(),
	}); err != nil {
		return fmt.Errorf("could not save profile: %w", err)
	}

	return nil
}
