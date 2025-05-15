package comm

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/ACLzz/go-qshare/log"
)

type Server struct {
	sock   net.Listener
	wg     *sync.WaitGroup
	log    log.Logger
	stopCh chan struct{}
	// TODO: use kind of logger instead of prints
}

func NewServer(port int, log log.Logger) (Server, error) {
	sock, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return Server{}, fmt.Errorf("start socket: %w", err)
	}

	return Server{
		sock:   sock,
		wg:     &sync.WaitGroup{},
		log:    log,
		stopCh: make(chan struct{}, 1),
	}, nil
}

func (s Server) Listen() {
	var (
		conn net.Conn
		err  error
	)
	s.wg.Add(1)
	defer s.wg.Done()
	s.log.Debug("communication server started", "addr", s.sock.Addr().String())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		conn, err = s.sock.Accept()
		if err != nil {
			select {
			case <-s.stopCh:
				return
			default:
				s.log.Error("accept conn", err)
			}
			continue
		}

		s.wg.Add(1)
		c := newCommConn(conn, s.log)
		go c.Accept(ctx, s.wg) // TODO: maybe limit amount of gorutines here
	}
}

func (s Server) Stop() error {
	s.log.Debug("stopping communication server...")
	if s.sock == nil {
		return nil
	}

	s.stopCh <- struct{}{}
	if err := s.sock.Close(); err != nil {
		return fmt.Errorf("close socket: %w", err)
	}
	s.wg.Wait()

	s.log.Debug("communication server stopped")
	return nil
}
