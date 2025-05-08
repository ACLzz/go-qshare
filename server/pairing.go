package server

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	pbConnections "github.com/ACLzz/go-qshare/protobuf/gen/connections"
	pbSecuregcm "github.com/ACLzz/go-qshare/protobuf/gen/securegcm"
	pbSecureMessage "github.com/ACLzz/go-qshare/protobuf/gen/securemessage"
	"google.golang.org/protobuf/proto"
)

func (cc *commConn) pairKey(msg []byte) error {
	d2dMessage, err := cc.decryptMessage(msg)
	if err != nil {
		return fmt.Errorf("decrypt message: %w", err)
	}

	fmt.Println(d2dMessage.String())

	_ = d2dMessage

	return ErrInvalidMessage
}

func (cc *commConn) initParing() error {
	cc.phase = pairingPhase
	msg, err := proto.Marshal(&pbConnections.V1Frame{
		Type: pbConnections.V1Frame_PAIRED_KEY_ENCRYPTION.Enum(),
		PairedKeyEncryption: &pbConnections.PairedKeyEncryptionFrame{
			SignedData: []byte{},
		},
	})
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	msg, err = cc.encryptMessage(msg)
	if err != nil {
		return fmt.Errorf("encrypt message: %w", err)
	}

	if err = cc.writeMessage(msg); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return nil
}

// https://gist.github.com/JonasDoe/48cda246b080d2fe3b55736c8bd1b998
func (cc *commConn) decryptMessage(msg []byte) (*pbSecuregcm.DeviceToDeviceMessage, error) {
	var hb pbSecureMessage.HeaderAndBody
	if err := proto.Unmarshal(msg, &hb); err != nil {
		return nil, fmt.Errorf("unmarshal header and body: %w", err)
	}

	bodyBuf := make([]byte, len(hb.GetBody()))
	bm := cipher.NewCBCDecrypter(cc.d2dDecryptBlock, hb.GetHeader().GetIv())
	bm.CryptBlocks(bodyBuf, hb.GetBody())

	var d2dMessage pbSecuregcm.DeviceToDeviceMessage
	if err := proto.Unmarshal(bodyBuf, &d2dMessage); err != nil {
		return nil, fmt.Errorf("unmarshal body: %w", err)
	}

	return &d2dMessage, nil
}

func (cc *commConn) encryptMessage(msg []byte) ([]byte, error) {
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("generate random iv: %w", err)
	}

	cbc := cipher.NewCBCEncrypter(cc.d2dEncryptBlock, iv)

	padLen := aes.BlockSize - len(msg)%aes.BlockSize
	padding := bytes.Repeat([]byte{byte(padLen)}, padLen)
	msg = append(msg, padding...)

	encMsg := make([]byte, len(msg))
	cbc.CryptBlocks(encMsg, msg)

	return encMsg, nil
}
