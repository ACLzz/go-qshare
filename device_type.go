package goqshare

type DeviceType uint

const (
	UnknownDevice = iota
	PhoneDevice
	TabletDevice
	LaptopDevice
)

func (t DeviceType) isValid() bool {
	return t <= LaptopDevice
}
