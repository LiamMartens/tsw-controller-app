package sdl_mgr

import (
	"errors"
	"fmt"

	usbhid "rafaelmartins.com/p/usbhid"
)

var _ ISDLMgr_HIDDevice = &SDLMgr_HIDDevice{}

func (d *SDLMgr_HIDDevice_Native) IsOpen() bool {
	d.Lock.Lock()
	defer d.Lock.Unlock()
	return d.Device != nil
}

func (d *SDLMgr_HIDDevice_Native) Open() error {
	d.Lock.Lock()
	defer d.Lock.Unlock()
	if d.Device != nil {
		return nil
	}
	device, err := d.DeviceInfo.Open()
	if err != nil {
		return err
	}
	d.Device = device
	return nil
}

func (d *SDLMgr_HIDDevice_Native) Close() error {
	d.Lock.Lock()
	defer d.Lock.Unlock()
	if d.Device != nil {
		err := d.Device.Close()
		d.Device = nil
		if err != nil {
			return err
		}
	}
	return nil
}

func (hd *SDLMgr_HIDDevice) Version() uint16 {
	if hd.Native_Backend_Device != nil {
		return hd.Native_Backend_Device.DeviceInfo.Release
	}
	if hd.Go_Backend_Device != nil {
		return hd.Go_Backend_Device.Version()
	}
	return 0
}

func (hd *SDLMgr_HIDDevice) Serial() string {
	if hd.Native_Backend_Device != nil {
		return hd.Native_Backend_Device.DeviceInfo.Serial
	}
	if hd.Go_Backend_Device != nil {
		return hd.Go_Backend_Device.SerialNumber()
	}
	return ""
}

func (hd *SDLMgr_HIDDevice) Open() error {
	if hd.Native_Backend_Device != nil {
		if hd.Native_Backend_Device.IsOpen() {
			return nil
		}

		return hd.Native_Backend_Device.Open()
	}

	if hd.Go_Backend_Device != nil {
		if hd.Go_Backend_Device.IsOpen() {
			return nil
		}

		err := hd.Go_Backend_Device.Open(false)
		if errors.Is(err, usbhid.ErrDeviceIsOpen) {
			return nil
		}
		return err
	}

	return fmt.Errorf("no device available")
}

func (hd *SDLMgr_HIDDevice) Close() error {
	if hd.Native_Backend_Device != nil {
		return hd.Native_Backend_Device.Close()
	}

	if hd.Go_Backend_Device != nil {
		return hd.Go_Backend_Device.Close()
	}

	return fmt.Errorf("no device available")
}

func (hd *SDLMgr_HIDDevice) GetOutputReport(id byte, length uint8) ([]byte, error) {
	hd.outputReportsState.Lock.Lock()
	defer hd.outputReportsState.Lock.Unlock()

	if err := hd.Open(); err != nil {
		return nil, err
	}

	if report, has_report := hd.outputReportsState.State[id]; has_report {
		return report, nil
	}

	empty_report := make([]byte, length)
	hd.outputReportsState.State[id] = empty_report
	return hd.outputReportsState.State[id], nil
}

func (hd *SDLMgr_HIDDevice) SetOutputReport(id byte, data []byte) error {
	hd.outputReportsState.Lock.Lock()
	defer hd.outputReportsState.Lock.Unlock()

	if err := hd.Open(); err != nil {
		return err
	}

	if hd.Native_Backend_Device != nil {
		payload := append([]byte{id}, data...)
		if _, err := hd.Native_Backend_Device.Device.Write(payload); err != nil {
			return err
		}
		hd.outputReportsState.State[id] = data
		return nil
	}

	if hd.Go_Backend_Device != nil {
		if err := hd.Go_Backend_Device.SetOutputReport(id, data); err != nil {
			return err
		}
		hd.outputReportsState.State[id] = data
		return nil
	}

	return fmt.Errorf("no valid device backend to set output report to")
}

func (hd *SDLMgr_HIDDevice) ReadFeatureReport(id byte, length uint8) ([]byte, error) {
	if err := hd.Open(); err != nil {
		return nil, err
	}

	if hd.Native_Backend_Device != nil {
		var report []byte = make([]byte, length)
		if _, err := hd.Native_Backend_Device.Device.GetFeatureReport(report); err != nil {
			return nil, err
		}
		return report[1:], nil
	}

	if hd.Go_Backend_Device != nil {
		report, err := hd.Go_Backend_Device.GetFeatureReport(id)
		if err != nil {
			return nil, err
		}
		return report, nil
	}

	return nil, fmt.Errorf("no hid device available")
}

func (hd *SDLMgr_HIDDevice) SendFeatureReport(id byte, data []byte) error {
	if err := hd.Open(); err != nil {
		return err
	}

	if hd.Native_Backend_Device != nil {
		var report []byte = []byte{id}
		report = append(report, data...)
		if _, err := hd.Native_Backend_Device.Device.SendFeatureReport(report); err != nil {
			return err
		}
		return nil
	}

	if hd.Go_Backend_Device != nil {
		err := hd.Go_Backend_Device.SetFeatureReport(id, data)
		if err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("no hid device available")
}
