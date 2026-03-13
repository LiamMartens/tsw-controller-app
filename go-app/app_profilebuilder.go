package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"tsw_controller_app/config"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

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
