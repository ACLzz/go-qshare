package server

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	pbConnections "github.com/ACLzz/go-qshare/protobuf/gen/connections"
	pbSecuregcm "github.com/ACLzz/go-qshare/protobuf/gen/securegcm"
)

const max_message_length = 32 * 1024 // 32 Kb

type commServer struct {
	sock   net.Listener
	wg     *sync.WaitGroup
	stopCh chan struct{}
	// TODO: use kind of logger instead of prints
}

func newCommServer(port int) (commServer, error) {
	sock, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return commServer{}, fmt.Errorf("start socket: %w", err)
	}

	return commServer{
		sock:   sock,
		wg:     &sync.WaitGroup{},
		stopCh: make(chan struct{}, 1),
	}, nil
}

func (cs commServer) Listen() {
	var (
		conn net.Conn
		err  error
	)
	cs.wg.Add(1)
	defer cs.wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		conn, err = cs.sock.Accept()
		if err != nil {
			select {
			case <-cs.stopCh:
				return
			default:
				fmt.Printf("ERR: accept conn: %s", err.Error())
			}
			continue
		}

		cs.wg.Add(1)
		c := newCommConn(conn)
		go c.accept(ctx, cs.wg) // TODO: maybe limit amount of gorutines here
	}
}

type (
	commConn struct {
		conn  net.Conn
		phase phase

		// TODO: leave only necessary info from messages in conn, delete full messages
		connRequest        *pbConnections.ConnectionRequestFrame
		ukeyClientInit     *pbSecuregcm.Ukey2ClientInit
		ukeyClientFinished *pbSecuregcm.Ukey2ClientFinished
		connResponse       *pbConnections.ConnectionRequestFrame
	}

	phase uint8
)

const (
	estSecureConnPhase phase = iota
	pairingPhase
	// ...
)

func newCommConn(conn net.Conn) commConn {
	cConn := commConn{
		conn:  conn,
		phase: estSecureConnPhase,
	}

	return cConn
}

func (cc *commConn) accept(ctx context.Context, wg *sync.WaitGroup) {
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

			fmt.Printf("message length: %d\n", msgLen)
			msgBuf := make([]byte, msgLen)

			// fetch message in bytes
			if _, err = cc.conn.Read(msgBuf); err != nil {
				fmt.Printf("ERR: fetch message: %s", err.Error())
				continue
			}

			// process message depending on current phase
			switch cc.phase {
			case estSecureConnPhase:
				err = cc.establishSecureConnection(msgBuf)
			}

			if err != nil {
				fmt.Printf("ERR: process message: %s", err.Error())
				continue
			}
		}
	}
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

func (cs commServer) Stop() error {
	if cs.sock == nil {
		return nil
	}

	cs.stopCh <- struct{}{}
	if err := cs.sock.Close(); err != nil {
		return fmt.Errorf("close socket: %w", err)
	}
	cs.wg.Wait()

	return nil
}
