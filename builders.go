package goqshare

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/mdns"
	"tinygo.org/x/bluetooth"
)

const ServiceType = "_FC9F5ED42C8A._tcp"
const txt_record = 0b00011110
const ascii_max = 126
const ascii_min = 32

var alphaNumRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

var (
	ErrInvalidPort     = errors.New("invalid port value")
	ErrInvalidAdapter  = errors.New("invalid adapter")
	ErrInvalidEndpoint = errors.New("invalid endpoint")
)

type serverBuilder struct {
	hostname string
	port     int
	ipv4     net.IP // FIXME: currently not used
	endpoint []byte
	adapter  *bluetooth.Adapter

	isHostnameSet         bool
	isPortSet             bool
	isIPv4Set             bool
	isEndpointSet         bool
	isBluetoothAdapterSet bool
}

func NewServer() *serverBuilder {
	return &serverBuilder{}
}

func (b *serverBuilder) WithHostname(hostname string) *serverBuilder {
	b.hostname = hostname
	b.isHostnameSet = true
	return b
}

func (b *serverBuilder) WithIPv4(ipv4 net.IP) *serverBuilder {
	b.ipv4 = ipv4
	b.isIPv4Set = true
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

func (b *serverBuilder) Build() (*Server, error) {
	if err := b.propagateDefaultValues(); err != nil {
		return nil, fmt.Errorf("propagate default values: %w", err)
	}

	mDNSService, err := newMDNSService(b.hostname, b.port)
	if err != nil {
		return nil, fmt.Errorf("new mDNS service: %w", err)
	}

	ad, err := newBLEAdvertisement(b.adapter, b.hostname)
	if err != nil {
		return nil, fmt.Errorf("create ble advertisement: %w", err)
	}

	server := Server{
		mDNSService: mDNSService,
		ad:          ad,
	}

	return &server, nil
}

func newMDNSService(hostname string, port int) (*mdns.MDNSService, error) {
	txt := []string{strconv.Itoa(txt_record), randomString(16), hostname}

	origHostName, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("get original hostname of machine: %w", err)
	}

	ips, err := net.LookupIP(origHostName)
	if err != nil {
		return nil, fmt.Errorf("determine host IP addresses for host: %w", err)
	}

	mDNSService, err := mdns.NewMDNSService(newName(), ServiceType, "", hostname, port, ips, txt)
	if err != nil {
		return nil, fmt.Errorf("set up mDNS service: %w", err)
	}

	return mDNSService, nil
}

func newBLEAdvertisement(adapter *bluetooth.Adapter, hostname string) (*bluetooth.Advertisement, error) {
	if err := adapter.Enable(); err != nil {
		return nil, fmt.Errorf("enable bluetooth adapter: %w", err)
	}

	ad := adapter.DefaultAdvertisement()

	bleUUID := bluetooth.New16BitUUID(0xfe2c)
	err := ad.Configure(bluetooth.AdvertisementOptions{
		LocalName:   fmt.Sprintf("%s QuickShare", hostname),
		ServiceData: []bluetooth.ServiceDataElement{{UUID: bleUUID, Data: []byte(randomString(10))}},
		Interval:    bluetooth.NewDuration(5 * time.Second),
		ManufacturerData: []bluetooth.ManufacturerDataElement{
			{CompanyID: 0xffff, Data: []byte{0x01, 0x02}},
		},
		AdvertisementType: bluetooth.AdvertisingTypeInd,
	})
	if err != nil {
		return nil, fmt.Errorf("configure default ble advertisements: %w", err)
	}

	return ad, nil
}

type propagateDefaultValueFn func() error

func (b *serverBuilder) propagateDefaultValues() error {
	funcs := []propagateDefaultValueFn{
		b.propHostname,
		b.propPort,
		b.propIPv4,
		b.propAdapter,
		b.propEndpoint,
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
		b.hostname += "."
		return nil
	}

	hostName, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("get os hostname: %w", err)
	}
	hostName += "."

	b.hostname = hostName
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
	defer l.Close()

	b.port = l.Addr().(*net.TCPAddr).Port
	return nil
}

func (b *serverBuilder) propIPv4() error {
	if b.isIPv4Set {
		return nil
	}

	b.ipv4 = net.IPv4zero
	return nil
}

func (b *serverBuilder) propAdapter() error {
	if b.isBluetoothAdapterSet {
		if b.adapter == nil {
			return ErrInvalidAdapter
		}
		return nil
	}

	b.adapter = bluetooth.DefaultAdapter
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
		endpoint[i] = byte(alphaNumRunes[rand.IntN(len(alphaNumRunes))])
	}
	b.endpoint = endpoint

	return nil
}

func randomString(length uint) string {
	s := make([]rune, length)
	for i := range length {
		s[i] = rune(rand.IntN(ascii_max-ascii_min) + ascii_min)
	}
	return string(s)
}

func newName() string {
	name := make([]byte, 10)
	name[0] = 0x23
	for i := range 4 {
		name[1+i] = byte(alphaNumRunes[rand.IntN(len(alphaNumRunes))])
	}

	copy(name[5:], []byte{0xFC, 0x9F, 0x5E})
	return base64.URLEncoding.EncodeToString(name)
}
