package adapter

import (
	pbConnections "github.com/ACLzz/go-qshare/internal/protobuf/gen/connections"
	pbSecuregcm "github.com/ACLzz/go-qshare/internal/protobuf/gen/securegcm"
	"google.golang.org/protobuf/proto"
)

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
		err = a.writeOfflineFrame(&keepAliveFrame)
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
