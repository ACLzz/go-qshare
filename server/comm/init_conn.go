package comm

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"

	"github.com/ACLzz/go-qshare/internal/crypt"
	pbConnections "github.com/ACLzz/go-qshare/protobuf/gen/connections"
	pbSecuregcm "github.com/ACLzz/go-qshare/protobuf/gen/securegcm"
	pbSecureMessage "github.com/ACLzz/go-qshare/protobuf/gen/securemessage"
	"google.golang.org/protobuf/proto"
)

const targetCipher = pbSecuregcm.Ukey2HandshakeCipher_P256_SHA512

var defaultUKEYMessageVersion = int32(1)

func (cc *commConn) processConnRequest(req *pbConnections.ConnectionRequestFrame) error {
	// just validate and let it go
	if len(req.EndpointInfo) == 0 {
		return ErrInvalidMessage
	}

	return nil
}

func (cc *commConn) processUKEYInitMessage(ukeyMessage *pbSecuregcm.Ukey2Message, rawMsg []byte) error {
	if ukeyMessage == nil {
		return ErrInvalidMessage
	}
	var err error

	var msg pbSecuregcm.Ukey2ClientInit
	if err := proto.Unmarshal(ukeyMessage.GetMessageData(), &msg); err != nil {
		return fmt.Errorf("unmarshal UKEY Client Init message: %w", err)
	}

	// validate
	isSHA512Found := false
	for _, comm := range msg.CipherCommitments {
		if comm.GetHandshakeCipher() == targetCipher {
			isSHA512Found = true
			break
		}
	}
	if !isSHA512Found {
		return ErrInvalidHandshakeCipher
	}

	// generate private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generate ecdsa private key: %w", err)
	}

	// response with server init
	serverInitMsg, err := newServerInitMessage(privateKey)
	if err != nil {
		return fmt.Errorf("marshal UKEY Message: %w", err)
	}
	err = cc.writeUKEYMessage(pbSecuregcm.Ukey2Message_SERVER_INIT, serverInitMsg)
	if err != nil {
		return fmt.Errorf("write ukey message: %w", err)
	}

	// add acquired data to cipher
	if err = cc.cipher.SetReceiverPrivateKey(privateKey); err != nil {
		return fmt.Errorf("add receiver private key: %w", err)
	}
	if err = cc.cipher.SetReceiverInitMessage(serverInitMsg); err != nil {
		return fmt.Errorf("add receiver init message: %w", err)
	}
	if err = cc.cipher.SetSenderInitMessage(rawMsg); err != nil {
		return fmt.Errorf("add sender init message: %w", err)
	}
	return nil
}

func (cc *commConn) processUKEYFinishedMessage(ukeyMessage *pbSecuregcm.Ukey2Message) error {
	if ukeyMessage == nil {
		return ErrInvalidMessage
	}
	var msg pbSecuregcm.Ukey2ClientFinished
	if err := proto.Unmarshal(ukeyMessage.GetMessageData(), &msg); err != nil {
		return fmt.Errorf("unmarshal UKEY Client Finish message: %w", err)
	}

	// unmarshal public key
	var genericPublicKey pbSecureMessage.GenericPublicKey
	if err := proto.Unmarshal(msg.GetPublicKey(), &genericPublicKey); err != nil {
		return fmt.Errorf("decode public key: %w", err)
	}

	// add acquired data to cipher
	if err := cc.cipher.SetSenderPublicKey(genericPublicKey.GetEcP256PublicKey()); err != nil {
		return fmt.Errorf("add sender public key: %w", err)
	}

	return nil
}

func (cc *commConn) processConnResponse(req *pbConnections.ConnectionResponseFrame) error {
	// validate
	if req == nil {
		return ErrInvalidMessage
	}
	if req.GetResponse() != pbConnections.ConnectionResponseFrame_ACCEPT {
		return ErrConnWasEndedByClient
	}

	// response with accept
	if err := cc.writeOfflineFrame(&pbConnections.V1Frame{
		Type: pbConnections.V1Frame_CONNECTION_RESPONSE.Enum(),
		ConnectionResponse: &pbConnections.ConnectionResponseFrame{
			Response: pbConnections.ConnectionResponseFrame_ACCEPT.Enum(),
			Status:   proto.Int32(0),
			OsInfo: &pbConnections.OsInfo{
				Type: pbConnections.OsInfo_LINUX.Enum(), // TODO: maybe add other oses
			},
		},
	}); err != nil {
		return fmt.Errorf("write offline frame: %w", err)
	}

	// init connection phase done. assign pairing phase and setup cipher
	if err := cc.cipher.Setup(); err != nil {
		return fmt.Errorf("setup cipher: %w", err)
	}
	cc.phase = pairingPhase
	return nil
}

func newServerInitMessage(privateKey *ecdsa.PrivateKey) ([]byte, error) {
	pbPublicKey, err := proto.Marshal(&pbSecureMessage.GenericPublicKey{
		Type: pbSecureMessage.PublicKeyType_EC_P256.Enum(),
		EcP256PublicKey: &pbSecureMessage.EcP256PublicKey{
			X: privateKey.PublicKey.X.FillBytes(make([]byte, 33)), // TODO: what does it do?
			Y: privateKey.PublicKey.Y.FillBytes(make([]byte, 33)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal Generic Public Key: %w", err)
	}

	random, err := crypt.RandomBytes(32)
	if err != nil {
		return nil, fmt.Errorf("generate random field for UKEY2 Server Init message: %w", err)
	}

	ukeyServerInit, err := proto.Marshal(&pbSecuregcm.Ukey2ServerInit{
		Version:         &defaultUKEYMessageVersion,
		Random:          random,
		HandshakeCipher: targetCipher.Enum(),
		PublicKey:       pbPublicKey,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal UKEY2 Server Init response: %w", err)
	}

	return ukeyServerInit, nil
}
