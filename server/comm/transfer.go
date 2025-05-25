package comm

import (
	"fmt"

	pbSharing "github.com/ACLzz/go-qshare/protobuf/gen/sharing"
)

func (cc *commConn) processIntroduction(msg *pbSharing.IntroductionFrame) error {
	if len(msg.GetTextMetadata()) == 0 {
		return ErrInvalidMessage
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
	cc.log.Debug("transfer completed", "size", len(cc.buf), "data", string(cc.buf))
	cc.phase = disconnectPhase

	return nil
}
