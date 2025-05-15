package comm

import (
	"fmt"

	"github.com/ACLzz/go-qshare/internal/crypt"
	pbSharing "github.com/ACLzz/go-qshare/protobuf/gen/sharing"
	"google.golang.org/protobuf/proto"
)

func (cc *commConn) processPairedKeyEncryption(msg *pbSharing.PairedKeyEncryptionFrame) error {
	if msg == nil {
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

	data, err := proto.Marshal(&pbSharing.Frame{
		Version: pbSharing.Frame_V1.Enum(),
		V1: &pbSharing.V1Frame{
			Type: pbSharing.V1Frame_PAIRED_KEY_ENCRYPTION.Enum(),
			PairedKeyEncryption: &pbSharing.PairedKeyEncryptionFrame{
				SecretIdHash: secretIDHash,
				SignedData:   signedData,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("marshal paired key encryption frame: %w", err)
	}

	if err := cc.writeMessage(data); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return nil
}

func (cc *commConn) processPairedKeyResult(msg *pbSharing.PairedKeyResultFrame) error {
	if msg == nil {
		return ErrInvalidMessage
	}

	data, err := proto.Marshal(&pbSharing.Frame{
		Version: pbSharing.Frame_V1.Enum(),
		V1: &pbSharing.V1Frame{
			Type: pbSharing.V1Frame_PAIRED_KEY_RESULT.Enum(),
			PairedKeyResult: &pbSharing.PairedKeyResultFrame{
				Status: pbSharing.PairedKeyResultFrame_UNABLE.Enum(),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("marshal paired key encryption frame: %w", err)
	}

	if err := cc.writeMessage(data); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	cc.phase = transferPhase
	return nil
}
