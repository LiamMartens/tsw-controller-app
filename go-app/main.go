package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"tsw_controller_app/config_loader"
	"tsw_controller_app/logger"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/ingest"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

var VERSION = "1.0.0"

var AXIOM_TOKEN = ""
var AXIOM_ORG_ID = "tswcontrollerutility-szqw"
var AXIOM_DATASET = "app_logs"

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	arg_proxy := flag.String("proxy", "", "Enter the proxy address")
	flag.Parse()

	fmt.Printf("Version %s\n", VERSION)

	config_dir, err := os.UserConfigDir()
	if err != nil {
		panic(fmt.Errorf("could not find user config directory %e", err))
	}

	exec_file, err := os.Executable()
	if err != nil {
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

	mode := AppConfig_Mode_Default
	var proxy_settings *AppConfig_ProxySettings
	if arg_proxy != nil && *arg_proxy != "" {
		fmt.Printf("enabling proxy mode: %s\n", *arg_proxy)
		mode = AppConfig_Mode_Proxy
		proxy_settings = &AppConfig_ProxySettings{
			Addr: *arg_proxy,
		}
	}

	app := NewApp(AppConfig{
		GlobalConfigDir: global_config_dir,
		LocalConfigDir:  local_config_dir,
		Mode:            mode,
		ProxySettings:   proxy_settings,
	})

	if AXIOM_TOKEN != "" && AXIOM_ORG_ID != "" {
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
						event_to_send := axiom.Event{ingest.TimestampField: time.Now(), "message": msg.Message}
						go ax.IngestEvents(context.Background(), AXIOM_DATASET, []axiom.Event{event_to_send})
					}
				}
			}()
		}
	}

	err = wails.Run(&options.App{
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
