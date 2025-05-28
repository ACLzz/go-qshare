package comm

import (
	"fmt"

	pbSecuregcm "github.com/ACLzz/go-qshare/internal/protobuf/gen/securegcm"
	pbSecureMessage "github.com/ACLzz/go-qshare/internal/protobuf/gen/securemessage"
	"google.golang.org/protobuf/proto"
)

func (a *Adapter) ValidateClientFinished(msg []byte) error {
	var (
		ukeyMessage        pbSecuregcm.Ukey2Message
		ukeyClientFinished pbSecuregcm.Ukey2ClientFinished
		genericPublicKey   pbSecureMessage.GenericPublicKey
	)

	// unmarshal client finished message
	if err := proto.Unmarshal(msg, &ukeyMessage); err != nil {
		return fmt.Errorf("unmarshal ukey message: %w", err)
	}
	if ukeyMessage.GetMessageType() != pbSecuregcm.Ukey2Message_CLIENT_FINISH {
		return ErrInvalidMessage
	}
	if err := proto.Unmarshal(ukeyMessage.GetMessageData(), &ukeyClientFinished); err != nil {
		return fmt.Errorf("unmarshal ukey client finished: %w", err)
	}

	// unmarshal public key
	if err := proto.Unmarshal(ukeyClientFinished.GetPublicKey(), &genericPublicKey); err != nil {
		return fmt.Errorf("decode public key: %w", err)
	}

	// add public key to the cipher
	if err := a.cipher.SetSenderPublicKey(genericPublicKey.GetEcP256PublicKey()); err != nil {
		return fmt.Errorf("add client public key: %w", err)
	}

	return nil
}
