package listener

import "errors"

var (
	ErrInternalError           = errors.New("internal error")
	ErrTextTransferNotExpected = errors.New("text transfer was not expected")
)
