package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"tsw_controller_app/logger"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) InstallTrainSimWorldMod() error {
	tsw_exe_path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Train Sim World 5/6 executable within TS2Prototype/Binaries/Win64 (TrainSimWorld.exe)",
	})
	if err != nil {
		return err
	}

	if !strings.HasSuffix(filepath.ToSlash(tsw_exe_path), "TS2Prototype/Binaries/Win64/TrainSimWorld.exe") {
		return fmt.Errorf("please select the TrainSimWorld.exe file in the game's TS2Prototype/Binaries/Win64 to install the mod")
	}

	var manifest ModAssets_Manifest
	manifest_json_bytes, err := embed_mod_assets_fs.ReadFile("embed/mod_assets/manifest.json")
	if err != nil {
		logger.Logger.Error("[App::InstallMod] failed to read manfiest file", "error", err)
		return err
	}

	if err := json.Unmarshal(manifest_json_bytes, &manifest); err != nil {
		return err
	}

	install_path := filepath.Dir(tsw_exe_path)
	/* go through files to copy */
	for _, entry := range manifest.Manifest {
		if entry.Action == ModAssets_Manifest_Entry_ActionType_Delete {
			// no action required if remove fails
			os.Remove(filepath.Join(install_path, entry.Path))
		} else if entry.Action == ModAssets_Manifest_Entry_ActionType_Copy {
			file_dir := filepath.Dir(entry.Path)
			if err := os.MkdirAll(filepath.Join(install_path, file_dir), 0755); err != nil {
				logger.Logger.Error("[App::InstallMod] could not create directory", "dir", filepath.Join(install_path, file_dir))
				return err
			}

			fh, err := embed_mod_assets_fs.Open(fmt.Sprintf("embed/mod_assets/%s", entry.Path))
			if err != nil {
				logger.Logger.Error("[App::InstallMod] could open file", "file", entry.Path)
				return fmt.Errorf("could not open file %e", err)
			}
			defer fh.Close()

			out, err := os.Create(filepath.Join(install_path, entry.Path))
			if err != nil {
				logger.Logger.Error("[App::InstallMod] could not create file", "file", filepath.Join(install_path, entry.Path))
				return fmt.Errorf("could not open create %e", err)
			}
			if _, err := io.Copy(out, fh); err != nil {
				logger.Logger.Error("[App::InstallMod] failed to copy file", "file", filepath.Join(install_path, entry.Path))
				return fmt.Errorf("failed to copy file: %w", err)
			}

			defer out.Close()
		}
	}

	/* write version file */
	os.WriteFile(filepath.Join(install_path, "ue4ss_tsw_controller_mod/Mods/TSWControllerMod/version.txt"), []byte(VERSION), 0755)
	a.program_config.LastInstalledModVersion = VERSION
	a.program_config.Save(filepath.Join(a.config.GlobalConfigDir, "program.json"))

	return nil
}

func (a *App) InstallTrainSimClassicMod() error {
	tsc_exe_path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select TS Classic executable (RailWorks64.exe)",
	})
	if err != nil {
		return err
	}

	if !strings.HasSuffix(filepath.ToSlash(tsc_exe_path), "/RailWorks64.exe") {
		return fmt.Errorf("please select the RailWorks64.exe file to install the mod")
	}

	var manifest ModAssets_Manifest
	manifest_json_bytes, err := embed_tsc_mod_assets_fs.ReadFile("embed/tsc_mod_assets/manifest.json")
	if err != nil {
		logger.Logger.Error("[App::InstallMod] failed to read manfiest file", "error", err)
		return err
	}

	if err := json.Unmarshal(manifest_json_bytes, &manifest); err != nil {
		return err
	}

	install_path := filepath.Dir(tsc_exe_path)
	/* go through files to copy */
	for _, entry := range manifest.Manifest {
		if entry.Action == ModAssets_Manifest_Entry_ActionType_Delete {
			// no action required if remove fails
			os.Remove(filepath.Join(install_path, entry.Path))
		} else if entry.Action == ModAssets_Manifest_Entry_ActionType_Copy {
			file_dir := filepath.Dir(entry.Path)
			if err := os.MkdirAll(filepath.Join(install_path, file_dir), 0755); err != nil {
				logger.Logger.Error("[App::InstallMod] could not create directory", "dir", filepath.Join(install_path, file_dir))
				return err
			}

			fh, err := embed_tsc_mod_assets_fs.Open(fmt.Sprintf("embed/tsc_mod_assets/%s", entry.Path))
			if err != nil {
				logger.Logger.Error("[App::InstallMod] could open file", "file", entry.Path)
				return fmt.Errorf("could not open file %e", err)
			}
			defer fh.Close()

			out, err := os.Create(filepath.Join(install_path, entry.Path))
			if err != nil {
				logger.Logger.Error("[App::InstallMod] could not create file", "file", filepath.Join(install_path, entry.Path))
				return fmt.Errorf("could not open create %e", err)
			}
			if _, err := io.Copy(out, fh); err != nil {
				logger.Logger.Error("[App::InstallMod] failed to copy file", "file", filepath.Join(install_path, entry.Path))
				return fmt.Errorf("failed to copy file: %w", err)
			}

			defer out.Close()
		}
	}

	/* write version file */
	os.WriteFile(filepath.Join(install_path, "plugins/version.txt"), []byte(VERSION), 0755)
	a.program_config.LastInstalledModVersion = VERSION
	a.program_config.Save(filepath.Join(a.config.GlobalConfigDir, "program.json"))

	return nil
}

