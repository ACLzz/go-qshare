package qclient

import "errors"

var (
	ErrInvalidAdapter    = errors.New("invalid adapter")
	ErrInvalidRandom     = errors.New("invalid random")
	ErrInvalidDeviceType = errors.New("invalid device type")
)

var (
	ErrInvalidServerInstance = errors.New("invalid server instance")
	ErrInvalidTextType       = errors.New("text type cannot be unknown")
)
