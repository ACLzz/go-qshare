package qserver

import "errors"

// builder
var (
	ErrInvalidPort       = errors.New("invalid port value")
	ErrInvalidAdapter    = errors.New("invalid adapter")
	ErrInvalidLogger     = errors.New("invalid logger")
	ErrInvalidEndpoint   = errors.New("invalid endpoint")
	ErrInvalidDeviceType = errors.New("invalid device type")
)
