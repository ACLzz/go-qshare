package qclient

import (
	"context"
	"fmt"

	"github.com/ACLzz/qshare"
)

func (c *Client) SendFiles(
	ctx context.Context,
	instance ServerInstance,
	files []qshare.FilePayload,
) error {
	if len(files) == 0 {
		return nil
	}

	cn, err := c.newConn(instance)
	if err != nil {
		return fmt.Errorf("create connection: %w", err)
	}
	defer cn.Disconnect()

	if err := cn.SetupTransfer(ctx); err != nil {
		return fmt.Errorf("setup transfer: %w", err)
	}

	if err := cn.SendFiles(ctx, files); err != nil {
		return fmt.Errorf("send files: %w", err)
	}

	return nil
}
