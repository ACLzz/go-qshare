package listener

import (
	"fmt"

	adapter "github.com/ACLzz/go-qshare/internal/adapter"
)

func (c *connection) processIntroduction(msg []byte) error {
	intro, err := c.adapter.UnmarshalIntroduction(msg)
	if err != nil {
		return err
	}

	if intro.HasText() {
		c.textPayload = intro.Text
		c.expectedPayloads++
	}
	if intro.HasFiles() {
		c.filePayloads = intro.Files
		c.expectedPayloads += len(c.filePayloads)
	}

	return nil
}

func (c *connection) processTransferRequest(msg []byte) error {
	if err := c.adapter.ValidateTransferRequest(msg); err != nil {
		return err
	}

	return c.adapter.SendTransferResponse(true)
}

func (c *connection) writeFileChunk(chunk adapter.FileChunk) error {
	file := c.filePayloads[chunk.FileID]
	if !file.IsNotified {
		c.filePayloads[chunk.FileID].IsNotified = true
		c.fileCallback(file.FilePayload, file.Pr)
	}

	if len(chunk.Body) > 0 {
		n, err := file.Pw.Write(chunk.Body)
		if err != nil {
			return fmt.Errorf("write chunk to pipe: %w", err)
		}

		if n != len(chunk.Body) {
			return ErrInternalError // TODO: another error
		}

		c.filePayloads[chunk.FileID].BytesSent += int64(n)
	}

	if chunk.IsFinalChunk {
		file.Pw.Close()
		c.log.Debug("file transfered", "filename", file.Title)

		c.receivedPayloads++
		c.checkIfLastPayload()
	}

	return nil
}

func (c *connection) writeText(text string) error {
	if c.textPayload == nil || c.textCallback == nil {
		return ErrTextTransferNotExpected
	}

	c.textCallback(c.textPayload.TextPayload, text)
	c.receivedPayloads++
	c.checkIfLastPayload()
	return nil
}

func (c *connection) checkIfLastPayload() {
	if c.receivedPayloads >= c.expectedPayloads {
		c.log.Debug("got last payload, disconnecting...")
		c.cancelCtx()
	}
}
