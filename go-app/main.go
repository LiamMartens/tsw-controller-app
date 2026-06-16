package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
	"tsw_controller_app/config_loader"
	"tsw_controller_app/logger"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/ingest"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

var VERSION = "1.0.0"

var AXIOM_TOKEN = ""
var AXIOM_ORG_ID = "tswcontrollerutility-szqw"
var AXIOM_DATASET = "app_logs"

//go:embed all:frontend/dist
var assets embed.FS

func isWayland() bool {
	if wayland, _ := os.LookupEnv("WAYLAND_DISPLAY"); wayland != "" {
		return true
	}
	return false
}

func initConfigDirs() (string, string) {
	config_dir, err := os.UserConfigDir()
	if err != nil {
		logger.Logger.Error("[main] could not determine user configuration directory", "error", err)
		panic(fmt.Errorf("could not find user config directory %e", err))
	}

	exec_file, err := os.Executable()
	if err != nil {
		logger.Logger.Error("[main] could not determine own executable", "error", err)
		panic(fmt.Errorf("could not find executable %e", err))
	}

	global_config_dir := filepath.Join(config_dir, "tswcontrollerapp/config")
	local_config_dir := filepath.Join(filepath.Dir(exec_file), "config")
	required_subpaths := []string{config_loader.DIR_SDL_MAPPINGS_NAME, config_loader.DIR_CALIBRATION_NAME, config_loader.DIR_PROFILES_NAME}

	os.MkdirAll(global_config_dir, 0o755)
	os.MkdirAll(local_config_dir, 0o755)
	for _, subpath := range required_subpaths {
		os.MkdirAll(filepath.Join(global_config_dir, subpath), 0o755)
	}

	return global_config_dir, local_config_dir
}

func initAxiom(session_id string) {
	if AXIOM_TOKEN == "" || AXIOM_ORG_ID == "" {
		return
	}

	ax, err := axiom.NewClient(axiom.SetPersonalTokenConfig(AXIOM_TOKEN, AXIOM_ORG_ID))
	if err != nil {
		logger.Logger.Error("could not instantiate logging client", "error", err)
	} else {
		go func() {
			ctx := context.Background()
			logchan, unsubscribe := logger.Logger.Listen()
			defer unsubscribe()
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-logchan:
					if msg.LogLevel == "info" || msg.LogLevel == "error" {
						event_to_send := axiom.Event{
							ingest.TimestampField: time.Now(),
							"version":             VERSION,
							"message":             msg.Message,
							"platform":            runtime.GOOS,
							"session_id":          session_id,
						}
						go ax.IngestEvents(context.Background(), AXIOM_DATASET, []axiom.Event{event_to_send})
					}
				}
			}
		}()
	}
}

func main() {
	if isWayland() {
		os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "1")
	}

	arg_proxy := flag.String("proxy", "", "Enter the proxy address")
	flag.Parse()

	logger.Logger.Debug("[main] version", "version", VERSION)

	mode := AppConfig_Mode_Default
	var proxy_settings *AppConfig_ProxySettings
	if arg_proxy != nil && *arg_proxy != "" {
		logger.Logger.Debug("[main] running in proxy mode", "proxy", *arg_proxy)
		mode = AppConfig_Mode_Proxy
		proxy_settings = &AppConfig_ProxySettings{
			Addr: *arg_proxy,
		}
	}

	global_config_dir, local_config_dir := initConfigDirs()

	app := NewApp(AppConfig{
		GlobalConfigDir: global_config_dir,
		LocalConfigDir:  local_config_dir,
		Mode:            mode,
		ProxySettings:   proxy_settings,
	})
	initAxiom(app.session_id)

	err := wails.Run(&options.App{
		Title:  "TSW Controller Utility",
		Width:  600,
		Height: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			TitleBar: mac.TitleBarDefault(),
			About: &mac.AboutInfo{
				Title:   "TSW Controller App",
				Message: "(c) 2026",
			},
		},
		Windows: &windows.Options{
			WebviewGpuIsDisabled: false,
		},
		Linux: &linux.Options{
			WindowIsTranslucent: false,
			WebviewGpuPolicy:    linux.WebviewGpuPolicyOnDemand,
		},
	})

	if err != nil {
		logger.Logger.Error("[main] error", "error", err)
	}
}
