package comm

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/ACLzz/go-qshare/internal/crypt"
	"github.com/ACLzz/go-qshare/log"
	pbConnections "github.com/ACLzz/go-qshare/protobuf/gen/connections"
	pbSecuregcm "github.com/ACLzz/go-qshare/protobuf/gen/securegcm"
	"google.golang.org/protobuf/proto"
)

const (
	max_message_length = 32 * 1024 // 32 Kb
	init_buf_size      = 1 * 1024  // 1 Kb
)

type commConn struct {
	conn                net.Conn
	cipher              crypt.Cipher
	log                 log.Logger
	nextExpectedMessage expectedMessage
	phase               phase
	seqNumber           int32
	buf                 []byte
}

func newCommConn(conn net.Conn, logger log.Logger) commConn {
	return commConn{
		conn:                conn,
		cipher:              crypt.NewCipher(true),
		log:                 logger,
		nextExpectedMessage: conn_request,
		phase:               initConnPhase,
		buf:                 make([]byte, 0, init_buf_size),
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
			if _, err = cc.conn.Read(lenBuf); err != nil {
				if errors.Is(err, io.EOF) {
					return
				}

				cc.log.Error("read msg length", err)
				continue
			}

			msgLen = binary.BigEndian.Uint32(lenBuf[:])
			if msgLen > max_message_length {
				cc.log.Error("message is too long", nil, "length", msgLen)
				return
			}

			cc.log.Debug("got message", "length", msgLen)
			msgBuf := make([]byte, msgLen)

			// fetch message in bytes
			if _, err = cc.conn.Read(msgBuf); err != nil {
				cc.log.Error("fetch message", err)
				continue
			}

			// route the message
			if err = cc.route(msgBuf); err != nil {
				if errors.Is(err, ErrInvalidMessage) {
					if err = cc.writeError(pbSecuregcm.Ukey2Alert_BAD_MESSAGE, ErrInvalidMessage.Error()); err != nil {
						cc.log.Error("send error to client", err)
					}
					cc.log.Warn("got invalid message", "expectedMessage", cc.nextExpectedMessage)
					continue
				}

				if errors.Is(err, ErrConnWasEndedByClient) {
					cc.log.Debug("conn was ended by client")
					return
				}

				cc.log.Error("process message", err)
				continue
			}
		}
	}
}

func (cc *commConn) processKeepAliveMessage(frame *pbConnections.OfflineFrame) {
	keepAliveFrame := pbConnections.V1Frame{
		Type: pbConnections.V1Frame_KEEP_ALIVE.Enum(),
		KeepAlive: &pbConnections.KeepAliveFrame{
			Ack: proto.Bool(true),
		},
	}

	var err error
	if cc.phase > initConnPhase {
		err = cc.encryptAndWrite(&keepAliveFrame)
	} else {
		err = cc.writeOfflineFrame(&keepAliveFrame)
	}
	if err != nil {
		cc.log.Error("send keep alive message", err)
	}
}

func (cc *commConn) processUKEYAlert(ukeyMessage *pbSecuregcm.Ukey2Message) {
	var alert pbSecuregcm.Ukey2Alert
	if err := proto.Unmarshal(ukeyMessage.GetMessageData(), &alert); err != nil {
		return
	}
	
	alertType, ok := pbSecuregcm.Ukey2Alert_AlertType_name[int32(alert.GetType())]
	if !ok {
		return
	}

	cc.log.Warn("got an alert", "type", alertType, "message", alert.GetErrorMessage())
}

func (cc *commConn) writeMessage(data []byte) error {
	// TODO: is there more efficient way to do this?
	msgLength := len(data)
	msgWithPrefix := make([]byte, 4, msgLength+4)
	msgWithPrefix[0] = byte(uint8(msgLength >> 24))
	msgWithPrefix[1] = byte(uint8(msgLength >> 16))
	msgWithPrefix[2] = byte(uint8(msgLength >> 8))
	msgWithPrefix[3] = byte(uint8(msgLength))
	msgWithPrefix = append(msgWithPrefix, data...)

	if _, err := cc.conn.Write(msgWithPrefix); err != nil {
		return fmt.Errorf("send message to client: %w", err)
	}

	cc.log.Debug("sent message to client")
	return nil
}

func (cc *commConn) clearBuf() {
	cc.buf = make([]byte, 0, init_buf_size)
}
