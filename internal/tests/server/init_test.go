package server_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/ACLzz/qshare"
	"github.com/ACLzz/qshare/internal/adapter"
	"github.com/ACLzz/qshare/internal/log"
	logMock "github.com/ACLzz/qshare/internal/mock"
	"github.com/ACLzz/qshare/internal/tests/helper"

	pbSecuregcm "github.com/ACLzz/qshare/internal/protobuf/gen/securegcm"
	"github.com/ACLzz/qshare/internal/rand"
	"github.com/ACLzz/qshare/qserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

func assertBadMessage(t *testing.T, resp []byte) {
	var (
		ukeyMessage pbSecuregcm.Ukey2Message
		ukeyAlert   pbSecuregcm.Ukey2Alert
	)
	require.NoError(t, proto.Unmarshal(resp, &ukeyMessage))
	assert.Equal(t, pbSecuregcm.Ukey2Message_ALERT, ukeyMessage.GetMessageType())
	require.NoError(t, proto.Unmarshal(ukeyMessage.GetMessageData(), &ukeyAlert))
	assert.Equal(t, pbSecuregcm.Ukey2Alert_BAD_MESSAGE, ukeyAlert.GetType())
	assert.Equal(t, adapter.ErrInvalidMessage.Error(), string(ukeyAlert.GetErrorMessage()))
}

func TestServer_ConnectionInit(t *testing.T) {
	var (
		rng      = rand.NewStatic(1)
		port     = helper.RandomPort()
		endpoint = "test"
		hostname = "text_test"
	)
	tests := map[string]struct {
		preapre func(t *testing.T, log *logMock.MockLogger)
		assert  func(t *testing.T, ctx context.Context, adp *adapter.Adapter)
	}{
		"success": {
			preapre: func(t *testing.T, log *logMock.MockLogger) {
				t.Helper()
				log.EXPECT().Info(gomock.Any()).AnyTimes()
				log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
			},
			assert: func(t *testing.T, ctx context.Context, adp *adapter.Adapter) {
				read := adp.Reader(ctx)

				require.NoError(t, adp.SendConnRequest(endpoint, hostname, qshare.LaptopDevice))
				require.NoError(t, adp.SendClientInitWithClientFinished())
				resp, err := read()
				require.NoError(t, err)
				require.NoError(t, adp.ValidateServerInit(resp))

				require.NoError(t, adp.SendConnResponse(true))
				resp, err = read()
				require.NoError(t, err)
				connResp, err := adp.UnmarshalConnResponse(resp)
				require.NoError(t, err)
				require.True(t, connResp.IsConnAccepted)
				require.NoError(t, adp.EnableEncryption())

				require.NoError(t, adp.SendPairedKeyEncryption())
				resp, err = read()
				require.NoError(t, err)
				require.NoError(t, adp.ValidatePairedKeyEncryption(resp))

				require.NoError(t, adp.SendPairedKeyResult())
				resp, err = read()
				require.NoError(t, err)
				require.NoError(t, adp.ValidatePairedKeyResult(resp))
			},
		},
		"error/no_connection_request": {
			preapre: func(t *testing.T, log *logMock.MockLogger) {
				t.Helper()
				log.EXPECT().Info(gomock.Any()).AnyTimes()
				log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
				log.EXPECT().Warn("got invalid message", gomock.Cond(func(x []any) bool {
					if len(x) != 2 {
						return false
					}
					return x[0] == "expectedMessage"
				}))
			},
			assert: func(t *testing.T, ctx context.Context, adp *adapter.Adapter) {
				read := adp.Reader(ctx)

				require.NoError(t, adp.SendConnResponse(true))

				resp, err := read()
				require.NoError(t, err)
				assertBadMessage(t, resp)
			},
		},
		"error/no_client_init_client_finished": {
			preapre: func(t *testing.T, log *logMock.MockLogger) {
				t.Helper()
				log.EXPECT().Info(gomock.Any()).AnyTimes()
				log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
				log.EXPECT().Warn("got invalid message", gomock.Cond(func(x []any) bool {
					if len(x) != 2 {
						return false
					}
					return x[0] == "expectedMessage"
				}))
			},
			assert: func(t *testing.T, ctx context.Context, adp *adapter.Adapter) {
				read := adp.Reader(ctx)

				require.NoError(t, adp.SendConnRequest(endpoint, hostname, qshare.LaptopDevice))
				require.NoError(t, adp.SendConnResponse(true))
				resp, err := read()
				require.NoError(t, err)
				assertBadMessage(t, resp)
			},
		},
		"error/no_paired_key_encryption": {
			preapre: func(t *testing.T, log *logMock.MockLogger) {
				t.Helper()
				log.EXPECT().Info(gomock.Any()).AnyTimes()
				log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
				log.EXPECT().Warn("got invalid message", gomock.Cond(func(x []any) bool {
					if len(x) != 2 {
						return false
					}
					return x[0] == "expectedMessage"
				}))
			},
			assert: func(t *testing.T, ctx context.Context, adp *adapter.Adapter) {
				read := adp.Reader(ctx)

				require.NoError(t, adp.SendConnRequest(endpoint, hostname, qshare.LaptopDevice))
				require.NoError(t, adp.SendClientInitWithClientFinished())
				resp, err := read()
				require.NoError(t, err)
				require.NoError(t, adp.ValidateServerInit(resp))

				require.NoError(t, adp.SendConnResponse(true))
				resp, err = read()
				require.NoError(t, err)
				connResp, err := adp.UnmarshalConnResponse(resp)
				require.NoError(t, err)
				require.True(t, connResp.IsConnAccepted)
				require.NoError(t, adp.EnableEncryption())

				require.NoError(t, adp.SendPairedKeyResult())
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl, ctx := gomock.WithContext(t.Context(), t)
			t.Cleanup(ctrl.Finish)

			logger := logMock.NewMockLogger(ctrl)
			if tt.preapre != nil {
				tt.preapre(t, logger)
			}

			server, err := qserver.NewBuilder().
				WithRandom(rng).
				WithPort(port).
				WithHostname(hostname).
				WithEndpoint(endpoint).
				WithLogger(logger).
				Build(helper.NewStubAuthCallback(true), nil, nil)
			require.NoError(t, err)

			require.NoError(t, server.Listen())
			t.Cleanup(func() {
				require.NoError(t, server.Stop())
				helper.WaitForMDNSTTL(t)
			})
			addr := fmt.Sprintf("%s:%d", "127.0.0.1", port)
			netAddr, err := net.ResolveTCPAddr("tcp", addr)
			require.NoError(t, err)
			c, err := net.DialTCP("tcp", nil, netAddr)
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, c.Close())
			})

			adp := adapter.New(c, log.NewLogger(), false, nil, nil, rng)
			t.Cleanup(adp.Disconnect)
			tt.assert(t, ctx, &adp)
		})
	}
}

