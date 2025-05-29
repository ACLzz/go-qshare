package listener

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"

	qshare "github.com/ACLzz/go-qshare"
	adapter "github.com/ACLzz/go-qshare/internal/adapter"
	"github.com/ACLzz/go-qshare/internal/crypt"
	"github.com/ACLzz/go-qshare/log"
)

type connection struct {
	ctx       context.Context
	cancelCtx context.CancelFunc
	adapter   adapter.Adapter
	conn      net.Conn
	log       log.Logger

	nextExpectedMessage expectedMessage
	phase               phase

	textPayload      *qshare.TextPayload // text payload can be sent only one
	filePayloads     map[int64]*adapter.FilePayload
	expectedPayloads int
	receivedPayloads int
	textCallback     qshare.TextCallback
	fileCallback     qshare.FileCallback
}

func newConnection(
	ctx context.Context,
	conn net.Conn,
	logger log.Logger,
	textCallback qshare.TextCallback,
	fileCallback qshare.FileCallback,
) *connection {
	cipher := crypt.NewCipher(true)
	connCtx, cancel := context.WithCancel(ctx)
	c := &connection{
		ctx:                 connCtx,
		cancelCtx:           cancel,
		conn:                conn,
		log:                 logger,
		nextExpectedMessage: conn_request,
		phase:               init_phase,
		textCallback:        textCallback,
		fileCallback:        fileCallback,
	}
	c.adapter = adapter.New(conn, logger, &cipher, c.writeFileChunk, c.writeText)

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
			} else if errors.Is(err, adapter.ErrTransferInProgress) {
				continue
			}

			c.log.Error("read message", err)
			continue
		}

		// route the message
		c.route(msg)
	}
}
