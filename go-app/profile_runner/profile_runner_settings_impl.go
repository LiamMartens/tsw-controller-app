package profile_runner

import (
	"tsw_controller_app/config"
	"tsw_controller_app/controller_mgr"
	"tsw_controller_app/map_utils"
)

func (s *ProfileRunnerSettings) Update(mutator func(s *ProfileRunnerSettings)) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	mutator(s)
}

func (s *ProfileRunnerSettings) GetSelectedProfiles() *map_utils.LockMap[controller_mgr.DeviceUniqueID, ProfileRunnerSettings_SelectedProfile] {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return s.SelectedProfilesByUniqueID
}

func (s *ProfileRunnerSettings) GetPreferredControlMode() config.PreferredControlMode {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return s.PreferredControlMode
}

func (s *ProfileRunnerSettings) SetPreferredControlMode(mode config.PreferredControlMode) {
	s.Mutex.Lock()
	s.PreferredControlMode = mode
	defer s.Mutex.Unlock()
}
