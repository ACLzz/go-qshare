package listener

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"

	"github.com/ACLzz/qshare"
	adapter "github.com/ACLzz/qshare/internal/adapter"
	"github.com/ACLzz/qshare/internal/rand"
)

type filePayload struct {
	Pd         qshare.FilePayload
	IsNotified bool
}

type connection struct {
	ctx       context.Context
	cancelCtx context.CancelFunc
	adapter   adapter.Adapter
	conn      net.Conn
	log       qshare.Logger

	nextExpectedMessage expectedMessage
	phase               phase

	textMeta         *adapter.TextMeta // text payload can be sent only one
	filePayloads     map[int64]*filePayload
	expectedPayloads int
	receivedPayloads int
	authCallback     qshare.AuthCallback
	textCallback     qshare.TextCallback
	fileCallback     qshare.FileCallback
}

func newConnection(
	ctx context.Context,
	conn net.Conn,
	logger qshare.Logger,
	authCallback qshare.AuthCallback,
	textCallback qshare.TextCallback,
	fileCallback qshare.FileCallback,
	r rand.Random,
) *connection {
	connCtx, cancel := context.WithCancel(ctx)
	c := &connection{
		ctx:                 connCtx,
		cancelCtx:           cancel,
		conn:                conn,
		log:                 logger,
		nextExpectedMessage: conn_request,
		phase:               init_phase,
		authCallback:        authCallback,
		textCallback:        textCallback,
		fileCallback:        fileCallback,
	}
	c.adapter = adapter.New(conn, logger, true, c.writeFileChunk, c.writeText, r)

	return c
}

func (c *connection) Accept(wg *sync.WaitGroup) {
	defer wg.Done()

	c.log.Debug("got new connection")
	defer c.log.Debug("connection closed")

	readMsg := c.adapter.Reader(c.ctx)
	for {
		// read the message
		msg, err := readMsg()
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, adapter.ErrMessageTooLong) {
				c.adapter.Disconnect()
				return
			} else if errors.Is(err, adapter.ErrOffsetMismatch) {
				c.adapter.SendBadMessageError()
			} else if errors.Is(err, adapter.ErrConnWasEndedByClient) {
				return
			}

			c.log.Error("read message", err)
			continue
		}

		// route the message
		c.route(msg)
	}
}
