package client

import (
	"io"

	"github.com/ACLzz/go-qshare/internal/comm"
	"github.com/ACLzz/go-qshare/internal/crypt"
)

type Client struct {
	cipher  *crypt.Cipher
	adapter comm.Adapter
}

func NewClient() Client {
	cipher := crypt.NewCipher(false)
	return Client{
		cipher: &cipher,
	}
}

func (c *Client) ListServers() ([]string, error) {
	return nil, nil
}

func (c *Client) SendText(text string) error {
	// net.DialTCP()
	return nil
}

func (c *Client) SendFile(pw *io.PipeWriter) error {
	return nil
}
