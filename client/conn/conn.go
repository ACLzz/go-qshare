package conn

import (
	"context"
	"fmt"
	"net"
	"sync"

	qshare "github.com/ACLzz/go-qshare"
	"github.com/ACLzz/go-qshare/internal/adapter"
	"github.com/ACLzz/go-qshare/internal/crypt"
	"github.com/ACLzz/go-qshare/log"
)

type Connection struct {
	cipher  *crypt.Cipher
	conn    net.Conn
	log     log.Logger
	adapter adapter.Adapter
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
		adapter: adapter.New(conn, logger, &cipher, nil, nil),
		wg:      wg,
		cfg:     cfg,
	}
}

func (c *Connection) SendText(ctx context.Context, text string) error {
	defer c.wg.Done()

	return nil
}

func (c *Connection) SendFiles() error {
	defer c.wg.Done()
	return nil
}

func (c *Connection) SetupTransfer(ctx context.Context) error {
	var err error
	read := c.adapter.Reader(ctx)

	if err = c.adapter.SendConnRequest(
		c.cfg.Hostname,
		c.cfg.EndpointID,
		c.cfg.DeviceType,
	); err != nil {
		return fmt.Errorf("send connection request: %w", err)
	}

	if err = c.adapter.SendClientInitWithClientFinished(); err != nil {
		return fmt.Errorf("send client init with client finished: %w", err)
	}

	msg, err := read()
	if err != nil {
		return err
	}
	_ = msg

	c.log.Debug("transfer set up", "lastMsgLength", len(msg))
	return nil
}
