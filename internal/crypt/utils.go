package crypt

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

var alphaNumRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandomBytes(size uint8) ([]byte, error) {
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

func RandomAlphaNum(size uint8) ([]byte, error) {
	buf := make([]byte, size)
	for i := range size {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphaNumRunes))))
		if err != nil {
			return nil, fmt.Errorf("generate index for random alphanum rune: %w", err)
		}
		buf[i] = byte(alphaNumRunes[idx.Int64()])
	}

	return buf, nil
}
