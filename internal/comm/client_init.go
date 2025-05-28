package comm

import (
	"fmt"

	pbSecuregcm "github.com/ACLzz/go-qshare/internal/protobuf/gen/securegcm"
	"google.golang.org/protobuf/proto"
)

const targetCipher = pbSecuregcm.Ukey2HandshakeCipher_P256_SHA512

func (a *Adapter) ValidateClientInit(msg []byte) error {
	var (
		ukeyMessage    pbSecuregcm.Ukey2Message
		ukeyClientInit pbSecuregcm.Ukey2ClientInit
	)

	// unmarshal client init message
	if err := proto.Unmarshal(msg, &ukeyMessage); err != nil {
		return fmt.Errorf("unmarshal ukey message: %w", err)
	}
	if ukeyMessage.GetMessageType() != pbSecuregcm.Ukey2Message_CLIENT_INIT {
		return ErrInvalidMessage
	}
	if err := proto.Unmarshal(ukeyMessage.GetMessageData(), &ukeyClientInit); err != nil {
		return fmt.Errorf("unmarshal ukey client init: %w", err)
	}

	// look for correct cipher commitment
	isSHA512Found := false
	for _, comm := range ukeyClientInit.CipherCommitments {
		if comm.GetHandshakeCipher() == targetCipher {
			isSHA512Found = true
			break
		}
	}
	if !isSHA512Found {
		return ErrInvalidMessage
	}

	// add client init message to the cipher
	if err := a.cipher.SetSenderInitMessage(msg); err != nil {
		return fmt.Errorf("add client init message")
	}
	return nil
}
