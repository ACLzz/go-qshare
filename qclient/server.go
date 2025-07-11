package qclient

import (
	"github.com/grandcat/zeroconf"
)

type ServerInstance struct {
	entry    *zeroconf.ServiceEntry
	endpoint string
	Hostname string
}
