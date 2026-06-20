package sdl_mgr

import "fmt"

func (hd *SDLMgr_HIDDevice) Version() uint16 {
	if hd.Go_Backend_Device != nil {
		return hd.Go_Backend_Device.Version()
	}
	if hd.Native_Backend_Device != nil {
		return hd.Native_Backend_Device.Release
	}
	return 0
}

func (hd *SDLMgr_HIDDevice) Serial() string {
	if hd.Go_Backend_Device != nil {
		return hd.Go_Backend_Device.SerialNumber()
	}
	if hd.Native_Backend_Device != nil {
		return hd.Native_Backend_Device.Serial
	}
	return ""
}

func (hd *SDLMgr_HIDDevice) ReadFeatureReport(id byte) ([]byte, error) {
	if hd.Go_Backend_Device != nil {
		if err := hd.Go_Backend_Device.Open(false); err != nil {
			return nil, err
		}
		defer hd.Go_Backend_Device.Close()

		report, err := hd.Go_Backend_Device.GetFeatureReport(id)
		if err != nil {
			return nil, err
		}
		return report, nil
	}

	if hd.Native_Backend_Device != nil {
		opened_device, err := hd.Native_Backend_Device.Open()
		if err != nil {
			return nil, err
		}
		defer opened_device.Close()

		var report []byte = []byte{id}
		if _, err := opened_device.GetFeatureReport(report); err != nil {
			return nil, err
		}
		return report[1:], nil
	}

	return nil, fmt.Errorf("no hid device available")
}

func (hd *SDLMgr_HIDDevice) SendFeatureReport(id byte, data []byte) error {
	if hd.Go_Backend_Device != nil {
		if err := hd.Go_Backend_Device.Open(false); err != nil {
			return err
		}
		defer hd.Go_Backend_Device.Close()

		err := hd.Go_Backend_Device.SetFeatureReport(id, data)
		if err != nil {
			return err
		}
		return nil
	}

	if hd.Native_Backend_Device != nil {
		opened_device, err := hd.Native_Backend_Device.Open()
		if err != nil {
			return err
		}
		defer opened_device.Close()

		var report []byte = []byte{id}
		report = append(report, data...)
		if _, err := opened_device.SendFeatureReport(report); err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("no hid device available")
}
