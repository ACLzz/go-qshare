package adapter

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"

	"github.com/ACLzz/go-qshare/internal/crypt"
	pbSecuregcm "github.com/ACLzz/go-qshare/internal/protobuf/gen/securegcm"
	pbSecureMessage "github.com/ACLzz/go-qshare/internal/protobuf/gen/securemessage"
	"google.golang.org/protobuf/proto"
)

func (a *Adapter) SendServerInit() error {
	// generate private-public key pair
	privateKey, publicKey, err := generatePrivatePublicKeys()
	if err != nil {
		return fmt.Errorf("generate private public key pair: %w", err)
	}

	// prepare server init message
	randomData, err := crypt.RandomBytes(32)
	if err != nil {
		return fmt.Errorf("generate random data: %w", err)
	}

	version := int32(1)
	serverInitMsg, err := proto.Marshal(&pbSecuregcm.Ukey2ServerInit{
		Version:         &version,
		Random:          randomData,
		HandshakeCipher: targetCipher.Enum(),
		PublicKey:       publicKey,
	})
	if err != nil {
		return fmt.Errorf("marshal server init message: %w", err)
	}

	// add server init message and private key to the cipher
	if err = a.cipher.SetReceiverPrivateKey(privateKey); err != nil {
		return fmt.Errorf("add receiver private key: %w", err)
	}
	serverInit, err := proto.Marshal(&pbSecuregcm.Ukey2Message{
		MessageType: pbSecuregcm.Ukey2Message_SERVER_INIT.Enum(),
		MessageData: serverInitMsg,
	})
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	if err = a.cipher.SetReceiverInitMessage(serverInit); err != nil {
		return fmt.Errorf("add receiver init message: %w", err)
	}

	// send message
	err = a.writeUKEYMessage(pbSecuregcm.Ukey2Message_SERVER_INIT, serverInitMsg)
	if err != nil {
		return fmt.Errorf("write ukey message: %w", err)
	}

	return nil
}

func (a *Adapter) ValidateServerInit(msg []byte) error {
	// unmarshal server init message
	var ukeyMessage pbSecuregcm.Ukey2Message
	if err := proto.Unmarshal(msg, &ukeyMessage); err != nil {
		return fmt.Errorf("unmarshal ukey message: %w", err)
	}

	if ukeyMessage.GetMessageType() != pbSecuregcm.Ukey2Message_SERVER_INIT {
		return ErrInvalidMessage
	}

	var serverInit pbSecuregcm.Ukey2ServerInit
	if err := proto.Unmarshal(ukeyMessage.GetMessageData(), &serverInit); err != nil {
		return fmt.Errorf("unmarshal server init: %w", err)
	}

	// validate
	if serverInit.GetHandshakeCipher() != targetCipher {
		return ErrInvalidMessage
	}

	// unmarshal public key
	var publicKey pbSecureMessage.GenericPublicKey
	if err := proto.Unmarshal(serverInit.GetPublicKey(), &publicKey); err != nil {
		return fmt.Errorf("unmarshal public key: %w", err)
	}
	if publicKey.GetEcP256PublicKey() == nil {
		return ErrInvalidMessage
	}

	// add acquired data to cipher
	if err := a.cipher.SetSenderPublicKey(publicKey.GetEcP256PublicKey()); err != nil {
		return fmt.Errorf("add sender public key: %w", err)
	}
	if err := a.cipher.SetSenderInitMessage(msg); err != nil {
		return fmt.Errorf("add sender public key: %w", err)
	}

	return nil
}

func generatePrivatePublicKeys() (*ecdsa.PrivateKey, []byte, error) {
	// generate private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate ecdsa private key: %w", err)
	}

	// prepare server init message
	pbPublicKey, err := proto.Marshal(&pbSecureMessage.GenericPublicKey{
		Type: pbSecureMessage.PublicKeyType_EC_P256.Enum(),
		EcP256PublicKey: &pbSecureMessage.EcP256PublicKey{
			X: privateKey.PublicKey.X.FillBytes(make([]byte, 33)),
			Y: privateKey.PublicKey.Y.FillBytes(make([]byte, 33)),
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("marshal Generic Public Key: %w", err)
	}

	return privateKey, pbPublicKey, nil
}
