package comm

import (
	"fmt"

	"github.com/ACLzz/go-qshare/internal/crypt"
	pbSharing "github.com/ACLzz/go-qshare/protobuf/gen/sharing"
	"google.golang.org/protobuf/proto"
)

func (cc *commConn) processPairedKeyEncryption() error {
	var frame pbSharing.Frame
	defer func() {
		cc.buf = make([]byte, init_buf_size)
	}()
	if err := proto.Unmarshal(cc.buf, &frame); err != nil {
		return fmt.Errorf("unmarshal offline frame: %w", err)
	}

	if len(frame.GetV1().GetPairedKeyEncryption().GetSignedData()) == 0 {
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
		return fmt.Errorf("write secure frame: %w", err)
	}

	return nil
}

func (cc *commConn) processPairedKeyResult() error {
	if err := cc.writeSecureFrame(&pbSharing.V1Frame{
		Type: pbSharing.V1Frame_PAIRED_KEY_RESULT.Enum(),
		PairedKeyResult: &pbSharing.PairedKeyResultFrame{
			Status: pbSharing.PairedKeyResultFrame_UNABLE.Enum(),
		},
	}); err != nil {
		return fmt.Errorf("write key result: %w", err)
	}

	cc.phase = transferPhase
	return nil
}
