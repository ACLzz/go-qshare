package listener

import (
	"context"
	"fmt"
	"net"
	"sync"

	qshare "github.com/ACLzz/go-qshare"
	"github.com/ACLzz/go-qshare/log"
)

type Listener struct {
	sock         net.Listener
	wg           *sync.WaitGroup
	log          log.Logger
	textCallback qshare.TextCallback
	fileCallback qshare.FileCallback
	stopCh       chan struct{}
}

func New(
	port int,
	log log.Logger,
	textCallback qshare.TextCallback,
	fileCallback qshare.FileCallback,
) (Listener, error) {
	sock, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return Listener{}, fmt.Errorf("start socket: %w", err)
	}

	return Listener{
		sock:         sock,
		wg:           &sync.WaitGroup{},
		log:          log,
		textCallback: textCallback,
		fileCallback: fileCallback,
		stopCh:       make(chan struct{}, 1),
	}, nil
}

func (l Listener) Listen() {
	var (
		conn net.Conn
		err  error
	)
	l.wg.Add(1)
	defer l.wg.Done()
	l.log.Debug("listener started", "addr", l.sock.Addr().String())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		conn, err = l.sock.Accept()
		if err != nil {
			select {
			case <-l.stopCh:
				return
			default:
				l.log.Error("accept conn", err)
			}
			continue
		}

		l.wg.Add(1)
		c := newConnection(ctx, conn, l.log, l.textCallback, l.fileCallback)
		go c.Accept(l.wg) // TODO: maybe limit amount of gorutines here
	}
}

func (l Listener) Stop() error {
	l.log.Debug("stopping listener...")
	if l.sock == nil {
		return nil
	}

	l.stopCh <- struct{}{}
	if err := l.sock.Close(); err != nil {
		return fmt.Errorf("close socket: %w", err)
	}
	l.wg.Wait()

	l.log.Debug("listener stopped")
	return nil
}
