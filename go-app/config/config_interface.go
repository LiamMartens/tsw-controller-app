package config

type DeviceConfiguration interface {
	GetUsbID() string
	GetUniqueID() string
	Matches(c DeviceConfiguration) bool
}
