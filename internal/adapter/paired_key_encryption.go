package adapter

import (
	"fmt"

	"github.com/ACLzz/go-qshare/internal/crypt"
	pbSharing "github.com/ACLzz/go-qshare/internal/protobuf/gen/sharing"
)

func (a *Adapter) ValidatePairedKeyEncryption(msg []byte) error {
	frame, err := unmarshalSharingFrame(msg)
	if err != nil {
		return err
	}

	if frame.GetType() != pbSharing.V1Frame_PAIRED_KEY_ENCRYPTION ||
		frame.GetPairedKeyEncryption() == nil ||
		len(frame.PairedKeyEncryption.GetSignedData()) == 0 {
		return ErrInvalidMessage
	}

	return nil
}

func (a *Adapter) SendPairedKeyEncryption() error {
	secretIDHash, err := crypt.RandomBytes(6)
	if err != nil {
		return fmt.Errorf("generate secret id hash: %w", err)
	}

	signedData, err := crypt.RandomBytes(72)
	if err != nil {
		return fmt.Errorf("generate signed data: %w", err)
	}

	return a.writeSecureFrame(&pbSharing.V1Frame{
		Type: pbSharing.V1Frame_PAIRED_KEY_ENCRYPTION.Enum(),
		PairedKeyEncryption: &pbSharing.PairedKeyEncryptionFrame{
			SecretIdHash: secretIDHash,
			SignedData:   signedData,
		},
	})
}
