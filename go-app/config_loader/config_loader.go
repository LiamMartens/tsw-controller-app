package config_loader

import (
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"strings"
	"tsw_controller_app/config"
)

const DIR_CALIBRATION_NAME = "calibration"
const DIR_SDL_MAPPINGS_NAME = "sdl_mappings"
const DIR_PROFILES_NAME = "profiles"

type ConfigLoader struct{}

type ConfigLoader_FromFS_Options struct {
	Path     string
	Embedded bool
}

func New() *ConfigLoader {
	return &ConfigLoader{}
}

func (c *ConfigLoader) FromFS(fsys fs.FS, options ConfigLoader_FromFS_Options) ([]config.Config_Controller_SDLMap, []config.Config_Controller_Calibration, []config.Config_Controller_Profile, []error) {
	/* since this method accepts a fsys we'll use path. not filepath */

	var errors []error

	calibration_files_dir := path.Join(DIR_CALIBRATION_NAME)
	sdl_mapping_files_dir := path.Join(DIR_SDL_MAPPINGS_NAME)
	profiles_files_dir := path.Join(DIR_PROFILES_NAME)

	calibration_file_entries, err := fs.ReadDir(fsys, calibration_files_dir)
	var parsed_calibration_files []config.Config_Controller_Calibration
	if err != nil {
		errors = append(errors, fmt.Errorf("could not read calibration directory %s (%e)", calibration_files_dir, err))
	} else {
		for _, entry := range calibration_file_entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				file_bytes, err := fs.ReadFile(fsys, path.Join(calibration_files_dir, entry.Name()))
				if err != nil {
					errors = append(errors, fmt.Errorf("could not read calibration file %s (%e)", entry.Name(), err))
					continue
				}
				calibration, err := config.ControllerCalibrationFromJSON(string(file_bytes))
				if err != nil {
					errors = append(errors, fmt.Errorf("could not parse calibration file %s (%e)", entry.Name(), err))
					continue
				}
				parsed_calibration_files = append(parsed_calibration_files, *calibration)
			}
		}
	}

	sdl_mappings_file_entries, err := fs.ReadDir(fsys, sdl_mapping_files_dir)
	var parsed_sdl_mappings_files []config.Config_Controller_SDLMap
	if err != nil {
		errors = append(errors, fmt.Errorf("could not read SDL mappings directory %s (%e)", sdl_mapping_files_dir, err))
	} else {
		for _, entry := range sdl_mappings_file_entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				file_bytes, err := fs.ReadFile(fsys, path.Join(sdl_mapping_files_dir, entry.Name()))
				if err != nil {
					errors = append(errors, fmt.Errorf("could not read SDL mapping file %s (%e)", entry.Name(), err))
					continue
				}
				sdl_mapping, err := config.ControllerSDLMapFromJSON(string(file_bytes))
				if err != nil {
					errors = append(errors, fmt.Errorf("could not parse SDL mapping file %s (%e)", entry.Name(), err))
					continue
				}
				parsed_sdl_mappings_files = append(parsed_sdl_mappings_files, *sdl_mapping)
			}
		}
	}

	profiles_file_entries, err := fs.ReadDir(fsys, profiles_files_dir)
	var parsed_profile_files []config.Config_Controller_Profile
	if err != nil {
		errors = append(errors, fmt.Errorf("could not read profiles directory %s (%e)", profiles_files_dir, err))
	} else {
		for _, entry := range profiles_file_entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				fullpath := path.Join(profiles_files_dir, entry.Name())
				filestat, err := fs.Stat(fsys, fullpath)
				if err != nil {
					errors = append(errors, fmt.Errorf("could not read profile info %s (%e)", entry.Name(), err))
					continue
				}
				file_bytes, err := fs.ReadFile(fsys, fullpath)
				if err != nil {
					errors = append(errors, fmt.Errorf("could not read profile file %s (%e)", entry.Name(), err))
					continue
				}
				profile_metadata := config.Config_Controller_Profile_Metadata{
					/* this should be the actual OS file path since it will be used later in OS contexts */
					Path:       filepath.Join(options.Path, fullpath),
					IsEmbedded: options.Embedded,
					UpdatedAt:  filestat.ModTime(),
				}
				profile, err := config.ControllerProfileFromJSON(string(file_bytes), profile_metadata)
				if err != nil {
					errors = append(errors, fmt.Errorf("could not parse profile %s (%e)", entry.Name(), err))
					continue
				}
				parsed_profile_files = append(parsed_profile_files, *profile)
			}
		}
	}

	return parsed_sdl_mappings_files, parsed_calibration_files, parsed_profile_files, errors
}
