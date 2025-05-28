package client

import "errors"

var (
	ErrInvalidAdapter    = errors.New("invalid adapter")
	ErrInvalidDeviceType = errors.New("invalid device type")
)

var (
	ErrInvalidServerInstance = errors.New("invalid server instance")
)
