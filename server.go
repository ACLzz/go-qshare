package goqshare

import (
	"fmt"
	"strings"

	"github.com/grandcat/zeroconf"
	"tinygo.org/x/bluetooth"
)

type (
	Server struct {
		mDNSConf mDNSConfig
		bleAD    *bluetooth.Advertisement

		mDNSServer *zeroconf.Server
	}

	mDNSConfig struct {
		name string
		port int
		txt  string
	}
)

const MDNsServiceType = "_FC9F5ED42C8A._tcp"

func (s *Server) Listen() error {
	if err := s.listen(); err != nil {
		if stopErr := s.Stop(); stopErr != nil {
			return fmt.Errorf("stopping server on failed start: %w: %w", stopErr, err)
		}

		return err
	}

	return nil
}

func (s *Server) listen() error {
	var err error
	s.mDNSServer, err = zeroconf.Register(
		s.mDNSConf.name,
		MDNsServiceType,
		"local.",
		s.mDNSConf.port,
		[]string{"n=" + s.mDNSConf.txt},
		nil,
	)
	if err != nil {
		return fmt.Errorf("start mDNS server: %w", err)
	}

	if err := s.bleAD.Start(); err != nil {
		return fmt.Errorf("start ble advertisement: %w", err)
	}

	return nil
}

func (s *Server) Stop() error {
	var err, gErr error
	if s.mDNSServer != nil {
		s.mDNSServer.Shutdown()
	}

	if s.bleAD != nil {
		if err = s.bleAD.Stop(); err != nil && !strings.Contains(err.Error(), "advertisement is not started") {
			gErr = fmt.Errorf("%w: stop ble advertisement: %w", gErr, err)
		}
	}

	return gErr
}
