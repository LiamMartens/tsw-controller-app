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
	if hd.Go_Backend_Device != nil {
		return hd.Go_Backend_Device.Version()
	}
	if hd.Native_Backend_Device != nil {
		return hd.Native_Backend_Device.DeviceInfo.Release
	}
	return 0
}

func (hd *SDLMgr_HIDDevice) Serial() string {
	if hd.Go_Backend_Device != nil {
		return hd.Go_Backend_Device.SerialNumber()
	}
	if hd.Native_Backend_Device != nil {
		return hd.Native_Backend_Device.DeviceInfo.Serial
	}
	return ""
}

func (hd *SDLMgr_HIDDevice) Open() error {
	if hd.Go_Backend_Device != nil && !hd.Go_Backend_Device.IsOpen() {
		err := hd.Go_Backend_Device.Open(false)
		if errors.Is(err, usbhid.ErrDeviceIsOpen) {
			return nil
		}
		return err
	}
	if hd.Native_Backend_Device != nil && !hd.Native_Backend_Device.IsOpen() {
		return hd.Native_Backend_Device.Open()
	}
	return fmt.Errorf("no device available")
}

func (hd *SDLMgr_HIDDevice) Close() error {
	if hd.Go_Backend_Device != nil {
		return hd.Go_Backend_Device.Close()
	}
	if hd.Native_Backend_Device != nil {
		return hd.Native_Backend_Device.Open()
	}
	return fmt.Errorf("no device available")
}

func (hd *SDLMgr_HIDDevice) ReadFeatureReport(id byte) ([]byte, error) {
	if err := hd.Open(); err != nil {
		return nil, err
	}

	if hd.Go_Backend_Device != nil {
		report, err := hd.Go_Backend_Device.GetFeatureReport(id)
		if err != nil {
			return nil, err
		}
		return report, nil
	}

	if hd.Native_Backend_Device != nil {
		var report []byte = []byte{id}
		if _, err := hd.Native_Backend_Device.Device.GetFeatureReport(report); err != nil {
			return nil, err
		}
		return report[1:], nil
	}

	return nil, fmt.Errorf("no hid device available")
}

func (hd *SDLMgr_HIDDevice) SendFeatureReport(id byte, data []byte) error {
	if err := hd.Open(); err != nil {
		return err
	}

	if hd.Go_Backend_Device != nil {
		err := hd.Go_Backend_Device.SetFeatureReport(id, data)
		if err != nil {
			return err
		}
		return nil
	}

	if hd.Native_Backend_Device != nil {
		var report []byte = []byte{id}
		report = append(report, data...)
		if _, err := hd.Native_Backend_Device.Device.SendFeatureReport(report); err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("no hid device available")
}
