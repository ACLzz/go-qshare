package comm

import (
	pbSharing "github.com/ACLzz/go-qshare/protobuf/gen/sharing"
)

func (cc *commConn) processIntroduction(msg *pbSharing.IntroductionFrame) error {
	if len(msg.GetTextMetadata()) == 0 {
		return ErrInvalidMessage
	}
	return nil
}
