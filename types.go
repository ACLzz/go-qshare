package goqshare

import "io"

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

type TextType uint8

const (
	TextUnknown TextType = iota
	TextText
	TextURL
	TextAddress
	TextPhoneNumber
)

type TextPayload struct {
	Type  TextType
	Title string
	Size  int64
}

type FileType uint8

const (
	FileUnknown FileType = iota
	FileImage
	FileVideo
	FileApp
	FileAudio
)

type FilePayload struct {
	Type     FileType
	Title    string
	MimeType string
	Size     int64
}

type (
	TextCallback func(meta TextPayload, text string)
	FileCallback func(meta FilePayload, pr *io.PipeReader)
	AuthCallback func(text *TextPayload, files []FilePayload, pin string) bool
)
