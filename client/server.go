package client

import (
	"github.com/grandcat/zeroconf"
)

type ServerInstance struct {
	entry    *zeroconf.ServiceEntry
	endpoint string
	Hostname string
}

func newServerInstance(entry *zeroconf.ServiceEntry) ServerInstance {
	// endpoint, err := base64.RawURLEncoding.DecodeString(entry.Instance)
	return ServerInstance{
		entry:    entry,
		Hostname: entry.HostName,
	}
}
