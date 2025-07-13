package qclient

import (
	"fmt"
	"net"

	"github.com/ACLzz/qshare"
	"github.com/ACLzz/qshare/internal/rand"
	"github.com/ACLzz/qshare/qclient/conn"
)

// Client discover, creates and manages connections to other servers.
type Client struct {
	log        qshare.Logger
	endpointID string // mdns endpoint id
	rand       rand.Random
}

func (c *Client) newConn(instance ServerInstance) (conn.Connection, error) {
	if len(instance.entry.AddrIPv4) == 0 {
		return conn.Connection{}, ErrInvalidServerInstance
	}

	addr := fmt.Sprintf("%s:%d", instance.entry.AddrIPv4[0].String(), instance.entry.Port)
	netAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return conn.Connection{}, fmt.Errorf("resolve server address: %w", err)
	}

	cn, err := net.DialTCP("tcp", nil, netAddr)
	if err != nil {
		return conn.Connection{}, fmt.Errorf("dial server: %w", err)
	}

	return conn.NewConnection(cn, c.log, conn.Config{
		EndpointID: c.endpointID,
		Hostname:   instance.Hostname,
	}, c.rand), nil
}
