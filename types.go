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

type (
	TextMeta struct {
		Type  TextType
		Title string
		Size  int64
	}
	TextPayload struct {
		Meta TextMeta
		Text string
	}
)

type FileType uint8

const (
	FileUnknown FileType = iota
	FileImage
	FileVideo
	FileApp
	FileAudio
)

type (
	FileMeta struct {
		Type     FileType
		Title    string
		MimeType string
		Size     int64
	}
	FilePayload struct {
		Meta FileMeta
		Pr   *io.PipeReader
	}
)

type (
	TextCallback func(payload TextPayload)
	FileCallback func(payload FilePayload)
	AuthCallback func(text *TextMeta, files []FileMeta, pin string) bool
)
