package adapter

import (
	"encoding/binary"
	"fmt"
	"math/rand/v2"

	pbConnections "github.com/ACLzz/go-qshare/internal/protobuf/gen/connections"
	pbSecuregcm "github.com/ACLzz/go-qshare/internal/protobuf/gen/securegcm"
	pbSecureMessage "github.com/ACLzz/go-qshare/internal/protobuf/gen/securemessage"
	pbSharing "github.com/ACLzz/go-qshare/internal/protobuf/gen/sharing"
	"google.golang.org/protobuf/proto"
)

func (a *Adapter) writeMessage(data []byte) error {
	msgLength := uint32(len(data))
	msgWithPrefix := make([]byte, msgLength+4)

	binary.BigEndian.PutUint32(msgWithPrefix[:4], msgLength)
	copy(msgWithPrefix[4:], data)
	if _, err := a.conn.Write(msgWithPrefix); err != nil {
		return fmt.Errorf("send message to client: %w", err)
	}

	return nil
}

func (a *Adapter) writeOfflineFrame(frame *pbConnections.V1Frame) error {
	offlineFrame, err := proto.Marshal(&pbConnections.OfflineFrame{
		Version: pbConnections.OfflineFrame_V1.Enum(),
		V1:      frame,
	})
	if err != nil {
		return err
	}

	if err = a.writeMessage(offlineFrame); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return nil
}

func (a *Adapter) writeUKEYMessage(t pbSecuregcm.Ukey2Message_Type, msg []byte) error {
	ukeyMsg, err := proto.Marshal(&pbSecuregcm.Ukey2Message{
		MessageType: t.Enum(),
		MessageData: msg,
	})
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	if err = a.writeMessage(ukeyMsg); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return nil
}

func (a *Adapter) writeSecureFrame(frame *pbSharing.V1Frame) error {
	data, err := proto.Marshal(&pbSharing.Frame{
		Version: pbSharing.Frame_V1.Enum(),
		V1:      frame,
	})
	if err != nil {
		return fmt.Errorf("marshal secure frame: %w", err)
	}

	messageID := rand.Int64()
	// message
	if err = a.encryptAndWrite(&pbConnections.V1Frame{
		Type: pbConnections.V1Frame_PAYLOAD_TRANSFER.Enum(),
		PayloadTransfer: &pbConnections.PayloadTransferFrame{
			PacketType: pbConnections.PayloadTransferFrame_DATA.Enum(),
			PayloadChunk: &pbConnections.PayloadTransferFrame_PayloadChunk{
				Offset: proto.Int64(0),
				Flags:  proto.Int32(0),
				Body:   data,
			},
			PayloadHeader: &pbConnections.PayloadTransferFrame_PayloadHeader{
				Id:          proto.Int64(messageID),
				Type:        pbConnections.PayloadTransferFrame_PayloadHeader_BYTES.Enum(),
				TotalSize:   proto.Int64(int64(len(data))),
				IsSensitive: proto.Bool(false),
			},
		},
	}); err != nil {
		return fmt.Errorf("encrypt and write message: %w", err)
	}

	// last chunk
	if err = a.encryptAndWrite(&pbConnections.V1Frame{
		Type: pbConnections.V1Frame_PAYLOAD_TRANSFER.Enum(),
		PayloadTransfer: &pbConnections.PayloadTransferFrame{
			PacketType: pbConnections.PayloadTransferFrame_DATA.Enum(),
			PayloadChunk: &pbConnections.PayloadTransferFrame_PayloadChunk{
				Offset: proto.Int64(int64(len(data))),
				Flags:  proto.Int32(1),
				Body:   []byte{},
			},
			PayloadHeader: &pbConnections.PayloadTransferFrame_PayloadHeader{
				Id:          proto.Int64(messageID),
				Type:        pbConnections.PayloadTransferFrame_PayloadHeader_BYTES.Enum(),
				TotalSize:   proto.Int64(int64(len(data))),
				IsSensitive: proto.Bool(false),
			},
		},
	}); err != nil {
		return fmt.Errorf("encrypt and write last chunk: %w", err)
	}

	return nil
}

func (a *Adapter) encryptAndWrite(frame *pbConnections.V1Frame) error {
	offlineFrame, err := proto.Marshal(&pbConnections.OfflineFrame{
		Version: pbConnections.OfflineFrame_V1.Enum(),
		V1:      frame,
	})
	if err != nil {
		return fmt.Errorf("marshal offline frame: %w", err)
	}

	a.seqNumber++
	hb, err := a.cipher.Encrypt(&pbSecuregcm.DeviceToDeviceMessage{
		SequenceNumber: proto.Int32(a.seqNumber),
		Message:        offlineFrame,
	})
	if err != nil {
		return fmt.Errorf("encrypt data: %w", err)
	}

	data, err := proto.Marshal(hb)
	if err != nil {
		return fmt.Errorf("marshal header and body: %w", err)
	}

	data, err = proto.Marshal(&pbSecureMessage.SecureMessage{
		HeaderAndBody: data,
		Signature:     a.cipher.Sign(data),
	})
	if err != nil {
		return fmt.Errorf("marshal secure message: %w", err)
	}

	if err := a.writeMessage(data); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return nil
}

func (a *Adapter) SendBadMessageError() {
	alertMsg, err := proto.Marshal(&pbSecuregcm.Ukey2Alert{
		Type:         pbSecuregcm.Ukey2Alert_BAD_MESSAGE.Enum(),
		ErrorMessage: proto.String(ErrInvalidMessage.Error()),
	})
	if err != nil {
		a.log.Error("marshal alert message", err)
		return
	}

	if err := a.writeUKEYMessage(pbSecuregcm.Ukey2Message_ALERT, alertMsg); err != nil {
		a.log.Error("send bad message error", err)
		return
	}
}
