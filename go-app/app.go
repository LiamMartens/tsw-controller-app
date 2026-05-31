package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
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

	"github.com/google/uuid"
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
	sdllib                     sdl_mgr.SDL_Library
	session_id                 string
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

func NewApp(
	appconfig AppConfig,
) *App {
	sdl_manager := sdl_mgr.New()

	program_config := config.LoadProgramConfigFromFile(filepath.Join(appconfig.GlobalConfigDir, "program.json"))
	if program_config.TSWAPIKeyLocation == "" {
		program_config.TSWAPIKeyLocation = program_config.AutoDetectTSWAPIKeyLocation()
	}

	return &App{
		session_id:     uuid.NewString(),
		config:         appconfig,
		program_config: program_config,
		config_loader:  config_loader.New(),
		sdl_manager:    sdl_manager,
	}
}

func (a *App) GetVersion() string {
	return VERSION
}

func (a *App) GetControlServerAddr() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip := localAddr.IP.String()
	port := a.connector.Port()
	return fmt.Sprintf("%s:%d", ip, port), nil
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
				if sdl_mapping.Matches(&c) {
					calibration = &c
					break
				}
			}
			if calibration != nil {
				logger.Logger.Debug("[App] registering SDL map and calibration for controller", "name", sdl_mapping.Name, "usb_id", sdl_mapping.UsbID, "unique_id", sdl_mapping.UniqueID)
				a.sdl_controller_manager.RegisterConfig(sdl_mapping, *calibration)
			}
		}

		for _, profile := range profiles {
			logger.Logger.Debug("[App] registering profile", "profile", profile.Id(), "name", profile.Name)
			a.profile_runner.RegisterProfile(profile)
		}
	}

	a.profile_runner.Resolve()
	runtime.EventsEmit(a.ctx, AppEventType_ProfilesUpdated)
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

func (a *App) SaveCalibration(data Interop_ControllerCalibration) error {
	sdl_mapping := config.Config_Controller_SDLMap{
		Name:     data.Name,
		UsbID:    data.DeviceID,
		UniqueID: data.UniqueID,
		Data:     []config.Config_Controller_SDLMap_Control{},
	}
	calibration := config.Config_Controller_Calibration{
		UsbID:    data.DeviceID,
		UniqueID: data.UniqueID,
		Data:     []config.Config_Controller_CalibrationData{},
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
					Thresholds:  control.Thresholds,
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
