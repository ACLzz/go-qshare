package qclient

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/ACLzz/qshare/internal/mdns"
	"github.com/grandcat/zeroconf"
)

const default_output_channel_size = 64

type ServerInstance struct {
	entry    *zeroconf.ServiceEntry
	endpoint string
	Hostname string
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
		case <-ctx.Done():
			return
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
		}
	}
}
