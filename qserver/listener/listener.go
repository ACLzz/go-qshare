package listener

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/ACLzz/qshare"
	"github.com/ACLzz/qshare/internal/rand"
)

type Listener struct {
	sock         net.Listener
	wg           *sync.WaitGroup
	log          qshare.Logger
	authCallback qshare.AuthCallback
	textCallback qshare.TextCallback
	fileCallback qshare.FileCallback
	stopCh       chan struct{}
	rand         rand.Random
}

func New(
	port int,
	log qshare.Logger,
	authCallback qshare.AuthCallback,
	textCallback qshare.TextCallback,
	fileCallback qshare.FileCallback,
	r rand.Random,
) (Listener, error) {
	sock, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return Listener{}, fmt.Errorf("start socket: %w", err)
	}

	return Listener{
		sock:         sock,
		wg:           &sync.WaitGroup{},
		log:          log,
		authCallback: authCallback,
		textCallback: textCallback,
		fileCallback: fileCallback,
		stopCh:       make(chan struct{}, 1),
		rand:         r,
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
		c := newConnection(ctx, conn, l.log, l.authCallback, l.textCallback, l.fileCallback, l.rand)
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
