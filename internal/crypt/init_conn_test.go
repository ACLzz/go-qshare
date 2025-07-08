package crypt

import (
	"crypto/aes"
	"crypto/ecdh"
	"crypto/hmac"
	"crypto/sha256"
	"testing"

	pbSecureMessage "github.com/ACLzz/qshare/internal/protobuf/gen/securemessage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCipher_Server_Setup(t *testing.T) {
	type (
		want struct {
			err           error
			serverHMACKey []byte
			clientHMACKey []byte
			decryptKey    []byte
			encryptKey    []byte
		}
	)

	tests := map[string]struct {
		prepare  func(t *testing.T, c *Cipher)
		expected want
	}{
		"success": { // TODO: remove this test on release with a new one
			/*
				pin_code: 255, 210, 73, 71, 222, 149, 243, 117, 126, 94, 112, 81, 122, 237, 220, 105, 95, 142, 240, 72, 175, 45, 132, 96, 34, 225, 151, 23, 103, 141, 155, 151
			*/
			prepare: func(t *testing.T, c *Cipher) {
				t.Helper()
				require.NoError(t, c.SetSenderInitMessage([]byte{8, 2, 18, 131, 1, 8, 1, 18, 32, 252, 161, 0, 84, 99, 43, 27, 77, 180, 100, 22, 28, 198, 252, 240, 156, 130, 15, 167, 128, 166, 43, 47, 96, 128, 120, 152, 223, 130, 235, 245, 195, 26, 68, 8, 100, 18, 64, 102, 248, 202, 190, 149, 126, 95, 187, 26, 233, 109, 246, 155, 98, 7, 154, 221, 252, 63, 148, 63, 221, 68, 28, 162, 71, 18, 23, 51, 60, 234, 43, 168, 197, 33, 219, 98, 28, 63, 57, 89, 59, 176, 253, 117, 230, 64, 214, 170, 50, 130, 8, 170, 42, 83, 20, 191, 39, 99, 87, 191, 111, 214, 37, 34, 23, 65, 69, 83, 95, 50, 53, 54, 95, 67, 66, 67, 45, 72, 77, 65, 67, 95, 83, 72, 65, 50, 53, 54}))
				// require.NoError(t, c.AddReceiverInitMessage([]byte{8, 3, 18, 112, 8, 1, 18, 32, 134, 192, 88, 54, 170, 176, 100, 167, 16, 204, 84, 128, 91, 178, 184, 185, 246, 212, 211, 91, 108, 198, 80, 84, 53, 78, 61, 229, 18, 133, 55, 150, 24, 100, 34, 72, 8, 1, 18, 68, 10, 32, 80, 236, 88, 71, 116, 4, 29, 87, 173, 177, 245, 32, 148, 64, 49, 254, 79, 171, 3, 160, 127, 89, 155, 111, 91, 6, 125, 151, 79, 26, 156, 132, 18, 32, 54, 115, 237, 135, 57, 251, 151, 104, 108, 189, 183, 115, 31, 137, 163, 172, 75, 232, 207, 207, 86, 86, 197, 148, 165, 219, 8, 30, 19, 101, 85, 59}))
				require.NoError(t, c.SetSenderPublicKey(&pbSecureMessage.EcP256PublicKey{
					X: []byte{0, 145, 118, 119, 31, 168, 169, 189, 27, 12, 162, 185, 82, 204, 183, 70, 206, 6, 188, 126, 133, 98, 113, 119, 72, 130, 94, 74, 243, 209, 118, 197, 132},
					Y: []byte{0, 215, 163, 28, 69, 145, 230, 14, 19, 151, 130, 229, 164, 131, 99, 162, 188, 237, 10, 1, 190, 109, 11, 1, 83, 8, 28, 47, 137, 254, 229, 141, 80},
				}))
				privateKey, err := ecdh.P256().NewPrivateKey([]byte{56, 255, 228, 203, 143, 183, 59, 92, 156, 90, 174, 101, 212, 254, 104, 133, 141, 41, 118, 201, 254, 16, 227, 88, 142, 133, 227, 246, 255, 221, 230, 246})
				require.NoError(t, err)
				c.receiverPrivateKey = privateKey
				c.receiverInitMsg = []byte{8, 3, 18, 112, 8, 1, 18, 32, 134, 192, 88, 54, 170, 176, 100, 167, 16, 204, 84, 128, 91, 178, 184, 185, 246, 212, 211, 91, 108, 198, 80, 84, 53, 78, 61, 229, 18, 133, 55, 150, 24, 100, 34, 72, 8, 1, 18, 68, 10, 32, 80, 236, 88, 71, 116, 4, 29, 87, 173, 177, 245, 32, 148, 64, 49, 254, 79, 171, 3, 160, 127, 89, 155, 111, 91, 6, 125, 151, 79, 26, 156, 132, 18, 32, 54, 115, 237, 135, 57, 251, 151, 104, 108, 189, 183, 115, 31, 137, 163, 172, 75, 232, 207, 207, 86, 86, 197, 148, 165, 219, 8, 30, 19, 101, 85, 59}
			},
			expected: want{
				serverHMACKey: []byte{223, 162, 25, 212, 110, 139, 238, 129, 79, 230, 89, 111, 155, 8, 46, 208, 135, 236, 58, 9, 222, 0, 87, 25, 150, 57, 78, 167, 215, 182, 199, 29},
				clientHMACKey: []byte{95, 173, 137, 8, 78, 81, 108, 45, 1, 252, 72, 26, 0, 203, 190, 28, 150, 203, 157, 124, 248, 165, 139, 54, 129, 209, 172, 70, 116, 227, 71, 157},
				decryptKey:    []byte{195, 72, 84, 207, 73, 134, 222, 67, 99, 253, 247, 24, 94, 198, 181, 208, 68, 19, 136, 230, 182, 2, 11, 12, 127, 125, 30, 163, 220, 153, 208, 89},
				encryptKey:    []byte{53, 14, 18, 240, 254, 45, 218, 244, 94, 127, 6, 32, 179, 147, 181, 91, 93, 215, 133, 242, 82, 237, 95, 194, 159, 26, 17, 130, 118, 110, 45, 2},
			},
		},
		"error/nothing_set": {
			expected: want{
				err: ErrInvalidCipher,
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := NewCipher(true)
			if tt.prepare != nil {
				tt.prepare(t, &c)
			}

			err := c.Setup()
			if tt.expected.err != nil {
				assert.ErrorIs(t, err, tt.expected.err)
				return
			}
			require.NoError(t, err)

			d2dDecryptBlock, err := aes.NewCipher(tt.expected.decryptKey)
			require.NoError(t, err)
			d2dEncryptBlock, err := aes.NewCipher(tt.expected.encryptKey)
			require.NoError(t, err)

			assert.Equal(t, hmac.New(sha256.New, tt.expected.serverHMACKey), c.receiverHMAC)
			assert.Equal(t, hmac.New(sha256.New, tt.expected.clientHMACKey), c.senderHMAC)
			assert.EqualValues(t, d2dDecryptBlock, c.decryptBlock)
			assert.EqualValues(t, d2dEncryptBlock, c.encryptBlock)
		})
	}
}
