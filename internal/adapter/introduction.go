package adapter

import (
	"io"

	qshare "github.com/ACLzz/go-qshare"
	pbSharing "github.com/ACLzz/go-qshare/internal/protobuf/gen/sharing"
)

type IntroductionFrame struct {
	Text  *qshare.TextPayload
	Files map[int64]*FilePayload // TODO: should it be map of pointers?
}

func (f IntroductionFrame) HasText() bool {
	return f.Text != nil
}

func (f IntroductionFrame) HasFiles() bool {
	return len(f.Files) > 0
}

func (a *Adapter) UnmarshalIntroduction(msg []byte) (IntroductionFrame, error) {
	frame, err := unmarshalSharingFrame(msg)
	if err != nil {
		return IntroductionFrame{}, err
	}

	if frame.GetType() != pbSharing.V1Frame_INTRODUCTION || frame.GetIntroduction() == nil {
		return IntroductionFrame{}, ErrInvalidMessage
	}

	// validate that at least one text or file are sent
	text := frame.GetIntroduction().GetTextMetadata()
	files := frame.GetIntroduction().GetFileMetadata()
	if len(files) == 0 && len(text) == 0 {
		return IntroductionFrame{}, ErrInvalidMessage
	}

	return IntroductionFrame{
		Text:  mapTextPayload(text),
		Files: mapFilePayloads(files),
	}, nil
}

type FilePayload struct {
	qshare.FilePayload
	IsNotified bool
	BytesSent  int64
	Pr         *io.PipeReader
	Pw         *io.PipeWriter
}

func newFilePayload(t qshare.FileType, title, mimeType string, payloadID, size int64) *FilePayload {
	pr, pw := io.Pipe()
	return &FilePayload{
		FilePayload: qshare.FilePayload{
			Type:     t,
			Title:    title,
			MimeType: mimeType,
			Size:     size,
		},
		IsNotified: false,
		Pr:         pr,
		Pw:         pw,
	}
}

func mapTextPayload(payloads []*pbSharing.TextMetadata) *qshare.TextPayload {
	if len(payloads) == 0 {
		return nil
	}

	var textType qshare.TextType
	switch payloads[0].GetType() {
	case pbSharing.TextMetadata_TEXT:
		textType = qshare.TextText
	case pbSharing.TextMetadata_URL:
		textType = qshare.TextURL
	case pbSharing.TextMetadata_ADDRESS:
		textType = qshare.TextAddress
	case pbSharing.TextMetadata_PHONE_NUMBER:
		textType = qshare.TextPhoneNumber
	default:
		textType = qshare.TextUnknown
	}

	return &qshare.TextPayload{
		Type:  textType,
		Title: payloads[0].GetTextTitle(),
		Size:  payloads[0].GetSize(),
	}
}

func mapFilePayloads(payloads []*pbSharing.FileMetadata) map[int64]*FilePayload {
	var fileType qshare.FileType
	mappedPayloads := make(map[int64]*FilePayload, len(payloads))

	for i := range payloads {
		switch payloads[i].GetType() {
		case pbSharing.FileMetadata_IMAGE:
			fileType = qshare.FileImage
		case pbSharing.FileMetadata_VIDEO:
			fileType = qshare.FileVideo
		case pbSharing.FileMetadata_APP:
			fileType = qshare.FileApp
		case pbSharing.FileMetadata_AUDIO:
			fileType = qshare.FileAudio
		default:
			fileType = qshare.FileUnknown
		}

		mappedPayloads[payloads[i].GetPayloadId()] = newFilePayload(
			fileType,
			payloads[i].GetName(),
			payloads[i].GetMimeType(),
			payloads[i].GetPayloadId(),
			payloads[i].GetSize(),
		)
	}

	return mappedPayloads
}
