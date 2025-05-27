package comm

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"

	qshare "github.com/ACLzz/go-qshare"
	adapter "github.com/ACLzz/go-qshare/internal/comm"
	"github.com/ACLzz/go-qshare/internal/crypt"
	"github.com/ACLzz/go-qshare/internal/payloads"
	pbSecuregcm "github.com/ACLzz/go-qshare/internal/protobuf/gen/securegcm"
	"github.com/ACLzz/go-qshare/log"
)

const (
	max_message_length = 5 * 1024 * 1024 // 5 MB
	init_buf_size      = 1 * 1024        // 1 Kb
)

type commConn struct {
	adapter adapter.Adapter
	conn    net.Conn
	cipher  *crypt.Cipher
	log     log.Logger

	nextExpectedMessage expectedMessage
	phase               phase
	buf                 []byte

	textPayload      *qshare.TextPayload // text payload can be sent only one
	filePayloads     map[int64]*payloads.FilePayload
	expectedPayloads int
	receivedPayloads int
	textCallback     qshare.TextCallback
	fileCallback     qshare.FileCallback
}

func newCommConn(
	conn net.Conn,
	logger log.Logger,
	textCallback qshare.TextCallback,
	fileCallback qshare.FileCallback,
) commConn {
	cipher := crypt.NewCipher(true)
	return commConn{
		adapter:             adapter.NewAdapter(conn, logger, &cipher),
		conn:                conn,
		cipher:              &cipher,
		log:                 logger,
		nextExpectedMessage: conn_request,
		phase:               init_phase,
		buf:                 make([]byte, 0, init_buf_size),
		textCallback:        textCallback,
		fileCallback:        fileCallback,
	}
}

func (cc *commConn) Accept(ctx context.Context, wg *sync.WaitGroup) {
	var (
		msgLen uint32
		err    error
		lenBuf = make([]byte, 4)
	)
	defer wg.Done()
	cc.log.Debug("got new connection")
	defer cc.log.Debug("connection closed")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// fetch length of the upcoming message
			if _, err = io.ReadFull(cc.conn, lenBuf); err != nil {
				if errors.Is(err, io.EOF) {
					cc.log.Debug("conn ended abruptly")
					return
				}

				cc.log.Error("read msg length", err)
				lenBuf = make([]byte, 4)
				continue
			}

			msgLen = binary.BigEndian.Uint32(lenBuf[:])
			lenBuf = make([]byte, 4)

			if msgLen > max_message_length {
				cc.log.Error("message is too long", nil, "length", msgLen)
				return
			}

			msgBuf := make([]byte, msgLen)
			// fetch message in bytes
			if _, err = io.ReadFull(cc.conn, msgBuf); err != nil {
				cc.log.Error("fetch message", err)
				continue
			}

			// route the message
			if err = cc.route(msgBuf); err != nil {
				if errors.Is(err, ErrInvalidMessage) {
					if err = cc.adapter.WriteError(pbSecuregcm.Ukey2Alert_BAD_MESSAGE, ErrInvalidMessage.Error()); err != nil {
						cc.log.Error("send error to client", err)
					}
					cc.log.Warn("got invalid message", "expectedMessage", cc.nextExpectedMessage)
					continue
				}

				if errors.Is(err, adapter.ErrConnWasEndedByClient) {
					cc.log.Debug("conn was ended by client")
					return
				}

				cc.log.Error("process message", err)
				continue
			}

			if cc.phase == disconnect_phase {
				cc.adapter.Disconnect()
				return
			}
		}
	}
}

func (cc *commConn) clearBuf() {
	cc.buf = make([]byte, 0, init_buf_size)
}
