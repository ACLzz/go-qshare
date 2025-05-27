package goqshare

import "io"

type (
	TextCallback func(meta TextPayload, text string)
	FileCallback func(meta FilePayload, pr *io.PipeReader)
)
