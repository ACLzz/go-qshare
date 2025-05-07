package server

import "errors"

var (
	ErrInvalidPort       = errors.New("invalid port value")
	ErrInvalidAdapter    = errors.New("invalid adapter")
	ErrInvalidEndpoint   = errors.New("invalid endpoint")
	ErrInvalidDeviceType = errors.New("invalid device type")
)

var (
	ErrInvalidMessage         = errors.New("got invalid message from client")
	ErrInvalidHandshakeCipher = errors.New("no supported handshake ciphers was found")
)
