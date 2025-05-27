package goqshare

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
