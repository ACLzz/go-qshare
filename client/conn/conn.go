package conn

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"sync"

	qshare "github.com/ACLzz/go-qshare"
	"github.com/ACLzz/go-qshare/internal/adapter"
	"github.com/ACLzz/go-qshare/internal/crypt"
	"github.com/ACLzz/go-qshare/log"
)

const max_text_title_length = 12

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
	// send introduction
	var title string
	if len(text) > max_text_title_length {
		title = text[:max_text_title_length-3]
		title += "..."
	} else {
		title = text
	}

	if err := c.adapter.SendIntroduction(adapter.IntroductionFrame{
		Text: &qshare.TextPayload{
			Type:  qshare.TextText,
			Title: title,
			Size:  int64(len(text)),
		},
	}); err != nil {
		return fmt.Errorf("send introduction message: %w", err)
	}

	// send transfer request
	if err := c.adapter.SendTransferRequest(); err != nil {
		return fmt.Errorf("send transfer request: %w", err)
	}

	// process server response
	c.log.Debug("waiting for server response...")
	read := c.adapter.Reader(ctx)
	msg, err := c.read(read)
	if err != nil {
		return fmt.Errorf("waiting for transfer response: %w", err)
	}

	isAccepted, err := c.adapter.UnmarshalTransferResponse(msg)
	if err != nil {
		return fmt.Errorf("unmarshal transfer response: %w", err)
	}
	if !isAccepted {
		c.adapter.Disconnect()
		return adapter.ErrConnWasEndedByClient
	}
	c.log.Debug("server accepted transfer")

	// send random data
	c.adapter.SendDataInChunks(rand.Int64(), []byte("random shit"))

	// send text
	if err := c.adapter.SendDataInChunks(rand.Int64(), []byte(text)); err != nil {
		return fmt.Errorf("send data in chunks")
	}

	// disconnect
	c.log.Debug("success transer, disconnecting...")
	c.adapter.Disconnect()
	return nil
}

func (c *Connection) SendFiles() error {
	defer c.wg.Done()
	return nil
}

func (c *Connection) SetupTransfer(ctx context.Context) error {
	var err error
	read := c.adapter.Reader(ctx)

	// init phase
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

	msg, err := c.read(read)
	if err != nil {
		return fmt.Errorf("waiting for server init: %w", err)
	}

	if err = c.adapter.ValidateServerInit(msg); err != nil {
		return fmt.Errorf("validate server init: %w", err)
	}

	if err = c.adapter.SendConnResponse(true); err != nil {
		return fmt.Errorf("send connection response: %w", err)
	}

	msg, err = c.read(read)
	if err != nil {
		return fmt.Errorf("wait for conn response: %w", err)
	}

	connResponse, err := c.adapter.UnmarshalConnResponse(msg)
	if err != nil {
		return fmt.Errorf("unmarshal conn response: %w", err)
	}

	if !connResponse.IsConnAccepted {
		c.adapter.Disconnect()
		return adapter.ErrConnWasEndedByClient
	}
	if err = c.adapter.EnableEncryption(); err != nil {
		return fmt.Errorf("enable encryption: %w", err)
	}

	// pairing phase

	if err = c.adapter.SendPairedKeyEncryption(); err != nil {
		return fmt.Errorf("send paired key encryption: %w", err)
	}
	msg, err = c.read(read)
	if err != nil {
		return fmt.Errorf("wait for paired key encryption: %w", err)
	}
	if err := c.adapter.ValidatePairedKeyEncryption(msg); err != nil {
		return fmt.Errorf("validate paired key encryption: %w", err)
	}

	if err = c.adapter.SendPairedKeyResult(); err != nil {
		return fmt.Errorf("send paired key result: %w", err)
	}
	msg, err = c.read(read)
	if err != nil {
		return fmt.Errorf("waiting for paired key result: %w", err)
	}
	if err := c.adapter.ValidatePairedKeyResult(msg); err != nil {
		return fmt.Errorf("validate paired key result: %w", err)
	}

	c.log.Debug("success transfer setup")
	return nil
}

func (c *Connection) read(read func() ([]byte, error)) ([]byte, error) {
	for {
		msg, err := read()
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, adapter.ErrMessageTooLong) {
				c.adapter.Disconnect()
			} else if errors.Is(err, adapter.ErrOffsetMismatch) {
				c.adapter.SendBadMessageError()
			} else if errors.Is(err, adapter.ErrTransferInProgress) {
				continue
			}

			c.log.Error("read message", err)
			return nil, err
		}

		return msg, nil
	}
}
