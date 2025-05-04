package goqshare

import (
	"os"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tinygo.org/x/bluetooth"
)

func TestBuilder(t *testing.T) {
	clk := NewStatic(time.Date(2025, 5, 4, 12, 23, 11, 0, time.UTC))
	hostname := "test_hostname"
	patches := gomonkey.NewPatches()
	patches.ApplyFunc(os.Hostname, func() (string, error) {
		return hostname, nil
	})
	t.Cleanup(patches.Reset)

	type (
		expectMDNSConf struct {
			name string
			txt  string
		}

		expect struct {
			mDNSConf expectMDNSConf
			err      error
		}
	)
	tests := map[string]struct {
		prepare func(t *testing.T, b *serverBuilder)
		expect  expect
	}{
		"success/default_values": {
			expect: expect{
				mDNSConf: expectMDNSConf{
					name: "I2FhYWH8n14AAA",
					txt:  "AC8vLy8vLy8vLy8vLy8vLy8NdGVzdF9ob3N0bmFtZQ",
				},
			},
		},
		"success/custom_values": {
			expect: expect{
				mDNSConf: expectMDNSConf{
					name: "I3NvbWX8n14AAA",
					txt:  "BC8vLy8vLy8vLy8vLy8vLy8EdGVzdA",
				},
			},
			prepare: func(t *testing.T, b *serverBuilder) {
				t.Helper()
				*b = *b.
					WithAdapter(bluetooth.DefaultAdapter).
					WithDeviceType(TabletDevice).
					WithEndpoint("some").
					WithHostname("test").
					WithPort(1234)
			},
		},
		"error/invalid_port/below_1024": {
			expect: expect{
				err: ErrInvalidPort,
			},
			prepare: func(t *testing.T, b *serverBuilder) {
				t.Helper()
				*b = *b.
					WithPort(524)
			},
		},
		"error/invalid_port/0": {
			expect: expect{
				err: ErrInvalidPort,
			},
			prepare: func(t *testing.T, b *serverBuilder) {
				t.Helper()
				*b = *b.
					WithPort(0)
			},
		},
		"error/invalid_port/negative": {
			expect: expect{
				err: ErrInvalidPort,
			},
			prepare: func(t *testing.T, b *serverBuilder) {
				t.Helper()
				*b = *b.
					WithPort(-1)
			},
		},
		"error/invalid_endpoint/lt_four": {
			expect: expect{
				err: ErrInvalidEndpoint,
			},
			prepare: func(t *testing.T, b *serverBuilder) {
				t.Helper()
				*b = *b.
					WithEndpoint("abc")
			},
		},
		"error/invalid_endpoint/gt_four": {
			expect: expect{
				err: ErrInvalidEndpoint,
			},
			prepare: func(t *testing.T, b *serverBuilder) {
				t.Helper()
				*b = *b.
					WithEndpoint("abcde")
			},
		},
		"error/invalid_adapter/not_exists": {
			expect: expect{
				err: ErrInvalidAdapter,
			},
			prepare: func(t *testing.T, b *serverBuilder) {
				t.Helper()
				*b = *b.
					WithAdapter(bluetooth.NewAdapter("not_exists"))
			},
		},
		"error/invalid_adapter/nil": {
			expect: expect{
				err: ErrInvalidAdapter,
			},
			prepare: func(t *testing.T, b *serverBuilder) {
				t.Helper()
				*b = *b.
					WithAdapter(nil)
			},
		},
		"error/invalid_device_type": {
			expect: expect{
				err: ErrInvalidDeviceType,
			},
			prepare: func(t *testing.T, b *serverBuilder) {
				t.Helper()
				*b = *b.
					WithDeviceType(4)
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			builder := NewServerBuilder(clk)
			if tt.prepare != nil {
				tt.prepare(t, builder)
			}

			server, err := builder.Build()
			if tt.expect.err != nil {
				assert.ErrorIs(t, err, tt.expect.err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, server)

			assert.Equal(t, tt.expect.mDNSConf.name, server.mDNSConf.name)
			assert.Equal(t, tt.expect.mDNSConf.txt, server.mDNSConf.txt)
			assert.Greater(t, server.mDNSConf.port, 1024)
			assert.NotNil(t, server.bleAD)
		})
	}
}