func TestServer_SetupTransfer(t *testing.T) {
	var (
		rng      = rand.NewStatic(1)
		port     = helper.RandomPort()
		endpoint = "test"
		hostname = "text_test"
	)
	tests := map[string]struct {
		preapre  func(t *testing.T, log *logMock.MockLogger)
		assert   func(t *testing.T, read func() ([]byte, error), adp *adapter.Adapter)
		callback func(t *testing.T, text *qshare.TextMeta, files []qshare.FileMeta, pin uint16) bool
	}{
		"success/text/text": {
			preapre: func(t *testing.T, log *logMock.MockLogger) {
				t.Helper()
				log.EXPECT().Info(gomock.Any()).AnyTimes()
				log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
			},
			assert: func(t *testing.T, read func() ([]byte, error), adp *adapter.Adapter) {
				text := "text"
				require.NoError(t, adp.SendIntroduction(adapter.IntroductionFrame{
					Text: &adapter.TextMeta{
						TextMeta: qshare.TextMeta{
							Type:  qshare.TextText,
							Title: "text...",
							Size:  int64(len(text)),
						},
						ID: 0,
					},
				}))
				require.NoError(t, adp.SendTransferRequest())

				resp, err := read()
				require.NoError(t, err)
				isAccepted, err := adp.UnmarshalTransferResponse(resp)
				require.NoError(t, err)
				assert.True(t, isAccepted)
			},
			callback: func(t *testing.T, text *qshare.TextMeta, files []qshare.FileMeta, pin uint16) bool {
				return true
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl, ctx := gomock.WithContext(t.Context(), t)
			t.Cleanup(ctrl.Finish)

			logger := logMock.NewMockLogger(ctrl)
			if tt.preapre != nil {
				tt.preapre(t, logger)
			}

			callback := func(text *qshare.TextMeta, files []qshare.FileMeta, pin uint16) bool {
				if tt.callback != nil {
					return tt.callback(t, text, files, pin)
				}
				return true
			}
			server, err := qserver.NewBuilder().
				WithRandom(rng).
				WithPort(port).
				WithHostname(hostname).
				WithEndpoint(endpoint).
				WithLogger(logger).
				Build(callback, nil, nil)
			require.NoError(t, err)

			require.NoError(t, server.Listen())
			t.Cleanup(func() {
				require.NoError(t, server.Stop())
				helper.WaitForMDNSTTL(t)
			})
			addr := fmt.Sprintf("%s:%d", "127.0.0.1", port)
			netAddr, err := net.ResolveTCPAddr("tcp", addr)
			require.NoError(t, err)
			c, err := net.DialTCP("tcp", nil, netAddr)
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, c.Close())
			})
			adp := adapter.New(c, log.NewLogger(), false, nil, nil, rng)
			read := adp.Reader(ctx)

			require.NoError(t, adp.SendConnRequest(endpoint, hostname, qshare.LaptopDevice))
			require.NoError(t, adp.SendClientInitWithClientFinished())
			resp, err := read()
			require.NoError(t, err)
			require.NoError(t, adp.ValidateServerInit(resp))

			require.NoError(t, adp.SendConnResponse(true))
			resp, err = read()
			require.NoError(t, err)
			connResp, err := adp.UnmarshalConnResponse(resp)
			require.NoError(t, err)
			require.True(t, connResp.IsConnAccepted)
			require.NoError(t, adp.EnableEncryption())

			require.NoError(t, adp.SendPairedKeyEncryption())
			resp, err = read()
			require.NoError(t, err)
			require.NoError(t, adp.ValidatePairedKeyEncryption(resp))

			require.NoError(t, adp.SendPairedKeyResult())
			resp, err = read()
			require.NoError(t, err)
			require.NoError(t, adp.ValidatePairedKeyResult(resp))

			t.Cleanup(adp.Disconnect)
			tt.assert(t, read, &adp)
		})
	}
}
