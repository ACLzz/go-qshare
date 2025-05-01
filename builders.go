package goqshare

import (
	"encoding/base64"
	"fmt"
	"math/rand/v2"
	"net"
	"os"
	"strconv"

	"github.com/hashicorp/mdns"
	"tinygo.org/x/bluetooth"
)

const TXT_RECORD = 0b00011110
const QS_TYPE = "_FC9F5ED42C8A._tcp"
const ASCII_MAX = 126
const ASCII_MIN = 32

var alphaNumRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type serverBuilder struct {
	hostname string
	port     uint
	ipv4     net.IP
	ipv6     net.IP
	adapter  *bluetooth.Adapter

	isHostnameSet         bool
	isPortSet             bool
	isIPv4Set             bool
	isIPv6Set             bool
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

func (b *serverBuilder) WithIPv6(ipv6 net.IP) *serverBuilder {
	b.ipv6 = ipv6
	b.isIPv6Set = true
	return b
}

func (b *serverBuilder) WithPort(port uint) *serverBuilder {
	b.port = port
	b.isPortSet = true
	return b
}

func (b *serverBuilder) WithAdapter(adapter *bluetooth.Adapter) *serverBuilder {
	b.adapter = adapter
	b.isBluetoothAdapterSet = true
	return b
}

func (b *serverBuilder) Build() (*Server, error) {
	hostname, err := b.getHostname()
	if err != nil {
		return nil, fmt.Errorf("get builder hostname: %w", err)
	}

	port, err := b.getPort()
	if err != nil {
		return nil, fmt.Errorf("get port: %w", err)
	}

	mDNSService, err := newMDNSService(hostname, port, b.getIPv4(), b.getIPv6())
	if err != nil {
		return nil, fmt.Errorf("new mDNS service: %w", err)
	}

	ad, err := newBLEAdvertisement(b.getAdapter(), hostname)
	if err != nil {
		return nil, fmt.Errorf("create ble advertisement: %w", err)
	}

	server := Server{
		mDNSService: mDNSService,
		ad:          ad,
	}

	return &server, nil
}

func newMDNSService(hostname string, port int, ipv4, ipv6 net.IP) (*mdns.MDNSService, error) {
	txt := []string{strconv.Itoa(TXT_RECORD), randomString(16), hostname}
	ips := []net.IP{ipv4, ipv6}
	mDNSService, err := mdns.NewMDNSService(newName(), QS_TYPE, "", hostname, port, ips, txt)
	if err != nil {
		return nil, fmt.Errorf("cannot set up mDNS service: %w", err)
	}

	return mDNSService, nil
}

func newBLEAdvertisement(adapter *bluetooth.Adapter, hostname string) (*bluetooth.Advertisement, error) {
	if err := adapter.Enable(); err != nil {
		return nil, fmt.Errorf("enable bluetooth adapter: %w", err)
	}

	ad := adapter.DefaultAdvertisement()

	bleUUID := bluetooth.UUID{0, 0, 0xfe, 0x2c}
	err := ad.Configure(bluetooth.AdvertisementOptions{
		ServiceUUIDs: []bluetooth.UUID{bleUUID},
		ServiceData:  []bluetooth.ServiceDataElement{{UUID: bleUUID, Data: []byte(randomString(10))}},
		LocalName:    fmt.Sprintf("%s QuickShare", hostname),
	})
	if err != nil {
		return nil, fmt.Errorf("configure default ble advertisements: %w", err)
	}

	return ad, nil
}

func (b *serverBuilder) getHostname() (string, error) {
	if b.isHostnameSet {
		return b.hostname + ".", nil
	}

	hostName, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("get os hostname: %w", err)
	}
	hostName += "."

	return hostName, nil
}

func (b *serverBuilder) getPort() (int, error) {
	if b.isPortSet {
		return int(b.port), nil
	}

	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, fmt.Errorf("open tcp socket: %w", err)
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}

func (b *serverBuilder) getIPv4() net.IP {
	if b.isIPv4Set {
		return b.ipv4
	}

	return net.IPv4zero
}

func (b *serverBuilder) getIPv6() net.IP {
	if b.isIPv6Set {
		return b.ipv6
	}

	return net.IPv6zero
}

func (b *serverBuilder) getAdapter() *bluetooth.Adapter {
	if b.isBluetoothAdapterSet {
		return b.adapter
	}

	return bluetooth.DefaultAdapter
}

func randomString(length uint) string {
	s := make([]rune, length)
	for i := range length {
		s[i] = rune(rand.IntN(ASCII_MAX-ASCII_MIN) + ASCII_MIN)
	}
	return string(s)
}

func newName() string {
	name := make([]byte, 10)
	for i := range 4 {
		name[1+i] = byte(alphaNumRunes[rand.IntN(len(alphaNumRunes))])
	}

	name[0] = 0x23
	copy(name[5:], []byte{0xFC, 0x9F, 0x5E})
	return base64.URLEncoding.EncodeToString([]byte(name))
}
