package comm

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	pbConnections "github.com/ACLzz/go-qshare/protobuf/gen/connections"
	pbSecuregcm "github.com/ACLzz/go-qshare/protobuf/gen/securegcm"
	pbSecureMessage "github.com/ACLzz/go-qshare/protobuf/gen/securemessage"
)

func (cc *commConn) processServiceMessage(msg []byte) error {
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
		cc.processKeepAliveMessage()
	}
	if isUKEYAlert {
		cc.processUKEYAlert(&ukeyMessage)
	}
	if isDisconnection {
		return ErrConnWasEndedByClient
	}

	if !isKeepAlive && !isUKEYAlert {
		return ErrNotServiceMessage
	}

	return nil
}

func (cc *commConn) processKeepAliveMessage() {
	keepAliveFrame := pbConnections.V1Frame{
		Type: pbConnections.V1Frame_KEEP_ALIVE.Enum(),
		KeepAlive: &pbConnections.KeepAliveFrame{
			Ack: proto.Bool(true),
		},
	}

	var err error
	if cc.phase > init_phase {
		err = cc.encryptAndWrite(&keepAliveFrame)
	} else {
		err = cc.writeOfflineFrame(&keepAliveFrame)
	}
	if err != nil {
		cc.log.Error("send keep alive message", err)
		return
	}

	cc.log.Debug("sent keep alive message")
}

func (cc *commConn) processUKEYAlert(ukeyMessage *pbSecuregcm.Ukey2Message) {
	var alert pbSecuregcm.Ukey2Alert
	if err := proto.Unmarshal(ukeyMessage.GetMessageData(), &alert); err != nil {
		return
	}

	alertType, ok := pbSecuregcm.Ukey2Alert_AlertType_name[int32(alert.GetType())]
	if !ok {
		return
	}

	cc.log.Warn("got an alert", "type", alertType, "message", alert.GetErrorMessage())
}

func (cc *commConn) decryptMessage(msg []byte) ([]byte, error) {
	// unmarshal secure message
	var secMsg pbSecureMessage.SecureMessage
	if err := proto.Unmarshal(msg, &secMsg); err != nil {
		return nil, fmt.Errorf("unmarshal secure message: %w", err)
	}
	if err := cc.cipher.ValidateSignature(secMsg.GetHeaderAndBody(), secMsg.GetSignature()); err != nil {
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
	d2dMsg, err := cc.cipher.Decrypt(&hb)
	if err != nil {
		return nil, fmt.Errorf("decrypt message: %w", err)
	}

	return d2dMsg.GetMessage(), nil
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
