package goqshare

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/mdns"
	"tinygo.org/x/bluetooth"
)

type Server struct {
	mDNSService *mdns.MDNSService
	ad          *bluetooth.Advertisement

	mDNSServer *mdns.Server
}

func (s *Server) Listen() error {
	mDNSServer, err := mdns.NewServer(&mdns.Config{Zone: s.mDNSService})
	if err != nil {
		return fmt.Errorf("set up mDNS server: %w", err)
	}
	s.mDNSServer = mDNSServer

	if err := s.ad.Start(); err != nil {
		log.Default().Printf("ble advertisements not started: %s", err.Error())
	}
	return nil
}

func (s *Server) Stop() (gErr error) {
	if err := s.mDNSServer.Shutdown(); err != nil {
		gErr = fmt.Errorf("%w: shutdown mDNS server: %w", gErr, err)
	}

	if err := s.ad.Stop(); err != nil && !strings.Contains(err.Error(), "advertisement is not started") {
		gErr = fmt.Errorf("%w: stop ble advertisements: %w", gErr, err)
	}

	return gErr
}
