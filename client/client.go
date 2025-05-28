package client

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/ACLzz/go-qshare/client/conn"
	"github.com/ACLzz/go-qshare/internal/mdns"
	"github.com/ACLzz/go-qshare/log"
	"github.com/grandcat/zeroconf"
)

const default_output_channel_size = 64

type Client struct {
	wg  *sync.WaitGroup
	log log.Logger
}

func (c *Client) ListServers(ctx context.Context) (chan ServerInstance, error) {
	resolv, err := zeroconf.NewResolver()
	if err != nil {
		return nil, fmt.Errorf("create mDNS resolver: %w", err)
	}

	entriesCh := make(chan *zeroconf.ServiceEntry)
	if err := resolv.Browse(ctx, mdns.MDNsServiceType, "*", entriesCh); err != nil {
		return nil, fmt.Errorf("browse available servers: %w", err)
	}

	outputCh := make(chan ServerInstance, default_output_channel_size)
	go c.listServersWorker(ctx, entriesCh, outputCh)

	return outputCh, nil
}

func (c *Client) listServersWorker(ctx context.Context, entriesCh chan *zeroconf.ServiceEntry, outputCh chan ServerInstance) {
	var entry *zeroconf.ServiceEntry
	defer close(entriesCh)
	defer close(outputCh)

	for {
		select {
		case entry = <-entriesCh:
			outputCh <- newServerInstance(entry)
			continue
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) SendText(instance ServerInstance, text string) error {
	cn, err := c.newConn(instance)
	if err != nil {
		return fmt.Errorf("create connection: %w", err)
	}

	c.wg.Add(1)
	go cn.SendText(text)
	return nil
}

func (c *Client) SendFile(pw *io.PipeWriter) error {
	return nil
}

func (c *Client) Close() {
	// TODO: send cancel to every connection
	c.wg.Wait()
}

func (c *Client) newConn(instance ServerInstance) (conn.Connection, error) {
	if len(instance.entry.AddrIPv4) == 0 {
		return conn.Connection{}, ErrInvalidServerInstance
	}

	addr := fmt.Sprintf("%s:%d", instance.entry.AddrIPv4[0].String(), instance.entry.Port)
	netAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return conn.Connection{}, fmt.Errorf("resolve server address: %w", err)
	}

	cn, err := net.DialTCP("tcp", nil, netAddr)
	if err != nil {
		return conn.Connection{}, fmt.Errorf("dial server: %w", err)
	}

	return conn.NewConnection(cn, c.log, conn.Config{
		EndpointID: instance.endpoint,
		Hostname:   instance.Hostname,
	}, c.wg), nil
}
