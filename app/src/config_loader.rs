use std::{fs, path::Path};

use log::{info, warn};

use super::config_defs::{controller_calibration::ControllerCalibration, controller_profile::ControllerProfile, controller_sdl_map::ControllerSdlMap};

pub struct ConfigLoader {
    pub controller_sdl_mappings: Vec<ControllerSdlMap>,
    pub controller_calibrations: Vec<ControllerCalibration>,
    pub controller_profiles: Vec<ControllerProfile>,
}

impl ConfigLoader {
    pub fn new() -> ConfigLoader {
        ConfigLoader {
            controller_sdl_mappings: Vec::new(),
            controller_calibrations: Vec::new(),
            controller_profiles: Vec::new(),
        }
    }

    pub fn register_sdl_mapping(&mut self, mapping: ControllerSdlMap) {
        self.controller_sdl_mappings.push(mapping);
    }

    pub fn register_calibration(&mut self, calibration: ControllerCalibration) {
        self.controller_calibrations.push(calibration);
    }

    pub fn load_from_dir<T: AsRef<str>>(&mut self, config_dir: Option<T>) {
        let config_dir_option = config_dir.as_ref();
        let config_dir = match config_dir_option {
            Some(dir) => dir.as_ref(),
            None => "config",
        };
        // read the calibration, sdl mappings and profile files from the provided config dir
        let sdl_mappings_path = Path::new(config_dir).join("sdl_mappings");
        let calibration_path = Path::new(config_dir).join("calibration");
        let profiles_path = Path::new(config_dir).join("profiles");

        let sdl_mapping_files = match fs::read_dir(sdl_mappings_path) {
            Ok(files) => files.into_iter().filter_map(Result::ok).collect(),
            Err(_) => Vec::new(),
        };
        let calibration_files = match fs::read_dir(calibration_path) {
            Ok(files) => files.into_iter().filter_map(Result::ok).collect(),
            Err(_) => Vec::new(),
        };
        let profile_files = match fs::read_dir(profiles_path) {
            Ok(files) => files.into_iter().filter_map(Result::ok).collect(),
            Err(_) => Vec::new(),
        };

        info!("Found {} SDL mapping files", sdl_mapping_files.len());
        for file in sdl_mapping_files.iter() {
            match fs::read_to_string(file.path()) {
                Ok(contents) => match serde_json::from_str(&contents) {
                    Ok(sdl_map) => {
                        info!("Successfully read SDL mapping file: {:?}", file.path());
                        self.controller_sdl_mappings.push(sdl_map);
                    }
                    Err(e) => {
                        warn!("Could not parse SDL mapping file {:?}: {}", file.path(), e);
                    }
                },
                Err(e) => {
                    warn!("Could not read SDL mapping file {:?}: {}", file.path(), e);
                }
            }
        }

        info!("Found {} calibration files", calibration_files.len());
        for file in calibration_files.iter() {
            match fs::read_to_string(file.path()) {
                Ok(contents) => match serde_json::from_str(&contents) {
                    Ok(sdl_map) => {
                        info!("Successfully read calibration file: {:?}", file.path());
                        self.controller_calibrations.push(sdl_map);
                    }
                    Err(e) => {
                        warn!("Could not parse calibration file {:?}: {}", file.path(), e);
                    }
                },
                Err(e) => {
                    warn!("Could not read calibration file {:?}: {}", file.path(), e);
                }
            }
        }

        info!("Found {} profile files", profile_files.len());
        for file in profile_files.iter() {
            match fs::read_to_string(file.path()) {
                Ok(contents) => match serde_json::from_str(&contents) {
                    Ok(sdl_map) => {
                        info!("Successfully read profile file: {:?}", file.path());
                        self.controller_profiles.push(sdl_map);
                    }
                    Err(e) => {
                        warn!("Could not parse profile file {:?}: {}", file.path(), e);
                    }
                },
                Err(e) => {
                    warn!("Could not read profile file {:?}: {}", file.path(), e);
                }
            }
        }
        /* sort */
        self.controller_profiles.sort_by(|a, b| a.name.cmp(&b.name));
    }

    pub fn export<T: AsRef<str>>(&self, config_dir: T) {
        let config_dir = config_dir.as_ref();
        let sdl_mappings_path = Path::new(config_dir).join("sdl_mappings");
        let calibration_path = Path::new(config_dir).join("calibration");
        let profiles_path = Path::new(config_dir).join("profiles");

        fs::create_dir_all(&sdl_mappings_path).unwrap();
        fs::create_dir_all(&calibration_path).unwrap();
        fs::create_dir_all(&profiles_path).unwrap();

        for mapping in self.controller_sdl_mappings.iter() {
            let file_path = sdl_mappings_path.join(format!("{}.json", mapping.name));
            let json = serde_json::to_string_pretty(mapping).unwrap();
            fs::write(file_path, json).unwrap();
        }

        for calibration in self.controller_calibrations.iter() {
            let file_path = calibration_path.join(format!("{}.json", calibration.usb_id));
            let json = serde_json::to_string_pretty(calibration).unwrap();
            fs::write(file_path, json).unwrap();
        }

        for profile in self.controller_profiles.iter() {
            let file_path = profiles_path.join(format!("{}.json", profile.name));
            let json = serde_json::to_string_pretty(profile).unwrap();
            fs::write(file_path, json).unwrap();
        }
    }

    pub fn find_sdl_mapping(&self, usb_id: &String) -> Option<&ControllerSdlMap> {
        self.controller_sdl_mappings.iter().find(|m| &m.usb_id == usb_id)
    }

    pub fn find_controller_calibration(&self, usb_id: &String) -> Option<&ControllerCalibration> {
        self.controller_calibrations.iter().find(|m| &m.usb_id == usb_id)
    }

    pub fn find_controller_profile<T: AsRef<str>>(&self, name: T, controller_usb_id: Option<String>) -> Option<&ControllerProfile> {
        let fallback_profile = self.controller_profiles.iter().find(|m| m.name == name.as_ref() && m.usb_id.is_none());

        match controller_usb_id {
            Some(usb_id) => {
                let override_profile = self
                    .controller_profiles
                    .iter()
                    .find(|m| m.name == name.as_ref() && m.usb_id.is_some() && m.usb_id.as_ref().unwrap() == &usb_id);
                match override_profile {
                    Some(profile) => Some(profile),
                    None => fallback_profile,
                }
            }
            None => fallback_profile,
        }
    }
}
