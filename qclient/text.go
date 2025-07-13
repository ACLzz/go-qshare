package qclient

import (
	"context"
	"fmt"

	"github.com/ACLzz/qshare"
)

// Establishes connection to specified server and sends provided text.
func (c *Client) SendText(
	ctx context.Context,
	instance ServerInstance,
	text string,
	t qshare.TextType,
) error {
	if t.IsUnknown() {
		return ErrInvalidTextType
	}

	cn, err := c.newConn(instance)
	if err != nil {
		return fmt.Errorf("create connection: %w", err)
	}
	defer cn.Disconnect()

	if err := cn.SetupTransfer(ctx); err != nil {
		return fmt.Errorf("setup transfer: %w", err)
	}

	if err := cn.SendText(ctx, text, t); err != nil {
		return fmt.Errorf("send text: %w", err)
	}

	return nil
}
