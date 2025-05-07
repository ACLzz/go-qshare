package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"math/big"

	pbConnections "github.com/ACLzz/go-qshare/protobuf/gen/connections"
	pbSecuregcm "github.com/ACLzz/go-qshare/protobuf/gen/securegcm"
	pbSecureMessage "github.com/ACLzz/go-qshare/protobuf/gen/securemessage"
	"google.golang.org/protobuf/proto"
)

const targetCipher = pbSecuregcm.Ukey2HandshakeCipher_P256_SHA512

var defaultUKEYMessageVersion = int32(1)

func (cc *commConn) establishSecureConnection(msg []byte) error {
	var frame pbConnections.V1Frame
	if cc.connRequest == nil || cc.ukeyClientFinished != nil {
		if err := proto.Unmarshal(msg, &frame); err != nil {
			return fmt.Errorf("unmarshal frame: %w", err)
		}
	}

	var ukeyMessage pbSecuregcm.Ukey2Message
	if cc.ukeyClientInit == nil || cc.ukeyClientFinished == nil {
		if err := proto.Unmarshal(msg, &ukeyMessage); err != nil {
			return fmt.Errorf("unmarshal UKEY message: %w", err)
		}

		if isAlertUKEY2MessageType(ukeyMessage.GetMessageType()) {
			return fmt.Errorf("got an error from client: '%s'", string(ukeyMessage.GetMessageData()))
		}
	}

	// first seek for the connection request
	if cc.connRequest == nil {
		if frame.GetType() != pbConnections.V1Frame_CONNECTION_REQUEST {
			return ErrInvalidMessage
		}

		var connRequest pbConnections.ConnectionRequestFrame
		if err := proto.Unmarshal([]byte(frame.GetConnectionRequest().GetEndpointName()), &connRequest); err != nil {
			return fmt.Errorf("unmarshal Connection request: %w", err)
		}

		return cc.processConnRequest(&connRequest)

	}

	// second seek for the ukey client init message
	if cc.ukeyClientInit == nil {
		if ukeyMessage.GetMessageType() != pbSecuregcm.Ukey2Message_CLIENT_INIT {
			return ErrInvalidMessage
		}

		var clientInitMsg pbSecuregcm.Ukey2ClientInit
		if err := proto.Unmarshal(ukeyMessage.GetMessageData(), &clientInitMsg); err != nil {
			return fmt.Errorf("unmarshal UKEY Client Init message: %w", err)
		}

		return cc.processUKEYInitMessage(&clientInitMsg)
	}

	// third seek for the ukey client finished message
	if cc.ukeyClientFinished == nil {
		if ukeyMessage.GetMessageType() != pbSecuregcm.Ukey2Message_CLIENT_FINISH {
			return ErrInvalidMessage
		}

		var clientFinishedMsg pbSecuregcm.Ukey2ClientFinished
		if err := proto.Unmarshal(ukeyMessage.GetMessageData(), &clientFinishedMsg); err != nil {
			return fmt.Errorf("unmarshal UKEY Client Finish message: %w", err)
		}

		return cc.processUKEYFinishedMessage(&clientFinishedMsg)
	}

	// finally seek for connection response
	if cc.connResponse == nil {
		if frame.GetType() != pbConnections.V1Frame_CONNECTION_REQUEST {
			return ErrInvalidMessage
		}

		// var connResponse pbConnections.ConnectionResponseFrame
		// if err := proto.Unmarshal([]byte(frame.ConnectionResponse.GetResponse()), &connResponse); err != nil {
		// 	return fmt.Errorf("unmarshal Connection response: %w", err)
		// }

		return cc.processConnResponse(frame.GetConnectionRequest())
	}

	return ErrInvalidMessage
}

func (cc *commConn) processConnRequest(req *pbConnections.ConnectionRequestFrame) error {
	if req == nil {
		return ErrInvalidMessage
	}

	if len(req.EndpointInfo) == 0 {
		return ErrInvalidMessage
	}

	cc.connRequest = req
	return nil
}

func (cc *commConn) processUKEYInitMessage(msg *pbSecuregcm.Ukey2ClientInit) error {
	cc.ukeyClientInit = msg

	// TODO: maybe also add another type of cipher
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

	// TODO: do we need to store private key?
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generate ecdsa private key: %w", err)
	}

	ukeyMsg, err := newServerInitMessage(privateKey)
	if err != nil {
		return fmt.Errorf("marshal UKEY2 Message: %w", err)
	}

	if err = cc.writeMessage(ukeyMsg); err != nil {
		return fmt.Errorf("send UKEY2 Message: %w", err)
	}

	return nil
}

func (cc *commConn) processUKEYFinishedMessage(req *pbSecuregcm.Ukey2ClientFinished) error {
	if req == nil {
		return ErrInvalidMessage
	}

	cc.ukeyClientFinished = req

	serverConnResponse, err := proto.Marshal(&pbConnections.V1Frame{
		Type:               pbConnections.V1Frame_CONNECTION_RESPONSE.Enum(),
		ConnectionResponse: &pbConnections.ConnectionResponseFrame{},
	})
	if err != nil {
		return fmt.Errorf("marshal Connection response: %w", err)
	}

	if err = cc.writeMessage(serverConnResponse); err != nil {
		return fmt.Errorf("write Connection response: %w", err)
	}
	return nil
}

func (cc *commConn) processConnResponse(req *pbConnections.ConnectionRequestFrame) error {
	if req == nil {
		return ErrInvalidMessage
	}

	cc.connResponse = req
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

	random, err := randomBytes(32)
	if err != nil {
		return nil, fmt.Errorf("generate random field for UKEY2 Server Init message: %w", err)
	}

	serverUKEYInit, err := proto.Marshal(&pbSecuregcm.Ukey2ServerInit{
		Version:         &defaultUKEYMessageVersion,
		Random:          random,
		HandshakeCipher: targetCipher.Enum(),
		PublicKey:       pbPublicKey,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal UKEY2 Server Init response: %w", err)
	}

	ukeyMsg, err := proto.Marshal(&pbSecuregcm.Ukey2Message{
		MessageType: pbSecuregcm.Ukey2Message_SERVER_INIT.Enum(),
		MessageData: serverUKEYInit,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal UKEY2 Message: %w", err)
	}

	return ukeyMsg, nil
}

func isAlertUKEY2MessageType(t pbSecuregcm.Ukey2Message_Type) bool {
	// TODO: fixme
	// for k := range pbSecuregcm.Ukey2Alert_AlertType_name {
	// 	if int32(t.Number()) == k {
	// 		return true
	// 	}
	// }

	return t >= 100
}

func randomBytes(size uint8) ([]byte, error) {
	buf := make([]byte, size)

	for i := range size {
		n, err := rand.Int(rand.Reader, big.NewInt(256))
		if err != nil {
			return nil, fmt.Errorf("generate random int: %w", err)
		}

		buf[i] = byte(n.Int64())
	}

	return buf, nil
}