func (a *App) InstallWondersOfSodorMod() error {
	exe_path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Wonders of Sodor executable within TS2Prototype/Binaries/Win64 (WondersOfSodor.exe)",
	})
	if err != nil {
		return err
	}

	if !strings.HasSuffix(filepath.ToSlash(exe_path), "TS2Prototype/Binaries/Win64/WondersOfSodor.exe") {
		return fmt.Errorf("please select the WondersOfSodor.exe file in the game's TS2Prototype/Binaries/Win64 to install the mod")
	}

	var manifest ModAssets_Manifest
	manifest_json_bytes, err := embed_mod_assets_fs.ReadFile("embed/mod_assets/manifest.json")
	if err != nil {
		logger.Logger.Error("[App::InstallMod] failed to read manifest file", "error", err)
		return err
	}

	if err := json.Unmarshal(manifest_json_bytes, &manifest); err != nil {
		return err
	}

	install_path := filepath.Dir(exe_path)
	/* go through files to copy */
	for _, entry := range manifest.Manifest {
		if entry.Action == ModAssets_Manifest_Entry_ActionType_Delete {
			// no action required if remove fails
			os.Remove(filepath.Join(install_path, entry.Path))
		} else if entry.Action == ModAssets_Manifest_Entry_ActionType_Copy {
			file_dir := filepath.Dir(entry.Path)
			if err := os.MkdirAll(filepath.Join(install_path, file_dir), 0755); err != nil {
				logger.Logger.Error("[App::InstallMod] could not create directory", "dir", filepath.Join(install_path, file_dir))
				return err
			}

			fh, err := embed_mod_assets_fs.Open(fmt.Sprintf("embed/mod_assets/%s", entry.Path))
			if err != nil {
				logger.Logger.Error("[App::InstallMod] could open file", "file", entry.Path)
				return fmt.Errorf("could not open file %e", err)
			}
			defer fh.Close()

			out, err := os.Create(filepath.Join(install_path, entry.Path))
			if err != nil {
				logger.Logger.Error("[App::InstallMod] could not create file", "file", filepath.Join(install_path, entry.Path))
				return fmt.Errorf("could not open create %e", err)
			}
			if _, err := io.Copy(out, fh); err != nil {
				logger.Logger.Error("[App::InstallMod] failed to copy file", "file", filepath.Join(install_path, entry.Path))
				return fmt.Errorf("failed to copy file: %w", err)
			}

			defer out.Close()
		}
	}

	/* write version file */
	os.WriteFile(filepath.Join(install_path, "ue4ss_tsw_controller_mod/Mods/TSWControllerMod/version.txt"), []byte(VERSION), 0755)
	a.program_config.LastInstalledModVersion = VERSION
	a.program_config.Save(filepath.Join(a.config.GlobalConfigDir, "program.json"))

	return nil
}

func (a *App) InstallRunningTrainMod() error {
	exe_path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Running Train executable within RUNNING RunningTrain/Binaries/Win64 directory (RunningTrain-Win64-Shipping.exe)",
	})
	if err != nil {
		return err
	}

	if !strings.HasSuffix(filepath.ToSlash(exe_path), "RunningTrain/Binaries/Win64/RunningTrain-Win64-Shipping.exe") {
		return fmt.Errorf("please select the RunningTrain-Win64-Shipping.exe file in the game's RunningTrain/Binaries/Win64 directory to install the mod")
	}

	var manifest ModAssets_Manifest
	manifest_json_bytes, err := embed_running_train_mod_assets_fs.ReadFile("embed/running_train_mod_assets/manifest.json")
	if err != nil {
		logger.Logger.Error("[App::InstallMod] failed to read manifest file", "error", err)
		return err
	}

	if err := json.Unmarshal(manifest_json_bytes, &manifest); err != nil {
		return err
	}

	install_path := filepath.Dir(exe_path)
	/* go through files to copy */
	for _, entry := range manifest.Manifest {
		if entry.Action == ModAssets_Manifest_Entry_ActionType_Delete {
			// no action required if remove fails
			os.Remove(filepath.Join(install_path, entry.Path))
		} else if entry.Action == ModAssets_Manifest_Entry_ActionType_Copy {
			file_dir := filepath.Dir(entry.Path)
			if err := os.MkdirAll(filepath.Join(install_path, file_dir), 0755); err != nil {
				logger.Logger.Error("[App::InstallMod] could not create directory", "dir", filepath.Join(install_path, file_dir))
				return err
			}

			fh, err := embed_running_train_mod_assets_fs.Open(fmt.Sprintf("embed/running_train_mod_assets/%s", entry.Path))
			if err != nil {
				logger.Logger.Error("[App::InstallMod] could open file", "file", entry.Path)
				return fmt.Errorf("could not open file %e", err)
			}
			defer fh.Close()

			out, err := os.Create(filepath.Join(install_path, entry.Path))
			if err != nil {
				logger.Logger.Error("[App::InstallMod] could not create file", "file", filepath.Join(install_path, entry.Path))
				return fmt.Errorf("could not open create %e", err)
			}
			if _, err := io.Copy(out, fh); err != nil {
				logger.Logger.Error("[App::InstallMod] failed to copy file", "file", filepath.Join(install_path, entry.Path))
				return fmt.Errorf("failed to copy file: %w", err)
			}

			defer out.Close()
		}
	}

	/* write version file */
	os.WriteFile(filepath.Join(install_path, "ue4ss_tsw_controller_mod/Mods/RunningTrainControllerMod/version.txt"), []byte(VERSION), 0755)
	a.program_config.LastInstalledModVersion = VERSION
	a.program_config.Save(filepath.Join(a.config.GlobalConfigDir, "program.json"))

	return nil
}
