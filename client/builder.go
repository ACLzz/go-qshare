package client

import (
	"fmt"
	"sync"

	"github.com/ACLzz/qshare"
	internalLog "github.com/ACLzz/qshare/internal/log"
	"github.com/ACLzz/qshare/internal/rand"
	"tinygo.org/x/bluetooth"
)

type clientBuilder struct {
	rand    rand.Random
	adapter *bluetooth.Adapter
	device  qshare.DeviceType
	logger  qshare.Logger
	// TODO: add hostname

	isLoggerSet     bool
	isRandomSet     bool
	isDeviceTypeSet bool
}

func NewBuilder() *clientBuilder {
	return &clientBuilder{}
}

func (b *clientBuilder) WithDeviceType(device qshare.DeviceType) *clientBuilder {
	b.device = device
	b.isDeviceTypeSet = true
	return b
}

func (b *clientBuilder) WithLogger(logger qshare.Logger) *clientBuilder {
	b.logger = logger
	b.isLoggerSet = true
	return b
}

func (b *clientBuilder) WithRandom(r rand.Random) *clientBuilder {
	b.rand = r
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
		rand:       b.rand,
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
