package conn

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/ACLzz/qshare"
	"github.com/ACLzz/qshare/internal/adapter"
	"github.com/ACLzz/qshare/internal/rand"
)

const max_title_length = 12

type Connection struct {
	conn    net.Conn
	log     qshare.Logger
	adapter adapter.Adapter
	wg      *sync.WaitGroup // TODO: maybe remove
	cfg     Config
	rand    rand.Random
}

type Config struct {
	EndpointID string
	Hostname   string
	DeviceType qshare.DeviceType
}

func NewConnection(conn net.Conn, logger qshare.Logger, cfg Config, wg *sync.WaitGroup, r rand.Random) Connection {
	return Connection{
		conn:    conn,
		log:     logger,
		adapter: adapter.New(conn, logger, false, nil, nil, r),
		wg:      wg,
		cfg:     cfg,
		rand:    r,
	}
}

func (c *Connection) SendText(ctx context.Context, text string) error {
	// send introduction
	var title string
	if len(text) > max_title_length {
		title = text[:max_title_length-3]
		title += "..."
	} else {
		title = text
	}
	textPayloadID := c.rand.Int64()

	if err := c.adapter.SendIntroduction(
		adapter.IntroductionFrame{
			Text: adapter.NewTextMeta(
				textPayloadID,
				qshare.TextText,
				title,
				int64(len(text)),
			),
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
	if err := c.adapter.SendDataInChunks(c.rand.Int64(), []byte("random")); err != nil {
		return fmt.Errorf("send random data: %w", err)
	}

	// send text
	if err := c.adapter.SendDataInChunks(textPayloadID, []byte(text)); err != nil {
		return fmt.Errorf("send data in chunks")
	}

	// disconnect
	c.log.Debug("success transer, disconnecting...")
	c.adapter.Disconnect() // TODO: move to a separate method so that client can use it on it's own
	return nil
}

func (c *Connection) SendFiles(ctx context.Context, files []qshare.FilePayload) error {
	if len(files) == 0 {
		return nil
	}

	// send introduction
	var filePayloads = map[int64]*qshare.FilePayload{}
	for i := range files {
		filePayloads[c.rand.Int64()] = &files[i]
	}

	if err := c.adapter.SendIntroduction(
		adapter.IntroductionFrame{
			Files: filePayloads,
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
	if err := c.adapter.SendDataInChunks(c.rand.Int64(), []byte("random")); err != nil {
		return fmt.Errorf("send random data: %w", err)
	}

	// send files
	for id := range filePayloads {
		if err := c.adapter.SendFileInChunks(id, *filePayloads[id]); err != nil {
			return fmt.Errorf("send file: %w", err)
		}
	}

	// disconnect
	c.log.Debug("success transer, disconnecting...")
	c.adapter.Disconnect() // TODO: move to a separate method so that client can use it on it's own
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
	msg, err := read()
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, adapter.ErrMessageTooLong) {
			c.adapter.Disconnect()
		} else if errors.Is(err, adapter.ErrOffsetMismatch) {
			c.adapter.SendBadMessageError()
		}

		c.log.Error("read message", err)
		return nil, err
	}

	return msg, nil
}
