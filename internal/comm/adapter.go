package comm

import (
	"net"

	"github.com/ACLzz/go-qshare/internal/crypt"
	pbConnections "github.com/ACLzz/go-qshare/internal/protobuf/gen/connections"
	"github.com/ACLzz/go-qshare/log"
	"google.golang.org/protobuf/proto"
)

type Adapter struct {
	conn   net.Conn
	cipher *crypt.Cipher
	log    log.Logger

	seqNumber            int32
	isEncrypted          bool
	isTransfer           bool
	fileChunkCallback    FileChunkCallback
	textTransferCallback TextTransferCallback
}

// TODO: remove cipher from arguments and manage it internally
func NewAdapter(
	conn net.Conn,
	logger log.Logger,
	cipher *crypt.Cipher,
	fileChunkCallback FileChunkCallback,
	textTransferCallback TextTransferCallback,
) Adapter {
	return Adapter{
		conn:                 conn,
		cipher:               cipher,
		log:                  logger,
		fileChunkCallback:    fileChunkCallback,
		textTransferCallback: textTransferCallback,
	}
}

func (a *Adapter) EnableEncryption() error {
	if err := a.cipher.Setup(); err != nil {
		return err
	}

	a.isEncrypted = true
	return nil
}

func (a *Adapter) EnableTransferHandler() {
	a.isTransfer = true
}

func (a *Adapter) Disconnect() {
	if err := a.WriteOfflineFrame(&pbConnections.V1Frame{
		Type: pbConnections.V1Frame_DISCONNECTION.Enum(),
		Disconnection: &pbConnections.DisconnectionFrame{
			AckSafeToDisconnect: proto.Bool(true),
		},
	}); err != nil {
		a.log.Error("error while disconnecting", err)
	}
}
