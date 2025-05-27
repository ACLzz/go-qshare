package comm

import (
	"fmt"
	"time"

	"github.com/ACLzz/go-qshare/internal/payloads"
	pbConnections "github.com/ACLzz/go-qshare/protobuf/gen/connections"
	pbSharing "github.com/ACLzz/go-qshare/protobuf/gen/sharing"
	"google.golang.org/protobuf/proto"
)

func (cc *commConn) processIntroduction(msg *pbSharing.IntroductionFrame) error {
	isText, isFile := msg.GetTextMetadata() != nil, msg.GetFileMetadata() != nil
	if !isText && !isFile {
		return ErrInvalidMessage
	}

	if isText {
		cc.textPayload = payloads.MapTextPayload(msg.GetTextMetadata()[0])
		cc.expectedPayloads++
	}
	if isFile {
		cc.filePayloads = payloads.MapFilePayloads(msg.GetFileMetadata())
		cc.expectedPayloads += len(cc.filePayloads)
	}

	return nil
}

func (cc *commConn) processTransferRequest(msg *pbSharing.ConnectionResponseFrame) error {
	if msg.GetStatus() != pbSharing.ConnectionResponseFrame_ACCEPT {
		return ErrInvalidMessage
	}

	if err := cc.writeSecureFrame(&pbSharing.V1Frame{
		Type: pbSharing.V1Frame_RESPONSE.Enum(),
		ConnectionResponse: &pbSharing.ConnectionResponseFrame{
			Status: pbSharing.ConnectionResponseFrame_ACCEPT.Enum(),
		},
	}); err != nil {
		return fmt.Errorf("write secure message: %w", err)
	}

	return nil
}

func (cc *commConn) processTransferComplete() error {
	var offFrame pbConnections.OfflineFrame
	if err := proto.Unmarshal(cc.buf, &offFrame); err == nil && offFrame.GetV1().GetType() == pbConnections.V1Frame_KEEP_ALIVE {
		cc.processKeepAliveMessage()
		return nil
	}

	fmt.Println(cc.buf)

	cc.receivedPayloads++
	if cc.receivedPayloads >= cc.expectedPayloads {
		cc.phase = disconnect_phase
		cc.log.Debug("got last payload, disconnecting...")
	}

	if cc.textPayload != nil {
		cc.textCallback(*cc.textPayload, string(cc.buf))
	}

	return nil
}

func (cc *commConn) processOngoingTransfer(msg []byte) (transferProgress, error) {
	var frame pbConnections.OfflineFrame
	if err := proto.Unmarshal(msg, &frame); err != nil {
		return transfer_not_started, nil
	}

	if frame.GetV1().GetType() != pbConnections.V1Frame_PAYLOAD_TRANSFER || frame.GetV1().GetPayloadTransfer() == nil {
		return transfer_not_started, nil
	}

	chunk := frame.GetV1().GetPayloadTransfer().GetPayloadChunk()
	header := frame.GetV1().GetPayloadTransfer().GetPayloadHeader()

	switch header.GetType() {
	case pbConnections.PayloadTransferFrame_PayloadHeader_FILE:
		if _, ok := cc.filePayloads[header.GetId()]; ok {
			return cc.writeFileChunk(header.GetId(), chunk)
		}
	case pbConnections.PayloadTransferFrame_PayloadHeader_BYTES:
		return cc.writeChunkToBuf(header, chunk)
	}

	return transfer_error, ErrInvalidMessage
}

func (cc *commConn) writeFileChunk(id int64, chunk *pbConnections.PayloadTransferFrame_PayloadChunk) (transferProgress, error) {
	file := cc.filePayloads[id]
	if !file.IsNotified {
		cc.filePayloads[id].IsNotified = true
		cc.fileCallback(file.FilePayload, file.Pr)
	}

	if chunk.GetOffset() != file.BytesSent {
		return transfer_error, ErrInvalidMessage // TODO: another error
	}

	if len(chunk.GetBody()) > 0 {
		cc.log.Debug("writing file chunk to client", "size", len(chunk.GetBody()))
		n, err := file.Pw.Write(chunk.GetBody())
		if err != nil {
			return transfer_error, fmt.Errorf("writing chunk to pipe: %w", err)
		}
		cc.log.Debug("written chunk", "size", n)

		if n != len(chunk.GetBody()) {
			return transfer_error, ErrInternalError // TODO: another error
		}

		cc.filePayloads[id].BytesSent += int64(n)
	}

	if chunk.GetFlags()&1 == 1 {
		time.Sleep(2 * time.Second)
		file.Pw.Close()
		if file.BytesSent != file.Size {
			return transfer_error, ErrInvalidMessage // TODO: another error
		}
		cc.log.Debug("file downloaded", "filename", file.Title)
		return transfer_finished, nil
	}

	return transfer_in_progress, nil
}

func (cc *commConn) writeChunkToBuf(
	header *pbConnections.PayloadTransferFrame_PayloadHeader,
	chunk *pbConnections.PayloadTransferFrame_PayloadChunk,
) (transferProgress, error) {
	if chunk.GetOffset() != int64(len(cc.buf)) {
		return transfer_error, ErrInvalidMessage
	}

	cc.buf = append(cc.buf, chunk.GetBody()...)
	if chunk.GetFlags()&1 == 1 {
		if header.GetTotalSize() != int64(len(cc.buf)) {
			return transfer_error, ErrInvalidMessage
		}
		return transfer_finished, nil
	}

	return transfer_in_progress, nil
}
