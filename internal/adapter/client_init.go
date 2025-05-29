package adapter

import (
	"crypto/sha512"
	"fmt"

	"github.com/ACLzz/go-qshare/internal/crypt"
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

func (a *Adapter) SendClientInitWithClientFinished() error {
	// prepare client finish
	clientFinished, err := a.marshalClientFinished()
	if err != nil {
		return fmt.Errorf("mrshal client finished: %w", err)
	}
	commitment := sha512.Sum512(clientFinished)

	// prepare client init
	random, err := crypt.RandomBytes(32)
	if err != nil {
		return fmt.Errorf("generate random bytes: %w", err)
	}
	clientInit, err := proto.Marshal(&pbSecuregcm.Ukey2ClientInit{
		Version: proto.Int32(1),
		Random:  random,
		CipherCommitments: []*pbSecuregcm.Ukey2ClientInit_CipherCommitment{
			{
				HandshakeCipher: pbSecuregcm.Ukey2HandshakeCipher_P256_SHA512.Enum(),
				Commitment:      commitment[:],
			},
		},
		NextProtocol: proto.String("AES_256_CBC-HMAC_SHA256"),
	})
	if err != nil {
		return fmt.Errorf("marshal client init: %w", err)
	}

	// send messages
	if err := a.writeUKEYMessage(pbSecuregcm.Ukey2Message_CLIENT_INIT, clientInit); err != nil {
		return fmt.Errorf("write client init message: %w", err)
	}

	if err := a.writeUKEYMessage(pbSecuregcm.Ukey2Message_CLIENT_FINISH, clientFinished); err != nil {
		return fmt.Errorf("write client finished message: %w", err)
	}
	// TODO: add required data to cipher

	return nil
}
