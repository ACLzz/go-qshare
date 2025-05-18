package comm

import "errors"

var (
	ErrInvalidMessage         = errors.New("got invalid message from client")
	ErrNotServiceMessage      = errors.New("message is not service message")
	ErrInvalidHandshakeCipher = errors.New("no supported handshake ciphers was found")
	ErrConnWasEndedByClient   = errors.New("connection was ended by client")
	ErrInternalError          = errors.New("internal error")
)

var (
	ErrInvalidSignatureScheme  = errors.New("got unsupported signature scheme")
	ErrInvalidIV               = errors.New("got invalid iv in header")
	ErrInvalidEncryptionScheme = errors.New("got unsupported encryption scheme")
)

var (
	ErrTransferNotFinishedYet = errors.New("transfer was not finished yet")
	ErrNotTransferMessage     = errors.New("not transfer message")
)
