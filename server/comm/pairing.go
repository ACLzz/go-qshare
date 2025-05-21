package comm

import (
	"fmt"

	"github.com/ACLzz/go-qshare/internal/crypt"
	pbSharing "github.com/ACLzz/go-qshare/protobuf/gen/sharing"
)

func (cc *commConn) processPairedKeyEncryption(frame *pbSharing.PairedKeyEncryptionFrame) error {
	if len(frame.GetSignedData()) == 0 {
		return ErrInvalidMessage
	}

	secretIDHash, err := crypt.RandomBytes(6)
	if err != nil {
		return fmt.Errorf("generate secret id hash: %w", err)
	}

	signedData, err := crypt.RandomBytes(72)
	if err != nil {
		return fmt.Errorf("generate signed data: %w", err)
	}

	if err := cc.writeSecureFrame(&pbSharing.V1Frame{
		Type: pbSharing.V1Frame_PAIRED_KEY_ENCRYPTION.Enum(),
		PairedKeyEncryption: &pbSharing.PairedKeyEncryptionFrame{
			SecretIdHash: secretIDHash,
			SignedData:   signedData,
		},
	}); err != nil {
		return fmt.Errorf("write paired key encryption: %w", err)
	}

	return nil
}

func (cc *commConn) processPairedKeyResult(frame *pbSharing.PairedKeyResultFrame) error {
	if frame.GetStatus() != pbSharing.PairedKeyResultFrame_UNABLE {
		return ErrInvalidMessage
	}

	if err := cc.writeSecureFrame(&pbSharing.V1Frame{
		Type: pbSharing.V1Frame_PAIRED_KEY_RESULT.Enum(),
		PairedKeyResult: &pbSharing.PairedKeyResultFrame{
			Status: pbSharing.PairedKeyResultFrame_UNABLE.Enum(),
		},
	}); err != nil {
		return fmt.Errorf("write paired key result: %w", err)
	}

	cc.phase = transferPhase
	return nil
}
