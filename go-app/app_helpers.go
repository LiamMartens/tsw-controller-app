package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	go_runtime "runtime"
	"strings"
	"tsw_controller_app/config"
	"tsw_controller_app/config_loader"
	"tsw_controller_app/logger"
	"tsw_controller_app/string_utils"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

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
