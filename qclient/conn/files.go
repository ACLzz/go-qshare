package conn

import (
	"context"
	"fmt"

	"github.com/ACLzz/qshare"
)

func (c *Connection) SendFiles(ctx context.Context, files []qshare.FilePayload) error {
	// send introduction
	var filePayloads = map[int64]*qshare.FilePayload{}
	for i := range files {
		filePayloads[c.rand.Int64()] = &files[i]
	}

	if err := c.sendTransferRequest(ctx, nil, &filePayloads); err != nil {
		return fmt.Errorf("send transfer request: %w", err)
	}

	// send files
	for id := range filePayloads {
		if err := c.adapter.SendFileInChunks(id, *filePayloads[id]); err != nil {
			return fmt.Errorf("send file: %w", err)
		}
	}

	c.log.Debug("success files transfer")
	return nil
}
