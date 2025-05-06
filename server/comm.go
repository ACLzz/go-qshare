package server

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/ACLzz/go-qshare/protobuf/gen/connections"
	"google.golang.org/protobuf/proto"
)

type commServer struct {
	sock   net.Listener
	wg     *sync.WaitGroup
	stopCh chan struct{}
	// TODO: use kind of logger instead of prints
}

const batchSize = 256

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
		go cs.handleConn(ctx, conn) // TODO: maybe limit amount of gorutines here
	}
}

func (cs commServer) handleConn(ctx context.Context, conn net.Conn) {
	var (
		n        int
		err      error
		buf      bytes.Buffer
		batchBuf = make([]byte, batchSize)
	)
	defer cs.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if n, err = conn.Read(batchBuf); err != nil {
				fmt.Printf("ERR: read conn: %s", err.Error())
				continue
			}

			buf.Write(batchBuf)
			if n < batchSize {
				var frame connections.ConnectionRequestFrame
				if err := proto.Unmarshal(buf.Bytes(), &frame); err != nil {
					fmt.Printf("ERR: unmarshal: %s", err.Error())
				}

				fmt.Println(frame.String())
				return
			}
		}
	}
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
