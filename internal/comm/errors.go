package comm

import "errors"

var (
	ErrConnWasEndedByClient = errors.New("connection was ended by client")
	ErrNotServiceMessage    = errors.New("message is not service message")
)

var (
	ErrInvalidSignatureScheme  = errors.New("got unsupported signature scheme")
	ErrInvalidIV               = errors.New("got invalid iv in header")
	ErrInvalidEncryptionScheme = errors.New("got unsupported encryption scheme")
)
