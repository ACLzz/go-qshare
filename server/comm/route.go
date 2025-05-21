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
	// all messages after init connection phase are encrypted
	if cc.phase > initConnPhase {
		msg, err = cc.decryptMessage(msg)
		if err != nil {
			return err
		}
	}

	if status, err := cc.processOngoingTransfer(msg); status > transfer_not_started {
		switch status {
		case transfer_in_progress:
			return nil
		case transfer_finished:
			defer cc.clearBuf()
		case transfer_error:
			if err != nil {
				return fmt.Errorf("process ongoing transfer: %w", err)
			}
		}
	}

	nextMessage := cc.nextExpectedMessage
	offFrame, ukeyMessage, sharingFrame, unmarshalErr := cc.unmarshalInboundMessage(msg)
	if unmarshalErr != nil {
		if err == nil {
			err = fmt.Errorf("unmarshal inbound message: %w", unmarshalErr)
		}
	}

	if err == nil {
		switch cc.phase {
		case initConnPhase:
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
		case pairingPhase:
			switch cc.nextExpectedMessage {
			// paired key encryption
			case paired_key_encryption:
				err = cc.processPairedKeyEncryption(sharingFrame.GetV1().GetPairedKeyEncryption())
				nextMessage = paired_key_result
				cc.log.Debug("written paired key encryption")
			// paired key result
			case paired_key_result:
				err = cc.processPairedKeyResult(sharingFrame.GetV1().GetPairedKeyResult())
				nextMessage = introduction
			}
		case transferPhase:
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

	// understand which message now must be handeled and unmarshal to correct type
	switch cc.nextExpectedMessage {
	case conn_request, conn_response:
		if err = proto.Unmarshal(msg, &offFrame); err != nil {
			return nil, nil, nil, fmt.Errorf("unmarshal offline frame: %w", err)
		}
		return &offFrame, nil, nil, nil
	case client_init, client_finish:
		if err = proto.Unmarshal(msg, &ukeyMessage); err != nil {
			return nil, nil, nil, fmt.Errorf("unmarshal ukey message: %w", err)
		}
		return nil, &ukeyMessage, nil, nil
	case paired_key_result, paired_key_encryption, introduction:
		if err = proto.Unmarshal(cc.buf, &sharingFrame); err != nil {
			return nil, nil, nil, fmt.Errorf("unmarshal sharing frame: %w", err)
		}
		return nil, nil, &sharingFrame, nil
	case accept_reject:
		return nil, nil, nil, nil
	}

	return nil, nil, nil, ErrInternalError
}

func (cc *commConn) processOngoingTransfer(msg []byte) (transferProgress, error) {
	var frame pbConnections.OfflineFrame
	if err := proto.Unmarshal(msg, &frame); err != nil {
		return transfer_not_started, nil
	}

	payloadTransfer := frame.GetV1().GetPayloadTransfer()
	if payloadTransfer == nil {
		return transfer_not_started, nil
	}

	chunk := payloadTransfer.GetPayloadChunk()
	if chunk.GetOffset() != int64(len(cc.buf)) {
		return transfer_error, ErrInvalidMessage
	}

	cc.buf = append(cc.buf, chunk.GetBody()...)
	if chunk.GetFlags()&1 == 1 {
		if payloadTransfer.GetPayloadHeader().GetTotalSize() != int64(len(cc.buf)) {
			return transfer_error, ErrInvalidMessage
		}
		return transfer_finished, nil
	}

	return transfer_in_progress, nil
}

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
