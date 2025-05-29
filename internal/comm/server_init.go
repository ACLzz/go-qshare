package comm

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
	if err = a.cipher.SetReceiverInitMessage(serverInitMsg); err != nil {
		return fmt.Errorf("add receiver init message: %w", err)
	}

	// send message
	err = a.WriteUKEYMessage(pbSecuregcm.Ukey2Message_SERVER_INIT, serverInitMsg)
	if err != nil {
		return fmt.Errorf("write ukey message: %w", err)
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
