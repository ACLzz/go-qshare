package comm

import (
	"errors"
	"fmt"

	pbConnections "github.com/ACLzz/go-qshare/protobuf/gen/connections"
	pbSecuregcm "github.com/ACLzz/go-qshare/protobuf/gen/securegcm"
	pbSharing "github.com/ACLzz/go-qshare/protobuf/gen/sharing"

	"google.golang.org/protobuf/proto"
)

var mysteryFrame = []byte{
	0x08, 0x01,
	0x12, 0x0b,
	0x08, 0x07,
	0x3a, 0x07,
	0x0d, 0x00, 0x00, 0x00, 0x00,
	0x10, 0x01,
}

func (cc *commConn) route(msg []byte) (err error) {
	// all messages after init connection phase are encrypted
	if cc.phase > init_phase {
		msg, err = cc.decryptMessage(msg)
		if err != nil {
			return err
		}
	}

	// process transfer if it is running
	if status, err := cc.processOngoingTransfer(msg); status > transfer_not_started {
		switch status {
		case transfer_in_progress:
			return nil
		case transfer_finished:
			if len(cc.buf) > 0 {
				defer cc.clearBuf()

				// TODO: FIXME (workaround)
				if len(cc.buf) == len(mysteryFrame) {
					isTheSame := true
					for i := range mysteryFrame {
						if mysteryFrame[i] != cc.buf[i] {
							isTheSame = false
						}
					}

					if isTheSame {
						return nil
					}
				}
			}
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
		case init_phase:
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
		case pairing_phase:
			switch cc.nextExpectedMessage {
			// paired key encryption
			case paired_key_encryption:
				err = cc.processPairedKeyEncryption(sharingFrame.GetV1().GetPairedKeyEncryption())
				nextMessage = paired_key_result
			// paired key result
			case paired_key_result:
				err = cc.processPairedKeyResult(sharingFrame.GetV1().GetPairedKeyResult())
				nextMessage = introduction
			}
		case transfer_phase:
			switch cc.nextExpectedMessage {
			// transfer introduction
			case introduction:
				err = cc.processIntroduction(sharingFrame.GetV1().GetIntroduction())
				nextMessage = accept_reject
			// accept or reject response
			case accept_reject:
				err = cc.processTransferRequest(sharingFrame.GetV1().GetConnectionResponse())
				nextMessage = transfer_complete
			// transfer complete
			case transfer_complete:
				err = cc.processTransferComplete()
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
	case paired_key_result, paired_key_encryption, introduction, accept_reject:
		if err = proto.Unmarshal(cc.buf, &sharingFrame); err != nil {
			return nil, nil, nil, fmt.Errorf("unmarshal sharing frame: %w", err)
		}
		return nil, nil, &sharingFrame, nil
	case transfer_complete:
		return nil, nil, nil, nil
	}

	return nil, nil, nil, ErrInternalError
}
