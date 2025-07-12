package conn

import (
	"context"
	"fmt"

	"github.com/ACLzz/qshare"
	"github.com/ACLzz/qshare/internal/adapter"
)

func (c *Connection) SendText(ctx context.Context, text string, t qshare.TextType) error {
	// send transfer request
	payloadID := c.rand.Int64()
	title := text
	if len(text) > max_title_length {
		title = text[:max_title_length-3] + "..."
	}

	if err := c.sendTransferRequest(ctx, adapter.NewTextMeta(
		payloadID,
		t,
		title,
		int64(len(text)),
	), nil); err != nil {
		return fmt.Errorf("send transfer request: %w", err)
	}

	// send text
	if err := c.adapter.SendDataInChunks(payloadID, []byte(text)); err != nil {
		return fmt.Errorf("send data in chunks")
	}

	c.log.Debug("success text transfer")
	return nil
}
