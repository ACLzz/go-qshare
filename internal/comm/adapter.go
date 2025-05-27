package comm

import (
	"encoding/binary"
	"fmt"
	"math/rand/v2"
	"net"

	"github.com/ACLzz/go-qshare/internal/crypt"
	pbConnections "github.com/ACLzz/go-qshare/internal/protobuf/gen/connections"
	pbSecuregcm "github.com/ACLzz/go-qshare/internal/protobuf/gen/securegcm"
	pbSecureMessage "github.com/ACLzz/go-qshare/internal/protobuf/gen/securemessage"
	pbSharing "github.com/ACLzz/go-qshare/internal/protobuf/gen/sharing"
	"github.com/ACLzz/go-qshare/log"
	"google.golang.org/protobuf/proto"
)

type Adapter struct {
	conn   net.Conn
	cipher *crypt.Cipher
	log    log.Logger

	seqNumber   int32
	isEncrypted bool
}

func NewAdapter(conn net.Conn, logger log.Logger, cipher *crypt.Cipher) Adapter {
	return Adapter{
		conn:   conn,
		cipher: cipher,
		log:    logger,
	}
}

func (a *Adapter) SetIsEncrypted(isEncrypted bool) {
	a.isEncrypted = isEncrypted
}

func (a *Adapter) writeMessage(data []byte) error {
	msgLength := uint32(len(data))
	msgWithPrefix := make([]byte, 4, msgLength+4)

	binary.BigEndian.PutUint32(msgWithPrefix[:4], msgLength)
	copy(msgWithPrefix[4:], data)
	if _, err := a.conn.Write(msgWithPrefix); err != nil {
		return fmt.Errorf("send message to client: %w", err)
	}

	return nil
}

func (a *Adapter) WriteOfflineFrame(frame *pbConnections.V1Frame) error {
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

func (a *Adapter) WriteUKEYMessage(t pbSecuregcm.Ukey2Message_Type, msg []byte) error {
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

func (a *Adapter) WriteSecureFrame(frame *pbSharing.V1Frame) error {
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

func (a *Adapter) WriteError(t pbSecuregcm.Ukey2Alert_AlertType, msg string) error {
	alertMsg, err := proto.Marshal(&pbSecuregcm.Ukey2Alert{
		Type:         t.Enum(),
		ErrorMessage: proto.String(msg),
	})
	if err != nil {
		return fmt.Errorf("marshal alert message: %w", err)
	}

	if err := a.WriteUKEYMessage(pbSecuregcm.Ukey2Message_ALERT, alertMsg); err != nil {
		return fmt.Errorf("write UKEY message: %w", err)
	}

	return nil
}

func (a *Adapter) ProcessServiceMessage(msg []byte) error {
	var (
		offFrame    pbConnections.OfflineFrame
		ukeyMessage pbSecuregcm.Ukey2Message
	)

	// don't care about errors here, because we don't know actual message type
	_ = proto.Unmarshal(msg, &offFrame)
	_ = proto.Unmarshal(msg, &ukeyMessage)

	isKeepAlive := offFrame.GetV1().GetType() == pbConnections.V1Frame_KEEP_ALIVE
	isDisconnection := offFrame.GetV1().GetType() == pbConnections.V1Frame_DISCONNECTION
	isUKEYAlert := ukeyMessage.GetMessageType() == pbSecuregcm.Ukey2Message_ALERT
	if isKeepAlive {
		a.sendKeepAliveMessage()
	}
	if isUKEYAlert {
		a.logUKEYAlert(&ukeyMessage)
	}
	if isDisconnection {
		return ErrConnWasEndedByClient
	}

	if !isKeepAlive && !isUKEYAlert {
		return ErrNotServiceMessage
	}

	return nil
}

func (a *Adapter) DecryptMessage(msg []byte) ([]byte, error) {
	// unmarshal secure message
	var secMsg pbSecureMessage.SecureMessage
	if err := proto.Unmarshal(msg, &secMsg); err != nil {
		return nil, fmt.Errorf("unmarshal secure message: %w", err)
	}
	if err := a.cipher.ValidateSignature(secMsg.GetHeaderAndBody(), secMsg.GetSignature()); err != nil {
		return nil, err
	}

	// unmarshal header and body
	var hb pbSecureMessage.HeaderAndBody
	if err := proto.Unmarshal(secMsg.GetHeaderAndBody(), &hb); err != nil {
		return nil, fmt.Errorf("unmarshal secure message: %w", err)
	}
	if err := validateHeaderAndBody(&hb); err != nil {
		return nil, err
	}

	// decrypt header and body into device to device message
	d2dMsg, err := a.cipher.Decrypt(&hb)
	if err != nil {
		return nil, fmt.Errorf("decrypt message: %w", err)
	}

	return d2dMsg.GetMessage(), nil
}

func (a *Adapter) Disconnect() {
	if err := a.WriteOfflineFrame(&pbConnections.V1Frame{
		Type: pbConnections.V1Frame_DISCONNECTION.Enum(),
		Disconnection: &pbConnections.DisconnectionFrame{
			AckSafeToDisconnect: proto.Bool(true),
		},
	}); err != nil {
		a.log.Error("error while disconnecting", err)
	}
}

func validateHeaderAndBody(hb *pbSecureMessage.HeaderAndBody) error {
	if hb.GetHeader().GetEncryptionScheme() != pbSecureMessage.EncScheme_AES_256_CBC {
		return ErrInvalidEncryptionScheme
	}
	if hb.GetHeader().GetSignatureScheme() != pbSecureMessage.SigScheme_HMAC_SHA256 {
		return ErrInvalidSignatureScheme
	}
	if len(hb.GetHeader().GetIv()) < 16 {
		return ErrInvalidIV
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

func (a *Adapter) sendKeepAliveMessage() {
	keepAliveFrame := pbConnections.V1Frame{
		Type: pbConnections.V1Frame_KEEP_ALIVE.Enum(),
		KeepAlive: &pbConnections.KeepAliveFrame{
			Ack: proto.Bool(true),
		},
	}

	var err error
	if a.isEncrypted {
		err = a.encryptAndWrite(&keepAliveFrame)
	} else {
		err = a.WriteOfflineFrame(&keepAliveFrame)
	}
	if err != nil {
		a.log.Error("send keep alive message", err)
		return
	}

	a.log.Debug("sent keep alive message")
}

func (a *Adapter) logUKEYAlert(ukeyMessage *pbSecuregcm.Ukey2Message) {
	var alert pbSecuregcm.Ukey2Alert
	if err := proto.Unmarshal(ukeyMessage.GetMessageData(), &alert); err != nil {
		return
	}

	alertType, ok := pbSecuregcm.Ukey2Alert_AlertType_name[int32(alert.GetType())]
	if !ok {
		return
	}

	a.log.Warn("got an alert", "type", alertType, "message", alert.GetErrorMessage())
}
