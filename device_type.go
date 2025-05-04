package goqshare

type DeviceType uint

const (
	UnknownDevice = iota
	PhoneDevice
	TabletDevice
	LaptopDevice
)

func (t DeviceType) IsValid() bool {
	return t <= LaptopDevice
}
