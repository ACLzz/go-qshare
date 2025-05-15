package comm

import (
	"fmt"

	pbConnections "github.com/ACLzz/go-qshare/protobuf/gen/connections"
	pbSecuregcm "github.com/ACLzz/go-qshare/protobuf/gen/securegcm"
	pbSharing "github.com/ACLzz/go-qshare/protobuf/gen/sharing"

	"google.golang.org/protobuf/proto"
)

func (cc *commConn) route(msg []byte) (err error) {
	// all messages in pairing phase are encrypted
	if cc.phase == pairingPhase {
		msg, err = cc.cipher.Decrypt(msg)
		if err != nil {
			return fmt.Errorf("decrypt message: %w", err)
		}
	}

	var (
		offFrame    pbConnections.OfflineFrame
		ukeyMessage pbSecuregcm.Ukey2Message
		secFrame    pbSharing.Frame
		nextMessage messageType
	)

	if cc.phase == initConnPhase {
		switch cc.nextExpectedMessage {
		case conn_request:
			if err = proto.Unmarshal(msg, &offFrame); err == nil {
				err = cc.processConnRequest(offFrame.GetV1().GetConnectionRequest())
				nextMessage = client_init
			}
		case client_init:
			if err = proto.Unmarshal(msg, &ukeyMessage); err == nil {
				err = cc.processUKEYInitMessage(&ukeyMessage, msg)
				nextMessage = client_finish
			}
		case client_finish:
			if err = proto.Unmarshal(msg, &ukeyMessage); err == nil {
				err = cc.processUKEYFinishedMessage(&ukeyMessage)
				nextMessage = conn_response
			}
		case conn_response:
			if err = proto.Unmarshal(msg, &offFrame); err == nil {
				err = cc.processConnResponse(offFrame.GetV1().GetConnectionResponse())
				nextMessage = paired_key_encryption
			}
		case unknown_message_type:
			err = ErrInternalError
		}
	} else if cc.phase == pairingPhase {
		if err = proto.Unmarshal(msg, &secFrame); err == nil {
			switch cc.nextExpectedMessage {
			case paired_key_encryption:
				err = cc.processPairedKeyEncryption(secFrame.GetV1().GetPairedKeyEncryption())
				nextMessage = paired_key_result
			case paired_key_result:
				err = cc.processPairedKeyResult(secFrame.GetV1().GetPairedKeyResult())
				nextMessage = introduction
			case unknown_message_type:
				err = ErrInternalError
			}
		}
	} else if cc.phase == transferPhase {
		switch cc.nextExpectedMessage {
		case introduction:
			cc.log.Debug("introduction phase achived") // TODO: remove
		case unknown_message_type:
			err = ErrInternalError
		}
	}

	if err != nil {
		proto.Unmarshal(msg, &secFrame)
		proto.Unmarshal(msg, &ukeyMessage)
		proto.Unmarshal(msg, &offFrame)
		// check if message was actually keepalive message or ukey alert
		if innerErr := cc.processKeepAliveMessage(msg); innerErr == nil {
			// do nothing
		} else if innerErr := cc.processUKEYAlert(msg); innerErr == nil {
			// do nothing
		} else {
			return err
		}
	}

	cc.nextExpectedMessage = nextMessage
	return nil
}
