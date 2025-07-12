package helper

import (
	"testing"
	"time"
)

func WaitForMDNSTTL(t *testing.T) {
	t.Helper()
	if !IsCI() {
		time.Sleep(3 * time.Second)
	}
}
