package payloads

import (
	"io"

	qshare "github.com/ACLzz/go-qshare"
	pbSharing "github.com/ACLzz/go-qshare/protobuf/gen/sharing"
)

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

func MapTextPayload(payload *pbSharing.TextMetadata) *qshare.TextPayload {
	var textType qshare.TextType
	switch payload.GetType() {
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
		Title: payload.GetTextTitle(),
		Size:  payload.GetSize(),
	}
}

func MapFilePayloads(payloads []*pbSharing.FileMetadata) map[int64]*FilePayload {
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
