package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"tsw_controller_app/config"
	"tsw_controller_app/controller_mgr"
	"tsw_controller_app/logger"
	"tsw_controller_app/profile_runner"
	"tsw_controller_app/string_utils"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

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
