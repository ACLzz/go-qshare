package server_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ACLzz/qshare/internal/mdns"
	"github.com/ACLzz/qshare/internal/rand"
	qserver "github.com/ACLzz/qshare/server"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/grandcat/zeroconf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tinygo.org/x/bluetooth"
)

func TestServer_StartStop(t *testing.T) {
	rng := rand.NewStatic(1746361391)

	t.Run("start_stop_with_listen", func(t *testing.T) {
		server, err := qserver.NewBuilder().
			WithRandom(rng).
			Build(nil, nil)
		require.NoError(t, err)
		require.NotNil(t, server)

		require.NoError(t, server.Listen())
		require.NoError(t, server.Stop())
	})

	t.Run("start_stop_without_listen", func(t *testing.T) {
		server, err := qserver.NewBuilder().
			WithRandom(rng).
			Build(nil, nil)
		require.NoError(t, err)
		require.NotNil(t, server)

		require.NoError(t, server.Stop())
	})
}

func TestServer_mDNS(t *testing.T) {
	rng := rand.NewStatic(1746361391)

	t.Run("with_custom_values", func(t *testing.T) {
		machineHostname, err := os.Hostname()
		require.NoError(t, err)

		hostname := "test_hostname"
		port := 6666

		server, err := qserver.NewBuilder().
			WithHostname(hostname).
			WithPort(port).
			WithRandom(rng).
			Build(nil, nil)
		require.NoError(t, err)
		require.NotNil(t, server)

		resolv, err := zeroconf.NewResolver()
		require.NoError(t, err)

		entriesCh := make(chan *zeroconf.ServiceEntry, 2)
		require.NoError(t, resolv.Browse(context.Background(), mdns.MDNsServiceType, "local.", entriesCh))

		select {
		case <-entriesCh:
			t.Fail()
		case <-time.After(3 * time.Second):
		}

		require.NoError(t, server.Listen())
		t.Cleanup(func() {
			require.NoError(t, server.Stop())
		})

		select {
		case entry := <-entriesCh:
			assert.Equal(t, machineHostname+".local.", entry.HostName)
			assert.Equal(t, port, entry.Port)
		case <-time.After(3 * time.Second):
			t.Fail()
		}
	})

	t.Run("with_default_values", func(t *testing.T) {
		machineHostname, err := os.Hostname()
		require.NoError(t, err)

		server, err := qserver.NewBuilder().
			WithRandom(rng).
			Build(nil, nil)
		require.NoError(t, err)
		require.NotNil(t, server)

		resolv, err := zeroconf.NewResolver()
		require.NoError(t, err)

		entriesCh := make(chan *zeroconf.ServiceEntry, 2)
		require.NoError(t, resolv.Browse(context.Background(), mdns.MDNsServiceType, "local.", entriesCh))

		select {
		case <-entriesCh:
			t.Fail()
		case <-time.After(3 * time.Second):
		}

		require.NoError(t, server.Listen())
		t.Cleanup(func() {
			require.NoError(t, server.Stop())
		})

		select {
		case entry := <-entriesCh:
			assert.Equal(t, machineHostname+".local.", entry.HostName)
			assert.Greater(t, entry.Port, 1024)
		case <-time.After(3 * time.Second):
			t.Fail()
		}
	})
}

func TestServer_bleAdvertisements(t *testing.T) {
	rng := rand.NewStatic(1746361391)

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

		server, err := qserver.NewBuilder().
			WithRandom(rng).
			Build(nil, nil)
		require.NoError(t, err)
		require.NotNil(t, server)

		require.NoError(t, server.Listen())
		t.Cleanup(func() {
			require.NoError(t, server.Stop())
		})

		assert.True(t, isAdapterStarted)
	})
}
