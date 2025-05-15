package comm

import "errors"

var (
	ErrInvalidMessage         = errors.New("got invalid message from client")
	ErrInvalidHandshakeCipher = errors.New("no supported handshake ciphers was found")
	ErrConnWasEndedByClient   = errors.New("connection was ended by client")
	ErrInternalError          = errors.New("internal error")
)
