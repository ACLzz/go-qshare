package client

import (
	"fmt"
	"math/rand/v2"
	"sync"

	qshare "github.com/ACLzz/go-qshare"
	internalLog "github.com/ACLzz/go-qshare/internal/log"
	"github.com/ACLzz/go-qshare/log"
	"tinygo.org/x/bluetooth"
)

type clientBuilder struct {
	rand    *rand.Rand
	adapter *bluetooth.Adapter
	device  qshare.DeviceType
	logger  log.Logger
	// TODO: add hostname

	isLoggerSet     bool
	isDeviceTypeSet bool
}

func NewBuilder(clk qshare.Clock) *clientBuilder {
	return &clientBuilder{
		rand: rand.New(clk),
	}
}

func (b *clientBuilder) WithDeviceType(device qshare.DeviceType) *clientBuilder {
	b.device = device
	b.isDeviceTypeSet = true
	return b
}

func (b *clientBuilder) WithLogger(logger log.Logger) *clientBuilder {
	b.logger = logger
	b.isLoggerSet = true
	return b
}

func (b *clientBuilder) Build(adEndpointID string) (*Client, error) {
	if err := b.propagateDefaultValues(); err != nil {
		return nil, fmt.Errorf("propagate default values: %w", err)
	}

	return &Client{
		wg:         &sync.WaitGroup{},
		log:        b.logger,
		endpointID: adEndpointID,
	}, nil
}

type propagateValueFn func() error

func (b *clientBuilder) propagateDefaultValues() error {
	funcs := []propagateValueFn{
		b.propDeviceType,
		b.propLogger,
	}

	var err error
	for i := range funcs {
		err = funcs[i]()
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *clientBuilder) propDeviceType() error {
	if b.isDeviceTypeSet {
		if !b.device.IsValid() {
			return ErrInvalidDeviceType
		}
		return nil
	}

	b.device = qshare.UnknownDevice
	return nil
}

func (b *clientBuilder) propLogger() error {
	if b.isLoggerSet {
		return nil
	}

	b.logger = internalLog.NewLogger()
	return nil
}
