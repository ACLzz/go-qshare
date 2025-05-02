package goqshare_test

import (
	"net"
	"os"
	"testing"
	"time"

	qshare "github.com/ACLzz/go-qshare"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/hashicorp/mdns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tinygo.org/x/bluetooth"
)

func TestServer_mDNS(t *testing.T) {
	t.Run("with_custom_values", func(t *testing.T) {
		ipv4 := net.IPv4(127, 0, 0, 15)
		ipv6 := net.IP{1, 2, 3, 4, 5, 6, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		hostname := "test_hostname"
		port := 6666

		server, err := qshare.NewServer().
			WithHostname(hostname).
			WithIPv4(ipv4).
			WithIPv6(ipv6).
			WithPort(uint(port)).
			Build()
		require.NoError(t, err)
		require.NotNil(t, server)

		entriesCh := make(chan *mdns.ServiceEntry, 1)
		mdns.Lookup(qshare.QS_TYPE, entriesCh)
		select {
		case <-entriesCh:
			t.Fail()
		case <-time.After(3 * time.Second):
		}

		require.NoError(t, server.Listen())
		t.Cleanup(func() {
			require.NoError(t, server.Stop())
		})

		entriesCh = make(chan *mdns.ServiceEntry, 1)
		mdns.Lookup(qshare.QS_TYPE, entriesCh)
		select {
		case entry := <-entriesCh:
			assert.Equal(t, ipv4[12:], entry.AddrV4)
			assert.Equal(t, ipv6, entry.AddrV6)
			assert.Equal(t, hostname+".", entry.Host)
			assert.Equal(t, port, entry.Port)
		case <-time.After(3 * time.Second):
			t.Fail()
		}
	})

	t.Run("with_default_values", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		t.Cleanup(patches.Reset)
		patches.ApplyFunc(os.Hostname, func() (string, error) {
			return "test", nil
		})

		server, err := qshare.NewServer().Build()
		require.NoError(t, err)
		require.NotNil(t, server)

		entriesCh := make(chan *mdns.ServiceEntry, 1)
		mdns.Lookup(qshare.QS_TYPE, entriesCh)
		select {
		case <-entriesCh:
			t.Fail()
		case <-time.After(3 * time.Second):
		}

		require.NoError(t, server.Listen())
		t.Cleanup(func() {
			require.NoError(t, server.Stop())
		})

		entriesCh = make(chan *mdns.ServiceEntry, 1)
		mdns.Lookup(qshare.QS_TYPE, entriesCh)
		select {
		case entry := <-entriesCh:
			assert.Equal(t, "test.", entry.Host)
			assert.Greater(t, entry.Port, 1024)
		case <-time.After(3 * time.Second):
			t.Fail()
		}
	})
}

func TestServer_bleAdvertisements(t *testing.T) {
	t.Run("adapter_available", func(t *testing.T) {
		isAdapterStarted := false

		patches := gomonkey.NewPatches()
		t.Cleanup(patches.Reset)
		patches.ApplyMethodFunc(bluetooth.DefaultAdapter, "Enable", func() error {
			patches.Reset()
			bluetooth.DefaultAdapter.Enable()
			isAdapterStarted = true
			return nil
		})

		server, err := qshare.NewServer().
			Build()
		require.NoError(t, err)
		require.NotNil(t, server)

		require.NoError(t, server.Listen())
		t.Cleanup(func() {
			require.NoError(t, server.Stop())
		})

		assert.True(t, isAdapterStarted)
	})
}
