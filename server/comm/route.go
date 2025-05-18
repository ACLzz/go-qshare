package comm

import (
	"errors"
	"fmt"

	pbConnections "github.com/ACLzz/go-qshare/protobuf/gen/connections"
	pbSecuregcm "github.com/ACLzz/go-qshare/protobuf/gen/securegcm"
	pbSecureMessage "github.com/ACLzz/go-qshare/protobuf/gen/securemessage"
	pbSharing "github.com/ACLzz/go-qshare/protobuf/gen/sharing"

	"google.golang.org/protobuf/proto"
)

func (cc *commConn) route(msg []byte) (err error) {
	nextMessage := cc.nextExpectedMessage
	offFrame, ukeyMessage, sharingFrame, err := cc.unmarshalInboundMessage(msg)
	if transferStatus := cc.processOngoingTransfer(offFrame); transferStatus != nil {
		if transferStatus == ErrTransferNotFinishedYet {
			return nil
		}
		if transferStatus != ErrNotTransferMessage {
			err = transferStatus
		}
	}

	if err == nil {
		if cc.phase == initConnPhase {
			switch cc.nextExpectedMessage {
			// connection request
			case conn_request:
				err = cc.processConnRequest(offFrame.GetV1().GetConnectionRequest())
				nextMessage = client_init
			// client init
			case client_init:
				err = cc.processUKEYInitMessage(ukeyMessage, msg)
				nextMessage = client_finish
			// client finish
			case client_finish:
				err = cc.processUKEYFinishedMessage(ukeyMessage)
				nextMessage = conn_response
			// connection response
			case conn_response:
				err = cc.processConnResponse(offFrame.GetV1().GetConnectionResponse())
				nextMessage = paired_key_encryption
			}
		} else if cc.phase == pairingPhase {
			switch cc.nextExpectedMessage {
			// paired key encryption
			case paired_key_encryption:
				err = cc.processPairedKeyEncryption()
				nextMessage = paired_key_result
				cc.log.Debug("written paired key encryption")
			// paired key result
			case paired_key_result:
				err = cc.processPairedKeyResult()
				nextMessage = introduction
			}
		} else if cc.phase == transferPhase {
			switch cc.nextExpectedMessage {
			// transfer introduction
			case introduction:
				err = cc.processIntroduction(sharingFrame.GetV1().GetIntroduction())
				nextMessage = accept_reject
			// accept or reject response
			case accept_reject:
				cc.log.Warn("accept reject achived")
			}
		}
	}
	offFrame.GetV1().GetPairedKeyEncryption()

	if err != nil {
		// check if message was actually service message
		if innerErr := cc.processServiceMessage(msg); innerErr != nil {
			if errors.Is(innerErr, ErrConnWasEndedByClient) {
				return innerErr
			}
			return err
		}
		return nil
	}

	cc.nextExpectedMessage = nextMessage
	return nil
}

func (cc *commConn) unmarshalInboundMessage(msg []byte) (*pbConnections.OfflineFrame, *pbSecuregcm.Ukey2Message, *pbSharing.Frame, error) {
	var (
		sharingFrame pbSharing.Frame
		offFrame     pbConnections.OfflineFrame
		ukeyMessage  pbSecuregcm.Ukey2Message
		err          error
	)

	// all messages after init connection phase are encrypted
	if cc.phase > initConnPhase {
		msg, err = cc.decryptMessage(msg)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	// understand which message now must be handeled and unmarshal to correct type
	switch cc.nextExpectedMessage {
	case conn_request, conn_response, paired_key_encryption:
		if err := proto.Unmarshal(msg, &offFrame); err != nil {
			return nil, nil, nil, fmt.Errorf("decrypt message: %w", err)
		}
		if cc.nextExpectedMessage == paired_key_encryption {
			var a pbSecuregcm.Ukey2Message
			var b pbSecuregcm.Ukey2Alert
			proto.Unmarshal(offFrame.GetV1().GetPayloadTransfer().GetPayloadChunk().GetBody(), &a)
			proto.Unmarshal(a.GetMessageData(), &b)
			// proto.Unmarshal(a.GetPayloadChunk().GetBody(), &offFrame)
			_ = offFrame
		}
		return &offFrame, nil, nil, nil
	case client_init, client_finish:
		if err := proto.Unmarshal(msg, &ukeyMessage); err != nil {
			return nil, nil, nil, fmt.Errorf("unmarshal ukey message: %w", err)
		}
		return nil, &ukeyMessage, nil, nil
	case paired_key_result, introduction:
		proto.Unmarshal(msg, &offFrame)
		proto.Unmarshal(msg, &ukeyMessage)
		if err := proto.Unmarshal(msg, &sharingFrame); err != nil {
			return nil, nil, &sharingFrame, nil
		}
	case accept_reject:
		return nil, nil, nil, nil
	}

	return nil, nil, nil, ErrInternalError
}

func (cc *commConn) processOngoingTransfer(frame *pbConnections.OfflineFrame) error {
	if frame == nil || frame.GetV1().GetPayloadTransfer() == nil {
		return ErrNotTransferMessage
	}

	chunk := frame.GetV1().GetPayloadTransfer().GetPayloadChunk()
	cc.buf = append(cc.buf, chunk.GetBody()...)
	if chunk.GetFlags()&1 == 1 {
		if len(cc.buf) != int(frame.GetV1().GetPayloadTransfer().GetPayloadHeader().GetTotalSize()) {
			return ErrInvalidMessage
		}
		return nil
	}

	return ErrTransferNotFinishedYet
}

func (cc *commConn) processServiceMessage(msg []byte) error {
	var (
		offFrame    pbConnections.OfflineFrame
		ukeyMessage pbSecuregcm.Ukey2Message
		err         error
	)

	// all messages after init connection phase are encrypted
	if cc.phase > initConnPhase {
		msg, err = cc.decryptMessage(msg)
		if err != nil {
			return fmt.Errorf("decrypt message: %w", err)
		}
	}

	// don't care about errors here, because we don't know actual message type
	_ = proto.Unmarshal(msg, &offFrame)
	_ = proto.Unmarshal(msg, &ukeyMessage)

	isKeepAlive := offFrame.GetV1().GetType() == pbConnections.V1Frame_KEEP_ALIVE
	isDisconnection := offFrame.GetV1().GetType() == pbConnections.V1Frame_DISCONNECTION
	isUKEYAlert := ukeyMessage.GetMessageType() == pbSecuregcm.Ukey2Message_ALERT
	if isKeepAlive {
		cc.processKeepAliveMessage(&offFrame)
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
