package server

import (
	"encoding/base64"
	"fmt"
	"math/rand/v2"
	"net"
	"os"

	qshare "github.com/ACLzz/go-qshare"
	internalLog "github.com/ACLzz/go-qshare/internal/log"
	"github.com/ACLzz/go-qshare/log"
	"github.com/ACLzz/go-qshare/server/comm"

	"tinygo.org/x/bluetooth"
)

var ble_service_data_base = [14]byte{252, 18, 142, 1, 66, 0, 0, 0, 0, 0, 0, 0, 0, 0}
var alphaNumRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type serverBuilder struct {
	rand *rand.Rand

	hostname string
	port     int
	endpoint []byte
	adapter  *bluetooth.Adapter
	device   qshare.DeviceType
	logger   log.Logger

	isHostnameSet         bool
	isPortSet             bool
	isEndpointSet         bool
	isBluetoothAdapterSet bool
	isDeviceTypeSet       bool
	isLoggerSet           bool
}

func NewServerBuilder(clk qshare.Clock) *serverBuilder {
	return &serverBuilder{
		rand: rand.New(clk),
	}
}

func (b *serverBuilder) WithHostname(hostname string) *serverBuilder {
	b.hostname = hostname
	b.isHostnameSet = true
	return b
}

func (b *serverBuilder) WithPort(port int) *serverBuilder {
	b.port = port
	b.isPortSet = true
	return b
}

func (b *serverBuilder) WithEndpoint(endpoint string) *serverBuilder {
	b.endpoint = []byte(endpoint)
	b.isEndpointSet = true
	return b
}

func (b *serverBuilder) WithAdapter(adapter *bluetooth.Adapter) *serverBuilder {
	b.adapter = adapter
	b.isBluetoothAdapterSet = true
	return b
}

func (b *serverBuilder) WithDeviceType(device qshare.DeviceType) *serverBuilder {
	b.device = device
	b.isDeviceTypeSet = true
	return b
}

func (b *serverBuilder) WithLogger(logger log.Logger) *serverBuilder {
	b.logger = logger
	b.isLoggerSet = true
	return b
}

func (b *serverBuilder) Build() (*Server, error) {
	if err := b.propagateDefaultValues(); err != nil {
		return nil, fmt.Errorf("propagate default values: %w", err)
	}

	bleAD, err := b.newBLEAdvertisement()
	if err != nil {
		return nil, fmt.Errorf("create ble advertisement: %w", err)
	}

	cs, err := comm.NewServer(b.port, b.logger)
	if err != nil {
		return nil, fmt.Errorf("create communication server: %w", err)
	}

	return &Server{
		bleAD:      bleAD,
		commServer: cs,
		conf: serverConfig{
			name: newName(b.endpoint),
			port: b.port,
			txt:  newTXT(b.hostname, b.device, b.rand),
		},
	}, nil
}

func (b *serverBuilder) newBLEAdvertisement() (*bluetooth.Advertisement, error) {
	ad := b.adapter.DefaultAdvertisement()
	bleUUID := bluetooth.New16BitUUID(0xfe2c)

	serviceData := make([]byte, len(ble_service_data_base)+10)
	copy(serviceData[0:], ble_service_data_base[:])
	for i := len(ble_service_data_base); i < len(serviceData); i++ {
		serviceData[i] = byte(b.rand.IntN(256)) // random byte
	}

	err := ad.Configure(bluetooth.AdvertisementOptions{
		ServiceData:       []bluetooth.ServiceDataElement{{UUID: bleUUID, Data: serviceData}},
		AdvertisementType: bluetooth.AdvertisingTypeScanInd,
	})
	if err != nil {
		return nil, fmt.Errorf("configure default ble advertisements: %w", err)
	}

	return ad, nil
}

type propagateValueFn func() error

func (b *serverBuilder) propagateDefaultValues() error {
	funcs := []propagateValueFn{
		b.propHostname,
		b.propPort,
		b.propBluAdapter,
		b.propEndpoint,
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

func (b *serverBuilder) propHostname() error {
	if b.isHostnameSet {
		return nil
	}

	var err error
	b.hostname, err = os.Hostname()
	if err != nil {
		return fmt.Errorf("get os hostname: %w", err)
	}

	return nil
}

func (b *serverBuilder) propPort() error {
	if b.isPortSet {
		if b.port <= 1024 {
			return ErrInvalidPort
		}

		return nil
	}

	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return fmt.Errorf("open tcp socket: %w", err)
	}
	defer l.Close() // TODO: handle error

	b.port = l.Addr().(*net.TCPAddr).Port
	return nil
}

func (b *serverBuilder) propBluAdapter() error {
	if b.isBluetoothAdapterSet {
		if b.adapter == nil {
			return ErrInvalidAdapter
		}
	} else {
		b.adapter = bluetooth.DefaultAdapter
	}

	if err := b.adapter.Enable(); err != nil {
		return fmt.Errorf("%w: enable bluetooth adapter: %w", ErrInvalidAdapter, err)
	}

	return nil
}

func (b *serverBuilder) propEndpoint() error {
	if b.isEndpointSet {
		if len(b.endpoint) != 4 {
			return ErrInvalidEndpoint
		}
		return nil
	}

	endpoint := make([]byte, 4)
	for i := range 4 {
		endpoint[i] = byte(alphaNumRunes[b.rand.IntN(len(alphaNumRunes))])
	}
	b.endpoint = endpoint

	return nil
}

func (b *serverBuilder) propDeviceType() error {
	if b.isDeviceTypeSet {
		if !b.device.IsValid() {
			return ErrInvalidDeviceType
		}
		return nil
	}

	b.device = qshare.UnknownDevice
	return nil
}

func (b *serverBuilder) propLogger() error {
	if b.isLoggerSet {
		return nil
	}

	b.logger = internalLog.NewLogger()
	return nil
}

func newName(endpoint []byte) string {
	name := make([]byte, 10)

	name[0] = 0x23                           // PCP
	copy(name[1:], endpoint)                 // endpoint
	copy(name[5:], []byte{0xFC, 0x9F, 0x5E}) // service ID

	return base64.RawURLEncoding.EncodeToString(name)
}

func newTXT(hostname string, device qshare.DeviceType, r *rand.Rand) string {
	n := make([]byte, 18+len(hostname))
	n[0] = byte(device << 1) // set device type and mark service "discoverable"

	for i := range 16 {
		n[1+i] = byte(r.IntN(256)) // random byte
	}

	// add hostname length and hostname
	n[17] = byte(len(hostname))
	copy(n[18:], []byte(hostname))

	return base64.RawURLEncoding.EncodeToString(n)
}
