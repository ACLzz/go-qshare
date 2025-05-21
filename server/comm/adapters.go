package comm

import (
	"fmt"
	"math/rand/v2"

	pbConnections "github.com/ACLzz/go-qshare/protobuf/gen/connections"
	pbSecuregcm "github.com/ACLzz/go-qshare/protobuf/gen/securegcm"
	pbSecureMessage "github.com/ACLzz/go-qshare/protobuf/gen/securemessage"
	pbSharing "github.com/ACLzz/go-qshare/protobuf/gen/sharing"
	"google.golang.org/protobuf/proto"
)

func (cc *commConn) writeError(t pbSecuregcm.Ukey2Alert_AlertType, msg string) error {
	alertMsg, err := proto.Marshal(&pbSecuregcm.Ukey2Alert{
		Type:         t.Enum(),
		ErrorMessage: proto.String(msg),
	})
	if err != nil {
		return fmt.Errorf("marshal alert message: %w", err)
	}

	if err := cc.writeUKEYMessage(pbSecuregcm.Ukey2Message_ALERT, alertMsg); err != nil {
		return fmt.Errorf("write UKEY message: %w", err)
	}

	return nil
}

func (cc *commConn) writeOfflineFrame(frame *pbConnections.V1Frame) error {
	offlineFrame, err := marshalOfflineFrame(frame)
	if err != nil {
		return err
	}

	if err = cc.writeMessage(offlineFrame); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return nil
}

func (cc *commConn) writeUKEYMessage(t pbSecuregcm.Ukey2Message_Type, msg []byte) error {
	ukeyMsg, err := proto.Marshal(&pbSecuregcm.Ukey2Message{
		MessageType: t.Enum(),
		MessageData: msg,
	})
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	if err = cc.writeMessage(ukeyMsg); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return nil
}

func (cc *commConn) writeSecureFrame(frame *pbSharing.V1Frame) error {
	data, err := proto.Marshal(&pbSharing.Frame{
		Version: pbSharing.Frame_V1.Enum(),
		V1:      frame,
	})
	if err != nil {
		return fmt.Errorf("marshal secure frame: %w", err)
	}

	messageID := rand.Int64()
	// message
	if err = cc.encryptAndWrite(&pbConnections.V1Frame{
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
	if err = cc.encryptAndWrite(&pbConnections.V1Frame{
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

// TODO: maybe only pass pbSharing.V1Frame and parametrize through args
func (cc *commConn) encryptAndWrite(payload *pbConnections.V1Frame) error {
	offlineFrame, err := marshalOfflineFrame(payload)
	if err != nil {
		return fmt.Errorf("marshal offline frame: %w", err)
	}

	cc.seqNumber++
	hb, err := cc.cipher.Encrypt(&pbSecuregcm.DeviceToDeviceMessage{
		SequenceNumber: proto.Int32(cc.seqNumber),
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
		Signature:     cc.cipher.Sign(data),
	})
	if err != nil {
		return fmt.Errorf("marshal secure message: %w", err)
	}

	if err := cc.writeMessage(data); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return nil
}

func marshalOfflineFrame(frame *pbConnections.V1Frame) ([]byte, error) {
	offlineFrame, err := proto.Marshal(&pbConnections.OfflineFrame{
		Version: pbConnections.OfflineFrame_V1.Enum(),
		V1:      frame,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal message: %w", err)
	}

	return offlineFrame, nil
}
