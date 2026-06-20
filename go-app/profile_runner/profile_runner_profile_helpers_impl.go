package profile_runner

import (
	"fmt"
	"sort"
	"tsw_controller_app/config"
	"tsw_controller_app/controller_mgr"
)

func (p *ProfileRunner) getSelectedProfileForDevice(device *controller_mgr.ControllerManager_ChangeEvent_Device) (ProfileRunnerSettings_SelectedProfile, bool) {
	selected_profile, has_selected_profile := p.Settings.GetSelectedProfiles().Get(device.UniqueID)

	/* try auto-selection */
	current_rail_class := p.CabDebugger.State.DrivableActorName
	if !has_selected_profile && current_rail_class != "" {
		scored_profiles := []ProfileRunner_ScoredProfileEntry{}

		p.Profiles.ForEach(func(profile config.Config_Controller_Profile, id string) bool {
			if (profile.AutoSelect == nil || !*profile.AutoSelect) ||
				profile.RailClassInformation == nil ||
				(profile.Controller != nil && *profile.Controller.UsbID != device.DeviceID) {
				/* skip if not-autoselect, rail class information is missing or the controller doesn't match */
				return true
			}

			/* we'll score any match which is not embedded higher than their embedded counterpart */
			score_factor := 1
			if !profile.Metadata.IsEmbedded {
				score_factor = 10
			}

			for _, rc_info := range *profile.RailClassInformation {
				if rc_info.ClassName == current_rail_class {
					is_controller_match := profile.Controller != nil && *profile.Controller.UsbID == device.DeviceID
					if is_controller_match {
						scored_profiles = append(scored_profiles, ProfileRunner_ScoredProfileEntry{Id: id, Score: 20 * score_factor})
					} else {
						scored_profiles = append(scored_profiles, ProfileRunner_ScoredProfileEntry{Id: id, Score: 10 * score_factor})
					}
					break
				}
			}

			return true
		})
		sort.Slice(scored_profiles, func(i, j int) bool {
			return scored_profiles[i].Score > scored_profiles[j].Score
		})

		if len(scored_profiles) > 0 {
			profile, _ := p.Profiles.Get(scored_profiles[0].Id)
			has_selected_profile = true
			selected_profile = ProfileRunnerSettings_SelectedProfile{
				Profile: profile,
			}
		}
	}

	return selected_profile, has_selected_profile
}

func (p *ProfileRunner) GetProfileNameToIdMap() map[string][]string {
	id_map_by_name := map[string][]string{}
	p.Profiles.ForEach(func(profile config.Config_Controller_Profile, id string) bool {
		if existing_ids, has_key := id_map_by_name[profile.Name]; has_key {
			id_map_by_name[profile.Name] = append(existing_ids, id)
		} else {
			id_map_by_name[profile.Name] = []string{id}
		}
		return true
	})
	return id_map_by_name
}

func (p *ProfileRunner) RegisterProfile(profile config.Config_Controller_Profile) {
	p.Profiles.Set(profile.Id(), profile)
}

func (p *ProfileRunner) Resolve() {
	/* resolves all the profiles */
	id_name_map := p.GetProfileNameToIdMap()
	p.Profiles.Mutex.Lock()
	defer p.Profiles.Mutex.Unlock()

	var resolve_profile func(profile config.Config_Controller_Profile) config.Config_Controller_Profile
	resolve_profile = func(profile config.Config_Controller_Profile) config.Config_Controller_Profile {
		if profile.Extends != nil && len(*profile.Extends) > 0 && profile.Name != *profile.Extends {
			if extend_from_profile_ids, has_extendable_ids := id_name_map[*profile.Extends]; has_extendable_ids {
				if len(extend_from_profile_ids) == 0 || len(extend_from_profile_ids) > 1 {
					/* only extend if there is one and only one profile to extend from */
					return profile
				}
				extend_from_profile := p.Profiles.Map[extend_from_profile_ids[0]]

				/*
					these are the control names which are defined in the profile we are currently resolving;
					these should be kept as they already have a definition
				*/
				existing_control_definitions := map[string]bool{}
				for _, control := range profile.Controls {
					existing_control_definitions[control.Name] = true
				}

				resolved_extend_from_profile := resolve_profile(extend_from_profile)
				for _, control := range resolved_extend_from_profile.Controls {
					if _, should_not_override := existing_control_definitions[control.Name]; !should_not_override {
						profile.Controls = append(profile.Controls, control)
					}
				}

				if profile.Controller == nil && extend_from_profile.Controller != nil {
					profile.Controller = extend_from_profile.Controller
				}
			}
		}
		return profile
	}

	for profile_id, profile := range p.Profiles.Map {
		p.Profiles.Map[profile_id] = resolve_profile(profile)
	}
}

func (p *ProfileRunner) ClearProfile(unique_id controller_mgr.DeviceUniqueID) {
	p.Settings.Update(func(s *ProfileRunnerSettings) {
		s.SelectedProfilesByUniqueID.Delete(unique_id)
	})
}

func (p *ProfileRunner) SetProfile(unique_id controller_mgr.DeviceUniqueID, id string) error {
	var err error = nil
	p.Settings.Update(func(s *ProfileRunnerSettings) {
		profile, is_valid_profile := p.Profiles.Get(id)
		if is_valid_profile {
			s.SelectedProfilesByUniqueID.Set(unique_id, ProfileRunnerSettings_SelectedProfile{
				Profile: profile,
			})
		} else {
			err = fmt.Errorf("could not find profile by ID %s", id)
		}
	})
	return err
}
