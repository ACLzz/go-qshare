package server

import (
	"fmt"
	"net"
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
				require.NoError(t, adp.SendIntroduction(adapter.IntroductionFrame{
					Text: &adapter.TextMeta{
						TextMeta: qshare.TextMeta{
							Type:  qshare.TextText,
							Title: "text...",
							Size:  4,
						},
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
				assert.Empty(t, files)
				assert.NotZero(t, pin)
				assert.Equal(t, "text...", text.Title)
				assert.Equal(t, int64(4), text.Size)
				assert.Equal(t, qshare.TextText, text.Type)
				return true
			},
		},
		"success/text/url": {
			preapre: func(t *testing.T, log *logMock.MockLogger) {
				t.Helper()
				log.EXPECT().Info(gomock.Any()).AnyTimes()
				log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
			},
			assert: func(t *testing.T, read func() ([]byte, error), adp *adapter.Adapter) {
				require.NoError(t, adp.SendIntroduction(adapter.IntroductionFrame{
					Text: &adapter.TextMeta{
						TextMeta: qshare.TextMeta{
							Type:  qshare.TextURL,
							Title: "https://example.org/...",
							Size:  int64(25),
						},
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
				assert.Empty(t, files)
				assert.NotZero(t, pin)
				assert.Equal(t, "https://example.org/...", text.Title)
				assert.Equal(t, int64(25), text.Size)
				assert.Equal(t, qshare.TextURL, text.Type)
				return true
			},
		},
		"success/text/address": {
			preapre: func(t *testing.T, log *logMock.MockLogger) {
				t.Helper()
				log.EXPECT().Info(gomock.Any()).AnyTimes()
				log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
			},
			assert: func(t *testing.T, read func() ([]byte, error), adp *adapter.Adapter) {
				require.NoError(t, adp.SendIntroduction(adapter.IntroductionFrame{
					Text: &adapter.TextMeta{
						TextMeta: qshare.TextMeta{
							Type:  qshare.TextAddress,
							Title: "USA, NY, Brooklyn...",
							Size:  int64(32),
						},
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
				assert.Empty(t, files)
				assert.NotZero(t, pin)
				assert.Equal(t, "USA, NY, Brooklyn...", text.Title)
				assert.Equal(t, int64(32), text.Size)
				assert.Equal(t, qshare.TextAddress, text.Type)
				return true
			},
		},
		"success/text/phone_number": {
			preapre: func(t *testing.T, log *logMock.MockLogger) {
				t.Helper()
				log.EXPECT().Info(gomock.Any()).AnyTimes()
				log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
			},
			assert: func(t *testing.T, read func() ([]byte, error), adp *adapter.Adapter) {
				require.NoError(t, adp.SendIntroduction(adapter.IntroductionFrame{
					Text: &adapter.TextMeta{
						TextMeta: qshare.TextMeta{
							Type:  qshare.TextText,
							Title: "+380683412...",
							Size:  int64(12),
						},
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
				assert.Empty(t, files)
				assert.NotZero(t, pin)
				assert.Equal(t, "+380683412...", text.Title)
				assert.Equal(t, int64(12), text.Size)
				assert.Equal(t, qshare.TextText, text.Type)
				return true
			},
		},
		"success/file/image": {
			preapre: func(t *testing.T, log *logMock.MockLogger) {
				t.Helper()
				log.EXPECT().Info(gomock.Any()).AnyTimes()
				log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
			},
			assert: func(t *testing.T, read func() ([]byte, error), adp *adapter.Adapter) {
				require.NoError(t, adp.SendIntroduction(adapter.IntroductionFrame{
					Files: map[int64]*qshare.FilePayload{
						1: {
							Meta: qshare.FileMeta{
								Type:     qshare.FileImage,
								Name:     "image.jpg",
								MimeType: "image/jpg",
								Size:     int64(12345),
							},
						},
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
				assert.Nil(t, text)
				assert.NotZero(t, pin)
				assert.Len(t, files, 1)
				assert.Equal(t, "image.jpg", files[0].Name)
				assert.Equal(t, qshare.FileImage, files[0].Type)
				assert.Equal(t, "image/jpg", files[0].MimeType)
				assert.Equal(t, int64(12345), files[0].Size)
				return true
			},
		},
		"success/file/video": {
			preapre: func(t *testing.T, log *logMock.MockLogger) {
				t.Helper()
				log.EXPECT().Info(gomock.Any()).AnyTimes()
				log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
			},
			assert: func(t *testing.T, read func() ([]byte, error), adp *adapter.Adapter) {
				require.NoError(t, adp.SendIntroduction(adapter.IntroductionFrame{
					Files: map[int64]*qshare.FilePayload{
						1: {
							Meta: qshare.FileMeta{
								Type:     qshare.FileVideo,
								Name:     "video.mp4",
								MimeType: "video/mp4",
								Size:     int64(1234),
							},
						},
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
				assert.Nil(t, text)
				assert.NotZero(t, pin)
				assert.Len(t, files, 1)
				assert.Equal(t, "video.mp4", files[0].Name)
				assert.Equal(t, qshare.FileVideo, files[0].Type)
				assert.Equal(t, "video/mp4", files[0].MimeType)
				assert.Equal(t, int64(1234), files[0].Size)
				return true
			},
		},
		"success/file/app": {
			preapre: func(t *testing.T, log *logMock.MockLogger) {
				t.Helper()
				log.EXPECT().Info(gomock.Any()).AnyTimes()
				log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
			},
			assert: func(t *testing.T, read func() ([]byte, error), adp *adapter.Adapter) {
				require.NoError(t, adp.SendIntroduction(adapter.IntroductionFrame{
					Files: map[int64]*qshare.FilePayload{
						1: {
							Meta: qshare.FileMeta{
								Type:     qshare.FileApp,
								Name:     "youtube.apk",
								MimeType: "application/vnd.android.package-archive",
								Size:     int64(123456),
							},
						},
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
				assert.Nil(t, text)
				assert.NotZero(t, pin)
				assert.Len(t, files, 1)
				assert.Equal(t, "youtube.apk", files[0].Name)
				assert.Equal(t, qshare.FileApp, files[0].Type)
				assert.Equal(t, "application/vnd.android.package-archive", files[0].MimeType)
				assert.Equal(t, int64(123456), files[0].Size)
				return true
			},
		},
		"success/file/audio": {
			preapre: func(t *testing.T, log *logMock.MockLogger) {
				t.Helper()
				log.EXPECT().Info(gomock.Any()).AnyTimes()
				log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
			},
			assert: func(t *testing.T, read func() ([]byte, error), adp *adapter.Adapter) {
				require.NoError(t, adp.SendIntroduction(adapter.IntroductionFrame{
					Files: map[int64]*qshare.FilePayload{
						1: {
							Meta: qshare.FileMeta{
								Type:     qshare.FileApp,
								Name:     "ringtone.wav",
								MimeType: "audio/wav",
								Size:     int64(123),
							},
						},
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
				assert.Nil(t, text)
				assert.NotZero(t, pin)
				assert.Len(t, files, 1)
				assert.Equal(t, "ringtone.wav", files[0].Name)
				assert.Equal(t, qshare.FileApp, files[0].Type)
				assert.Equal(t, "audio/wav", files[0].MimeType)
				assert.Equal(t, int64(123), files[0].Size)
				return true
			},
		},
		"success/file/unknown": {
			preapre: func(t *testing.T, log *logMock.MockLogger) {
				t.Helper()
				log.EXPECT().Info(gomock.Any()).AnyTimes()
				log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
			},
			assert: func(t *testing.T, read func() ([]byte, error), adp *adapter.Adapter) {
				require.NoError(t, adp.SendIntroduction(adapter.IntroductionFrame{
					Files: map[int64]*qshare.FilePayload{
						1: {
							Meta: qshare.FileMeta{
								Type:     qshare.FileUnknown,
								Name:     "cv.pdf",
								MimeType: "application/pdf",
								Size:     int64(1233),
							},
						},
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
				assert.Nil(t, text)
				assert.NotZero(t, pin)
				assert.Len(t, files, 1)
				assert.Equal(t, "cv.pdf", files[0].Name)
				assert.Equal(t, qshare.FileUnknown, files[0].Type)
				assert.Equal(t, "application/pdf", files[0].MimeType)
				assert.Equal(t, int64(1233), files[0].Size)
				return true
			},
		},
		"success/many_files_and_text": {
			preapre: func(t *testing.T, log *logMock.MockLogger) {
				t.Helper()
				log.EXPECT().Info(gomock.Any()).AnyTimes()
				log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
			},
			assert: func(t *testing.T, read func() ([]byte, error), adp *adapter.Adapter) {
				require.NoError(t, adp.SendIntroduction(adapter.IntroductionFrame{
					Text: &adapter.TextMeta{
						TextMeta: qshare.TextMeta{
							Type:  qshare.TextURL,
							Title: "https://example.org/...",
							Size:  int64(25),
						},
					},
					Files: map[int64]*qshare.FilePayload{
						1: {
							Meta: qshare.FileMeta{
								Type:     qshare.FileUnknown,
								Name:     "cv.pdf",
								MimeType: "application/pdf",
								Size:     int64(1233),
							},
						},
						2: {
							Meta: qshare.FileMeta{
								Type:     qshare.FileApp,
								Name:     "ringtone.wav",
								MimeType: "audio/wav",
								Size:     int64(123),
							},
						},
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
				assert.NotZero(t, pin)
				assert.Equal(t, "https://example.org/...", text.Title)
				assert.Equal(t, int64(25), text.Size)
				assert.Equal(t, qshare.TextURL, text.Type)

				assert.Len(t, files, 2)
				cvID := 0
				ringtoneID := 1
				if files[1].Name == "cv.pdf" {
					cvID = 1
					ringtoneID = 0
				}
				assert.Equal(t, "cv.pdf", files[cvID].Name)
				assert.Equal(t, qshare.FileUnknown, files[cvID].Type)
				assert.Equal(t, "application/pdf", files[cvID].MimeType)
				assert.Equal(t, int64(1233), files[cvID].Size)

				assert.Equal(t, "ringtone.wav", files[ringtoneID].Name)
				assert.Equal(t, qshare.FileApp, files[ringtoneID].Type)
				assert.Equal(t, "audio/wav", files[ringtoneID].MimeType)
				assert.Equal(t, int64(123), files[ringtoneID].Size)
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
