package qclient

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"sync"

	"github.com/ACLzz/qshare"
	"github.com/ACLzz/qshare/internal/mdns"
	"github.com/ACLzz/qshare/internal/rand"
	"github.com/ACLzz/qshare/qclient/conn"
	"github.com/grandcat/zeroconf"
)

const default_output_channel_size = 64

type Client struct {
	wg         *sync.WaitGroup // TODO: review if we need it at all
	log        qshare.Logger
	endpointID string // mdns endpoint id
	rand       rand.Random
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

func (c *Client) listServersWorker(
	ctx context.Context,
	entriesCh chan *zeroconf.ServiceEntry,
	outputCh chan ServerInstance,
) {
	var entry *zeroconf.ServiceEntry
	defer close(entriesCh)
	defer close(outputCh)

	for {
		select {
		case entry = <-entriesCh:
			endpoint, err := base64.RawURLEncoding.DecodeString(entry.Instance)
			if err != nil {
				c.log.Error("decode server instance endpoint", err)
			}
			outputCh <- ServerInstance{
				entry:    entry,
				endpoint: string(endpoint),
				Hostname: entry.HostName,
			}
			continue
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) SendText(ctx context.Context, instance ServerInstance, text string) error {
	cn, err := c.newConn(instance)
	if err != nil {
		return fmt.Errorf("create connection: %w", err)
	}

	if err := cn.SetupTransfer(ctx); err != nil {
		return fmt.Errorf("setup transfer: %w", err)
	}

	if err := cn.SendText(ctx, text); err != nil {
		return fmt.Errorf("send text: %w", err)
	}

	return nil
}

func (c *Client) SendFiles(
	ctx context.Context,
	instance ServerInstance,
	files []qshare.FilePayload,
) error {
	cn, err := c.newConn(instance)
	if err != nil {
		return fmt.Errorf("create connection: %w", err)
	}

	if err := cn.SetupTransfer(ctx); err != nil {
		return fmt.Errorf("setup transfer: %w", err)
	}

	if err := cn.SendFiles(ctx, files); err != nil {
		return fmt.Errorf("send files: %w", err)
	}

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
		EndpointID: c.endpointID,
		Hostname:   instance.Hostname,
	}, c.wg, c.rand), nil
}
