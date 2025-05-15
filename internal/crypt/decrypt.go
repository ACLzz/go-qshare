package crypt

import (
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"

	pbSecuregcm "github.com/ACLzz/go-qshare/protobuf/gen/securegcm"
	pbSecureMessage "github.com/ACLzz/go-qshare/protobuf/gen/securemessage"
	"google.golang.org/protobuf/proto"
)

func (c *Cipher) Decrypt(msg []byte) ([]byte, error) {
	var secMsg pbSecureMessage.SecureMessage
	if err := proto.Unmarshal(msg, &secMsg); err != nil {
		return nil, fmt.Errorf("unmarshal secure message: %w", err)
	}

	if err := validateSecureMessage(&secMsg, c.senderHMACKey); err != nil {
		return nil, err
	}

	var hb pbSecureMessage.HeaderAndBody
	if err := proto.Unmarshal(secMsg.GetHeaderAndBody(), &hb); err != nil {
		return nil, fmt.Errorf("unmarshal secure message: %w", err)
	}

	if err := validateHeaderAndBody(&hb); err != nil {
		return nil, err
	}

	body := make([]byte, len(hb.GetBody()))
	dec := cipher.NewCBCDecrypter(c.decryptBlock, hb.GetHeader().GetIv())
	dec.CryptBlocks(body, hb.GetBody())
	body = unpad(body)

	var d2dMsg pbSecuregcm.DeviceToDeviceMessage
	if err := proto.Unmarshal(body, &d2dMsg); err != nil {
		return nil, fmt.Errorf("unmarshal device to device message: %w", err)
	}

	return d2dMsg.GetMessage(), nil
}

// unpad: un-padding pcks7 padding
func unpad(body []byte) []byte {
	addedChars := int(body[len(body)-1])
	return body[:len(body)-addedChars]
}

func validateSecureMessage(msg *pbSecureMessage.SecureMessage, senderKey []byte) error {
	h := hmac.New(sha256.New, senderKey)
	if _, err := h.Write(msg.GetHeaderAndBody()); err != nil {
		return fmt.Errorf("get hmac_sha256 from header and body: %w", err)
	}

	expectedSig := h.Sum(nil)
	givenSig := msg.GetSignature()

	for i := range expectedSig {
		if expectedSig[i] != givenSig[i] {
			return ErrInvalidSecureMessageSignature
		}
	}

	return nil
}

func validateHeaderAndBody(hb *pbSecureMessage.HeaderAndBody) error {
	if hb.GetHeader().GetEncryptionScheme() != pbSecureMessage.EncScheme_AES_256_CBC {
		return ErrInvalidEncryptionScheme
	}
	if hb.GetHeader().GetSignatureScheme() != pbSecureMessage.SigScheme_HMAC_SHA256 {
		return ErrInvalidSignatureScheme
	}
	if len(hb.GetHeader().GetIv()) < 16 {
		return ErrInvalidIV
	}

	return nil
}
