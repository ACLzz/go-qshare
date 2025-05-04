package server

import (
	"fmt"
	"net"
)

type commServer struct {
	sock net.Listener
	// TODO: use kind of logger instead of prints
}

func newCommServer(port int) (commServer, error) {
	sock, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return commServer{}, fmt.Errorf("start socket: %w", err)
	}

	return commServer{
		sock: sock,
	}, nil
}

func (cs commServer) Listen() {
	for {
		conn, err := cs.sock.Accept()
		if err != nil {
			fmt.Printf("ERR: accept conn: %s", err.Error())
			continue
		}
		go cs.handle(conn) // TODO: maybe limit amount of gorutines here
	}
}

func (cs commServer) handle(conn net.Conn) {
	var (
		n   int
		err error
		buf = make([]byte, 0)
	)

	for {
		if n, err = conn.Read(buf); err != nil {
			fmt.Printf("ERR: read conn: %s", err.Error())
			continue
		}

		fmt.Printf("recv (%d): '%s'", len(buf), string(buf))

		if n == 0 {
			break
		}

		if _, err := conn.Write(buf); err != nil {
			fmt.Printf("ERR: write conn: %s", err.Error())
		}
	}

	if err = conn.Close(); err != nil {
		fmt.Printf("ERR: close conn: %s", err.Error())
	}
}

func (cs commServer) Stop() error {
	if cs.sock == nil {
		return nil
	}

	// TODO: gracefully close all conns and exit from Listen loop
	if err := cs.sock.Close(); err != nil {
		return fmt.Errorf("close socket: %w", err)
	}

	return nil
}
