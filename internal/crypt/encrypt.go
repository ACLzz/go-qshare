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
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("generate random iv: %w", err)
	}

	enc := cipher.NewCBCEncrypter(c.encryptBlock, iv)

	padLen := aes.BlockSize - len(msg)%aes.BlockSize
	padding := bytes.Repeat([]byte{byte(padLen)}, padLen)
	msg = append(msg, padding...)

	encMsg := make([]byte, len(msg))
	enc.CryptBlocks(encMsg, msg)

	return encMsg, nil
}
