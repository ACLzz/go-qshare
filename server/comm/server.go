package comm

import (
	"context"
	"fmt"
	"net"
	"sync"
)

type Server struct {
	sock   net.Listener
	wg     *sync.WaitGroup
	stopCh chan struct{}
	// TODO: use kind of logger instead of prints
}

func NewServer(port int) (Server, error) {
	sock, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return Server{}, fmt.Errorf("start socket: %w", err)
	}

	return Server{
		sock:   sock,
		wg:     &sync.WaitGroup{},
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		conn, err = s.sock.Accept()
		if err != nil {
			select {
			case <-s.stopCh:
				return
			default:
				fmt.Printf("ERR: accept conn: %s", err.Error())
			}
			continue
		}

		s.wg.Add(1)
		c := newCommConn(conn)
		go c.Accept(ctx, s.wg) // TODO: maybe limit amount of gorutines here
	}
}

func (s Server) Stop() error {
	if s.sock == nil {
		return nil
	}

	s.stopCh <- struct{}{}
	if err := s.sock.Close(); err != nil {
		return fmt.Errorf("close socket: %w", err)
	}
	s.wg.Wait()

	return nil
}
