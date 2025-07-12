package qserver

import (
	"os"
	"strings"
	"testing"

	"github.com/ACLzz/qshare"
	"github.com/ACLzz/qshare/internal/log"
	"github.com/ACLzz/qshare/internal/rand"
	"github.com/ACLzz/qshare/internal/tests/helper"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tinygo.org/x/bluetooth"
)

func TestBuilder(t *testing.T) {
	rng := rand.NewStatic(1746361391)
	logger := log.NewLogger()

	type (
		expectConf struct {
			name string
			txt  string
		}

		expect struct {
			conf expectConf
			err  error
		}
	)
	tests := map[string]struct {
		prepare func(t *testing.T, b *serverBuilder)
		expect  expect
	}{
		"success/default_values": {
			expect: expect{
				conf: expectConf{
					name: "I2FhYWH8n14AAA",
					txt:  "AC8vLy8vLy8vLy8vLy8vLy8NdGVzdF9ob3N0bmFtZQ",
				},
			},
			prepare: func(t *testing.T, b *serverBuilder) {
				hostname := "test_hostname"
				patches := gomonkey.NewPatches()
				patches.ApplyFunc(os.Hostname, func() (string, error) {
					return hostname, nil
				})
				t.Cleanup(patches.Reset)
			},
		},
		"success/custom_values": {
			expect: expect{
				conf: expectConf{
					name: "I3NvbWX8n14AAA",
					txt:  "BC8vLy8vLy8vLy8vLy8vLy8EdGVzdA",
				},
			},
			prepare: func(t *testing.T, b *serverBuilder) {
				t.Helper()
				*b = *b.
					WithAdapter(bluetooth.DefaultAdapter).
					WithDeviceType(qshare.TabletDevice).
					WithEndpoint("some").
					WithHostname("test").
					WithLogger(logger).
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
		"error/invalid_logger": {
			expect: expect{
				err: ErrInvalidLogger,
			},
			prepare: func(t *testing.T, b *serverBuilder) {
				t.Helper()
				*b = *b.
					WithLogger(nil)
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// skip tests that cannot be tested in ci
			if helper.IsCI() && strings.Contains(name, "invalid_adapter") {
				t.Skip()
			}

			builder := NewBuilder().
				WithRandom(rng)
			if tt.prepare != nil {
				tt.prepare(t, builder)
			}

			server, err := builder.Build(nil, nil)
			if tt.expect.err != nil {
				assert.ErrorIs(t, err, tt.expect.err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, server)

			assert.Equal(t, tt.expect.conf.name, server.conf.name)
			assert.Equal(t, tt.expect.conf.txt, server.conf.txt)
			assert.Greater(t, server.conf.port, 1024)
		})
	}
}
