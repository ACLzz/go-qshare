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
	pbConnections "github.com/ACLzz/go-qshare/protobuf/gen/connections"
	pbSecuregcm "github.com/ACLzz/go-qshare/protobuf/gen/securegcm"
	"google.golang.org/protobuf/proto"
)

const max_message_length = 32 * 1024 // 32 Kb

type commConn struct {
	conn                net.Conn
	nextExpectedMessage messageType
	phase               phase
	cipher              crypt.Cipher
}

func newCommConn(conn net.Conn) commConn {
	cConn := commConn{
		conn:                conn,
		nextExpectedMessage: conn_request,
		phase:               initConnPhase,
		cipher:              crypt.NewCipher(true),
	}

	return cConn
}

func (cc *commConn) Accept(ctx context.Context, wg *sync.WaitGroup) {
	var (
		msgLen uint32
		err    error
		lenBuf = make([]byte, 4)
	)
	defer wg.Done()

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

				fmt.Printf("ERR: read msg length: %s", err.Error())
				continue
			}

			msgLen = binary.BigEndian.Uint32(lenBuf[:])
			if msgLen > max_message_length {
				fmt.Printf("ERR: message is too long (%d)\n", msgLen)
				return
			}

			fmt.Printf("DBG: message length: %d\n", msgLen)
			msgBuf := make([]byte, msgLen)

			// fetch message in bytes
			if _, err = cc.conn.Read(msgBuf); err != nil {
				fmt.Printf("ERR: fetch message: %s", err.Error())
				continue
			}

			// route the message
			if err = cc.route(msgBuf); err != nil {
				if errors.Is(err, ErrInvalidMessage) {
					if err = cc.writeError(pbSecuregcm.Ukey2Alert_BAD_MESSAGE, ErrInvalidMessage.Error()); err != nil {
						fmt.Printf("ERR: send error to client: %s", err.Error())
					}
					continue
				}

				if errors.Is(err, ErrConnWasEndedByClient) {
					fmt.Println("DBG: conn was ended by client")
					return
				}

				fmt.Printf("ERR: process message: %s", err.Error())
				continue
			}
		}
	}
}

func (cc *commConn) processKeepAliveMessage(msg []byte) error {
	var frame pbConnections.OfflineFrame
	if err := proto.Unmarshal(msg, &frame); err != nil {
		return fmt.Errorf("unmarshal frame: %w", err)
	}

	if frame.GetV1().GetType() != pbConnections.V1Frame_KEEP_ALIVE {
		return ErrInvalidMessage
	}

	if err := cc.writeOfflineFrame(&pbConnections.V1Frame{
		Type: pbConnections.V1Frame_KEEP_ALIVE.Enum(),
		KeepAlive: &pbConnections.KeepAliveFrame{
			Ack: proto.Bool(true),
		},
	}); err != nil {
		return fmt.Errorf("write offline frame: %w", err)
	}

	return nil
}

func (cc *commConn) processUKEYAlert(msg []byte) error {
	var ukeyMessage pbSecuregcm.Ukey2Message
	if err := proto.Unmarshal(msg, &ukeyMessage); err != nil {
		return fmt.Errorf("unmarshal ukey message: %w", err)
	}

	var alert pbSecuregcm.Ukey2Alert
	if err := proto.Unmarshal(ukeyMessage.GetMessageData(), &alert); err != nil {
		return fmt.Errorf("unmarshal alert: %w", err)
	}

	fmt.Printf("ERR: got an alert (%s)\n", pbSecuregcm.Ukey2Alert_AlertType_name[int32(alert.GetType())])
	fmt.Printf("ERR: alert message (%s)\n", alert.GetErrorMessage())
	return nil
}

func (cc *commConn) writeError(t pbSecuregcm.Ukey2Alert_AlertType, msg string) error {
	alertMsg, err := proto.Marshal(&pbSecuregcm.Ukey2Alert{
		Type:         t.Enum(),
		ErrorMessage: proto.String(msg),
	})
	if err != nil {
		return fmt.Errorf("marshal alert message: %w", err)
	}

	if err := cc.writeUKEYMessage(pbSecuregcm.Ukey2Message_ALERT, alertMsg); err != nil {
		return fmt.Errorf("write UKEY message: %w", err)
	}

	return nil
}

func (cc *commConn) writeOfflineFrame(frame *pbConnections.V1Frame) error {
	offlineFrame, err := proto.Marshal(&pbConnections.OfflineFrame{
		Version: pbConnections.OfflineFrame_V1.Enum(),
		V1:      frame,
	})
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	if err = cc.writeMessage(offlineFrame); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return nil
}

func (cc *commConn) writeUKEYMessage(t pbSecuregcm.Ukey2Message_Type, msg []byte) error {
	ukeyMsg, err := proto.Marshal(&pbSecuregcm.Ukey2Message{
		MessageType: t.Enum(),
		MessageData: msg,
	})
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	if err = cc.writeMessage(ukeyMsg); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return nil
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

	return nil
}
