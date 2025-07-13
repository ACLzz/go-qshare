package server

import (
	"fmt"
	"io"
	"net"
	"sync"
	"testing"

	"github.com/ACLzz/qshare"
	"github.com/ACLzz/qshare/internal/adapter"
	"github.com/ACLzz/qshare/internal/log"
	logMock "github.com/ACLzz/qshare/internal/mock"
	"github.com/ACLzz/qshare/internal/rand"
	"github.com/ACLzz/qshare/internal/tests/helper"
	"github.com/ACLzz/qshare/qserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var spies sync.Map

func TestServer_Transfer(t *testing.T) {
	var (
		rng      = rand.NewStatic(1)
		port     = helper.RandomPort()
		endpoint = "test"
		hostname = "text_test"
	)
	tests := map[string]struct {
		introductionFrame adapter.IntroductionFrame
		textCallback      func(t *testing.T, payload qshare.TextPayload)
		fileCallback      func(t *testing.T, payload qshare.FilePayload)
		preapre           func(t *testing.T, log *logMock.MockLogger)
		assert            func(t *testing.T, read func() ([]byte, error), adp *adapter.Adapter)
	}{
		"success/text": {
			textCallback: func(t *testing.T, payload qshare.TextPayload) {
				assert.Equal(t, "Hello World!", payload.Text)

				assert.Equal(t, "Hello...", payload.Meta.Title)
				assert.Equal(t, int64(12), payload.Meta.Size)
				assert.Equal(t, qshare.TextText, payload.Meta.Type)
			},
			introductionFrame: adapter.IntroductionFrame{
				Text: &adapter.TextMeta{
					TextMeta: qshare.TextMeta{
						Type:  qshare.TextText,
						Title: "Hello...",
						Size:  12,
					},
					ID: 1,
				},
			},
			preapre: func(t *testing.T, log *logMock.MockLogger) {
				t.Helper()
				log.EXPECT().Info(gomock.Any()).AnyTimes()
				log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
			},
			assert: func(t *testing.T, read func() ([]byte, error), adp *adapter.Adapter) {
				require.NoError(t, adp.SendDataInChunks(1, []byte("Hello World!")))
			},
		},
		"success/file": {
			fileCallback: func(t *testing.T, payload qshare.FilePayload) {
				buf := make([]byte, payload.Meta.Size)
				_, err := io.ReadFull(payload.Pr, buf)
				require.NoError(t, err)
				assert.Equal(t, []byte("Hello World!"), buf)
			},
			introductionFrame: adapter.IntroductionFrame{
				Files: map[int64]*qshare.FilePayload{
					1: {
						Meta: qshare.FileMeta{
							Type:     qshare.FileImage,
							Name:     "image.jpg",
							MimeType: "image/jpg",
							Size:     12,
						},
					},
				},
			},
			preapre: func(t *testing.T, log *logMock.MockLogger) {
				t.Helper()
				log.EXPECT().Info(gomock.Any()).AnyTimes()
				log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
			},
			assert: func(t *testing.T, read func() ([]byte, error), adp *adapter.Adapter) {
				pr, pw := io.Pipe()
				defer func() {
					require.NoError(t, pw.Close())
				}()
				go func() {
					_, err := pw.Write([]byte("Hello World!"))
					require.NoError(t, err)
				}()

				require.NoError(t, adp.SendFileInChunks(1, qshare.FilePayload{
					Meta: qshare.FileMeta{
						Type:     qshare.FileImage,
						Name:     "image.jpg",
						MimeType: "image/jpg",
						Size:     12,
					},
					Pr: pr,
					Pw: pw,
				}))
			},
		},
		"success/many": {
			fileCallback: func(t *testing.T, payload qshare.FilePayload) {
				if val, ok := spies.Load(t.Name()); ok {
					spies.Swap(t.Name(), val.(int)+1)
				} else {
					spies.Store(t.Name(), 1)
				}

				buf := make([]byte, payload.Meta.Size)
				_, err := io.ReadFull(payload.Pr, buf)
				require.NoError(t, err)

				switch payload.Meta.Name {
				case "image.jpg":
					assert.Equal(t, []byte("Hello World!"), buf)
				case "image2.jpg":
					assert.Equal(t, []byte("Hello World2!"), buf)
				default:
					t.Error("unknown file")
				}
			},
			textCallback: func(t *testing.T, payload qshare.TextPayload) {
				assert.Equal(t, "Hello World3!", payload.Text)

				assert.Equal(t, "Hello...", payload.Meta.Title)
				assert.Equal(t, int64(13), payload.Meta.Size)
				assert.Equal(t, qshare.TextText, payload.Meta.Type)
			},
			introductionFrame: adapter.IntroductionFrame{
				Text: &adapter.TextMeta{
					TextMeta: qshare.TextMeta{
						Type:  qshare.TextText,
						Title: "Hello...",
						Size:  13,
					},
					ID: 3,
				},
				Files: map[int64]*qshare.FilePayload{
					1: {
						Meta: qshare.FileMeta{
							Type:     qshare.FileImage,
							Name:     "image.jpg",
							MimeType: "image/jpg",
							Size:     12,
						},
					},
					2: {
						Meta: qshare.FileMeta{
							Type:     qshare.FileImage,
							Name:     "image2.jpg",
							MimeType: "image/jpg",
							Size:     13,
						},
					},
				},
			},
			preapre: func(t *testing.T, log *logMock.MockLogger) {
				t.Helper()
				log.EXPECT().Info(gomock.Any()).AnyTimes()
				log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
				t.Cleanup(func() {
					val, ok := spies.Load(t.Name())
					assert.True(t, ok)
					assert.Equal(t, 2, val.(int))
				})
			},
			assert: func(t *testing.T, read func() ([]byte, error), adp *adapter.Adapter) {
				pr1, pw1 := io.Pipe()
				pr2, pw2 := io.Pipe()
				defer func() {
					require.NoError(t, pw1.Close())
					require.NoError(t, pw2.Close())
				}()
				go func() {
					_, err := pw1.Write([]byte("Hello World!"))
					require.NoError(t, err)
					_, err = pw2.Write([]byte("Hello World2!"))
					require.NoError(t, err)
				}()

				require.NoError(t, adp.SendFileInChunks(1, qshare.FilePayload{
					Meta: qshare.FileMeta{
						Type:     qshare.FileImage,
						Name:     "image.jpg",
						MimeType: "image/jpg",
						Size:     12,
					},
					Pr: pr1,
					Pw: pw1,
				}))
				require.NoError(t, adp.SendFileInChunks(2, qshare.FilePayload{
					Meta: qshare.FileMeta{
						Type:     qshare.FileImage,
						Name:     "image2.jpg",
						MimeType: "image/jpg",
						Size:     13,
					},
					Pr: pr2,
					Pw: pw2,
				}))
				require.NoError(t, adp.SendDataInChunks(3, []byte("Hello World3!")))
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

			textCallback := func(payload qshare.TextPayload) {
				if tt.textCallback != nil {
					tt.textCallback(t, payload)
				} else {
					t.Error("text callback mustn't be called in this test")
				}
			}
			fileCallback := func(payload qshare.FilePayload) {
				if tt.fileCallback != nil {
					tt.fileCallback(t, payload)
				} else {
					t.Error("file callback mustn't be called in this test")
				}
			}
			server, err := qserver.NewBuilder().
				WithRandom(rng).
				WithPort(port).
				WithHostname(hostname).
				WithEndpoint(endpoint).
				WithLogger(logger).
				Build(helper.StubAuthCallback(true), textCallback, fileCallback)
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
			require.NoError(t, adp.SendIntroduction(tt.introductionFrame))
			require.NoError(t, adp.SendTransferRequest())

			resp, err = read()
			require.NoError(t, err)
			isAccepted, err := adp.UnmarshalTransferResponse(resp)
			require.NoError(t, err)
			require.True(t, isAccepted)

			t.Cleanup(adp.Disconnect)
			tt.assert(t, read, &adp)
		})
	}
}
