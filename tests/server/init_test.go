package server_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/ACLzz/qshare"
	"github.com/ACLzz/qshare/internal/adapter"
	"github.com/ACLzz/qshare/internal/log"
	"github.com/ACLzz/qshare/internal/mock"
	pbSecuregcm "github.com/ACLzz/qshare/internal/protobuf/gen/securegcm"
	"github.com/ACLzz/qshare/internal/rand"
	"github.com/ACLzz/qshare/qserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

func assertBadMessage(t *testing.T, resp []byte) {
	var ukeyMessage pbSecuregcm.Ukey2Message
	require.NoError(t, proto.Unmarshal(resp, &ukeyMessage))
	assert.Equal(t, ukeyMessage.GetMessageType(), pbSecuregcm.Ukey2Message_ALERT)
	assert.Equal(t, ukeyMessage.GetMessageData(), adapter.ErrInvalidMessage.Error())
}

func TestServer_ConnectionInit(t *testing.T) {
	var (
		rng      = rand.NewStatic(1)
		port     = 6666
		endpoint = "test"
		hostname = "text_test"
	)
	tests := map[string]struct {
		preapre func(t *testing.T, log *mock.MockLogger)
		assert  func(t *testing.T, ctx context.Context, adp *adapter.Adapter)
	}{
		"success": {
			preapre: func(t *testing.T, log *mock.MockLogger) {
				t.Helper()
				log.EXPECT().Info(gomock.Any()).AnyTimes()
				log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
			},
			assert: func(t *testing.T, ctx context.Context, adp *adapter.Adapter) {
				t.Helper()
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
				adp.EnableEncryption()

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
			preapre: func(t *testing.T, log *mock.MockLogger) {
				t.Helper()
				log.EXPECT().Info(gomock.Any()).AnyTimes()
				log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
				log.EXPECT().Warn("got invalid message", gomock.Cond(func(x []any) bool {
					if len(x) != 2 {
						return false
					}
					return x[0] == "expectedMessage"
				})).Times(2)
				log.EXPECT().Error("send bad message error", gomock.Any())
			},
			assert: func(t *testing.T, ctx context.Context, adp *adapter.Adapter) {
				t.Helper()
				read := adp.Reader(ctx)

				require.NoError(t, adp.SendConnResponse(true))

				resp, err := read()
				require.NoError(t, err)
				assertBadMessage(t, resp)
			},
		},
		"error/no_connection_response": {
			preapre: func(t *testing.T, log *mock.MockLogger) {
				t.Helper()
				log.EXPECT().Info(gomock.Any()).AnyTimes()
				log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
				log.EXPECT().Warn("got invalid message", gomock.Cond(func(x []any) bool {
					if len(x) != 2 {
						return false
					}
					return x[0] == "expectedMessage"
				})).Times(2)
				log.EXPECT().Error("send bad message error", gomock.Any())
			},
			assert: func(t *testing.T, ctx context.Context, adp *adapter.Adapter) {
				t.Helper()
				read := adp.Reader(ctx)

				require.NoError(t, adp.SendConnRequest(endpoint, hostname, qshare.LaptopDevice))
				require.NoError(t, adp.SendClientInitWithClientFinished())
				resp, err := read()
				require.NoError(t, err)
				require.NoError(t, adp.ValidateServerInit(resp))

				require.NoError(t, adp.SendPairedKeyEncryption())
				resp, err = read()
				require.NoError(t, err)
				assertBadMessage(t, resp)
			},
		},
		"error/no_paired_key_encryption": {
			preapre: func(t *testing.T, log *mock.MockLogger) {
				t.Helper()
				log.EXPECT().Info(gomock.Any()).AnyTimes()
				log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
				log.EXPECT().Warn("got invalid message", gomock.Cond(func(x []any) bool {
					if len(x) != 2 {
						return false
					}
					return x[0] == "expectedMessage"
				})).Times(2)
				log.EXPECT().Error("send bad message error", gomock.Any())
			},
			assert: func(t *testing.T, ctx context.Context, adp *adapter.Adapter) {
				t.Helper()
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
				adp.EnableEncryption()

				require.NoError(t, adp.SendPairedKeyResult())
				resp, err = read()
				require.NoError(t, err)
				assertBadMessage(t, resp)
			},
		},
		"error/no_paired_key_result": {
			preapre: func(t *testing.T, log *mock.MockLogger) {
				t.Helper()
				log.EXPECT().Info(gomock.Any()).AnyTimes()
				log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
				log.EXPECT().Warn("got invalid message", gomock.Cond(func(x []any) bool {
					if len(x) != 2 {
						return false
					}
					return x[0] == "expectedMessage"
				})).Times(2)
				log.EXPECT().Error("send bad message error", gomock.Any())
			},
			assert: func(t *testing.T, ctx context.Context, adp *adapter.Adapter) {
				t.Helper()
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
				adp.EnableEncryption()

				require.NoError(t, adp.SendPairedKeyEncryption())
				resp, err = read()
				require.NoError(t, err)
				require.NoError(t, adp.ValidatePairedKeyEncryption(resp))

				require.NoError(t, adp.SendPairedKeyEncryption())
				resp, err = read()
				require.NoError(t, err)
				assertBadMessage(t, resp)
			},
		},
	}
	for name, tt := range tests {
		if name != "error/no_connection_requires" {
			continue
		}
		t.Run(name, func(t *testing.T) {
			ctrl, ctx := gomock.WithContext(t.Context(), t)
			t.Cleanup(ctrl.Finish)

			logger := mock.NewMockLogger(ctrl)
			if tt.preapre != nil {
				tt.preapre(t, logger)
			}

			server, err := qserver.NewBuilder().
				WithRandom(rng).
				WithPort(port).
				WithHostname(hostname).
				WithEndpoint(endpoint).
				WithLogger(logger).
				Build(nil, nil)
			require.NoError(t, err)

			require.NoError(t, server.Listen())
			t.Cleanup(func() {
				require.NoError(t, server.Stop())
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
