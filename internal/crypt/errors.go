package crypt

import "errors"

var ErrInvalidCipher = errors.New("cipher was not prepared for setup")

var (
	ErrInvalidSenderInitMessage = errors.New("invalid sender init message")
	ErrInvalidSenderPublicKey   = errors.New("invalid sender public key")
)

var (
	ErrInvalidReceiverInitMessage = errors.New("invalid receiver init message")
	ErrInvalidReceiverPrivateKey  = errors.New("invalid receiver private key")
)

var (
	ErrInvalidSignatureScheme        = errors.New("got unsupported signature scheme")
	ErrInvalidIV                     = errors.New("got invalid iv in header")
	ErrInvalidEncryptionScheme       = errors.New("got unsupported encryption scheme")
	ErrInvalidSecureMessageSignature = errors.New("got invalid signature for secure message")
)
