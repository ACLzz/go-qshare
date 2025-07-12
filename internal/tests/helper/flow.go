package helper

import (
	"testing"
	"time"

	"github.com/ACLzz/qshare"
)

func NewStubAuthCallback(resp bool) qshare.AuthCallback {
	return func(text *qshare.TextMeta, files []qshare.FileMeta, pin uint16) bool {
		return resp
	}
}

func WaitForMDNSTTL(t *testing.T) {
	t.Helper()
	if !IsCI() {
		time.Sleep(3 * time.Second)
	}
}
