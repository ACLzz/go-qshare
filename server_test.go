package goqshare_test

import (
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

func TestServer_StartStop(t *testing.T) {
	t.Run("start_stop_with_listen", func(t *testing.T) {
		server, err := qshare.NewServer().Build()
		require.NoError(t, err)
		require.NotNil(t, server)

		require.NoError(t, server.Listen())
		require.NoError(t, server.Stop())
	})

	t.Run("start_stop_without_listen", func(t *testing.T) {
		server, err := qshare.NewServer().Build()
		require.NoError(t, err)
		require.NotNil(t, server)

		require.NoError(t, server.Stop())
	})
}

func TestServer_mDNS(t *testing.T) {
	t.Run("with_custom_values", func(t *testing.T) {
		hostname := "test_hostname"
		port := 6666

		server, err := qshare.NewServer().
			WithHostname(hostname).
			WithPort(port).
			Build()
		require.NoError(t, err)
		require.NotNil(t, server)

		entriesCh := make(chan *mdns.ServiceEntry, 1)
		mdns.Lookup(qshare.ServiceType, entriesCh)
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
		mdns.Lookup(qshare.ServiceType, entriesCh)
		select {
		case entry := <-entriesCh:
			assert.Equal(t, hostname+".", entry.Host)
			assert.Equal(t, port, entry.Port)
		case <-time.After(3 * time.Second):
			t.Fail()
		}
	})

	t.Run("with_default_values", func(t *testing.T) {
		machineHostname, err := os.Hostname()
		require.NoError(t, err)

		server, err := qshare.NewServer().Build()
		require.NoError(t, err)
		require.NotNil(t, server)

		entriesCh := make(chan *mdns.ServiceEntry, 1)
		mdns.Lookup(qshare.ServiceType, entriesCh)
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
		mdns.Lookup(qshare.ServiceType, entriesCh)
		select {
		case entry := <-entriesCh:
			assert.Equal(t, machineHostname+".", entry.Host)
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
