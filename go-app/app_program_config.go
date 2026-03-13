package main

import (
	"path/filepath"
	"tsw_controller_app/config"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

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
