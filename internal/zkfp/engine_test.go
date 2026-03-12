package zkfp

import (
	"testing"
	"time"
)

func TestNewEngine(t *testing.T) {
	eng, err := NewEngine("")
	if err != nil {
		if err != errWindowsOnly {
			t.Logf("NewEngine error (may be driver/sensor): %v", err)
		}
		return
	}
	defer eng.Close()
	if _, err := eng.Init(); err != nil {
		t.Logf("Init (expected if no sensor): %v", err)
		return
	}
	// Stub or real: CaptureTemplate with short timeout
	_, _, err = eng.CaptureTemplate(100 * time.Millisecond)
	if err != nil && err != errWindowsOnly {
		t.Logf("CaptureTemplate (expected timeout or no sensor): %v", err)
	}
}
