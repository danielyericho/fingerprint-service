//go:build !windows

package zkfp

import (
	"errors"
	"time"
)

var errWindowsOnly = errors.New("zkfp engine is only supported on Windows")

// Engine stub for non-Windows.
type Engine struct{}

// NewEngine returns error on non-Windows.
func NewEngine(progID string) (*Engine, error) {
	return nil, errWindowsOnly
}

// Init returns error on non-Windows.
func (e *Engine) Init() (int32, error) {
	return 0, errWindowsOnly
}

// Close is a no-op on stub.
func (e *Engine) Close() error { return nil }

// SetFPEngineVersion is a no-op.
func (e *Engine) SetFPEngineVersion(ver string) error { return nil }

// BeginCapture returns error.
func (e *Engine) BeginCapture() error { return errWindowsOnly }

// CancelCapture returns error.
func (e *Engine) CancelCapture() error { return errWindowsOnly }

// BeginEnroll returns error.
func (e *Engine) BeginEnroll() error { return errWindowsOnly }

// CancelEnroll returns error.
func (e *Engine) CancelEnroll() error { return errWindowsOnly }

// SetEnrollCount is a no-op.
func (e *Engine) SetEnrollCount(n int32) error { return nil }

// GetTemplateAsString returns error.
func (e *Engine) GetTemplateAsString() (string, error) { return "", errWindowsOnly }

// GetTemplateAsStringEx returns error.
func (e *Engine) GetTemplateAsStringEx(engineVersion string) (string, error) { return "", errWindowsOnly }

// VerFingerFromStr returns error.
func (e *Engine) VerFingerFromStr(reg, ver string, doLearning bool) (bool, error) { return false, errWindowsOnly }

// CreateFPCacheDBEx returns error.
func (e *Engine) CreateFPCacheDBEx() (int32, error) { return 0, errWindowsOnly }

// FreeFPCacheDBEx returns error.
func (e *Engine) FreeFPCacheDBEx(handle int32) error { return errWindowsOnly }

// AddRegTemplateStrToFPCacheDBEx returns error.
func (e *Engine) AddRegTemplateStrToFPCacheDBEx(fpcHandle, fpID int32, t9, t10 string) (int32, error) {
	return -1, errWindowsOnly
}

// IdentificationFromStrInFPCacheDB returns error.
func (e *Engine) IdentificationFromStrInFPCacheDB(fpcHandle int32, verTemplateStr string) (fpID, score, processed int32, err error) {
	return -1, 0, 0, errWindowsOnly
}

// CaptureTemplate returns error.
func (e *Engine) CaptureTemplate(timeout time.Duration) (template9, template10 string, err error) {
	return "", "", errWindowsOnly
}

// EnrollTemplate returns error.
func (e *Engine) EnrollTemplate(enrollCount int, timeout time.Duration) (template9, template10 string, err error) {
	return "", "", errWindowsOnly
}
