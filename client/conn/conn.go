package conn

import (
	"fmt"
	"net"
	"sync"

	qshare "github.com/ACLzz/go-qshare"
	"github.com/ACLzz/go-qshare/internal/comm"
	"github.com/ACLzz/go-qshare/internal/crypt"
	"github.com/ACLzz/go-qshare/log"
)

type Connection struct {
	cipher  *crypt.Cipher
	conn    net.Conn
	log     log.Logger
	adapter comm.Adapter
	wg      *sync.WaitGroup
	cfg     Config
}

type Config struct {
	EndpointID string
	Hostname   string
	DeviceType qshare.DeviceType
}

func NewConnection(conn net.Conn, logger log.Logger, cfg Config, wg *sync.WaitGroup) Connection {
	cipher := crypt.NewCipher(false)

	return Connection{
		cipher:  &cipher,
		conn:    conn,
		log:     logger,
		adapter: comm.NewAdapter(conn, logger, &cipher, nil, nil),
		wg:      wg,
		cfg:     cfg,
	}
}

func (c *Connection) SendText(text string) error {
	defer c.wg.Done()
	if err := c.adapter.SendConnRequest(
		c.cfg.Hostname,
		c.cfg.EndpointID,
		c.cfg.DeviceType,
	); err != nil {
		return fmt.Errorf("send connection request: %w", err)
	}

	return nil
}

func (c *Connection) SendFiles() error {
	defer c.wg.Done()
	return nil
}
