package crypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

func (c *Cipher) Encrypt(msg []byte) ([]byte, error) {
	msg = pad(msg)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("generate random iv: %w", err)
	}

	enc := cipher.NewCBCEncrypter(c.encryptBlock, iv)
	encMsg := make([]byte, len(msg)) // TODO: maybe reusue msg
	enc.CryptBlocks(encMsg, msg)

	return encMsg, nil
}

func pad(body []byte) []byte {
	padLen := aes.BlockSize - len(body)%aes.BlockSize
	if padLen > 0 {
		padding := bytes.Repeat([]byte{byte(padLen)}, padLen)
		body = append(body, padding...)
	}

	return body
}
