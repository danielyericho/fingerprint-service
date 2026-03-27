//go:build windows

package zkfp

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// Engine wraps the ZKFPEngX ActiveX control for Windows.
type Engine struct {
	obj    *ole.IDispatch
	mu     sync.Mutex
	inited bool
}

// NewEngine creates a new Engine. Call Init() before use.
func NewEngine(progID string) (*Engine, error) {
	if progID == "" {
		progID = defaultProgID
	}
	if err := ole.CoInitialize(0); err != nil {
		ole.CoUninitialize()
		ole.CoInitialize(0)
	}
	clsid, err := ole.CLSIDFromProgID(progID)
	if err != nil {
		return nil, fmt.Errorf("CLSIDFromProgID: %w (ensure ZKFinger driver/OCX is installed)", err)
	}
	unknown, err := ole.CreateInstance(clsid, ole.IID_IDispatch)
	if err != nil {
		return nil, fmt.Errorf("create ZKFPEngX: %w (ensure ZKFinger driver/OCX is installed)", err)
	}
	disp, err := unknown.QueryInterface(ole.IID_IDispatch)
	unknown.Release()
	if err != nil {
		return nil, fmt.Errorf("IDispatch: %w", err)
	}
	return &Engine{obj: disp}, nil
}

// Init initializes the fingerprint engine. Set engineVersion to "9" or "10" before or after.
func (e *Engine) Init() (int32, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	v, err := oleutil.CallMethod(e.obj, "InitEngine")
	if err != nil {
		return 0, fmt.Errorf("InitEngine: %w", err)
	}
	defer v.Clear()
	e.inited = true
	return variantToInt32(v), nil
}

// Close releases the engine and uninitializes COM.
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.obj != nil {
		_, _ = oleutil.CallMethod(e.obj, "EndEngine")
		e.obj.Release()
		e.obj = nil
	}
	e.inited = false
	ole.CoUninitialize()
	return nil
}

// SetFPEngineVersion sets algorithm version "9" or "10".
func (e *Engine) SetFPEngineVersion(ver string) error {
	_, err := oleutil.PutProperty(e.obj, "FPEngineVersion", ver)
	return err
}

// BeginCapture starts single-finger capture.
func (e *Engine) BeginCapture() error {
	_, err := oleutil.CallMethod(e.obj, "BeginCapture")
	return err
}

// CancelCapture cancels capture.
func (e *Engine) CancelCapture() error {
	_, err := oleutil.CallMethod(e.obj, "CancelCapture")
	return err
}

// BeginEnroll starts enrollment (multiple presses). Set EnrollCount first if needed.
func (e *Engine) BeginEnroll() error {
	_, err := oleutil.CallMethod(e.obj, "BeginEnroll")
	return err
}

// CancelEnroll cancels enrollment.
func (e *Engine) CancelEnroll() error {
	_, err := oleutil.CallMethod(e.obj, "CancelEnroll")
	return err
}

// SetEnrollCount sets number of presses for enrollment (e.g. 3).
func (e *Engine) SetEnrollCount(n int32) error {
	_, err := oleutil.PutProperty(e.obj, "EnrollCount", n)
	return err
}

// GetTemplateAsString returns the current template as string.
func (e *Engine) GetTemplateAsString() (string, error) {
	v, err := oleutil.CallMethod(e.obj, "GetTemplateAsString")
	if err != nil {
		return "", err
	}
	defer v.Clear()
	return v.ToString(), nil
}

// GetTemplateAsStringEx returns template for algorithm version "9" or "10".
func (e *Engine) GetTemplateAsStringEx(engineVersion string) (string, error) {
	v, err := oleutil.CallMethod(e.obj, "GetTemplateAsStringEx", engineVersion)
	if err != nil {
		return "", err
	}
	defer v.Clear()
	return v.ToString(), nil
}

// VerFingerFromStr performs 1:1 verification.
func (e *Engine) VerFingerFromStr(regTemplateStr, verTemplateStr string, doLearning bool) (bool, error) {
	regChanged := false
	v, err := oleutil.CallMethod(e.obj, "VerFingerFromStr", regTemplateStr, verTemplateStr, doLearning, &regChanged)
	if err != nil {
		return false, err
	}
	defer v.Clear()
	val := v.Value()
	if b, ok := val.(bool); ok {
		return b, nil
	}
	return false, nil
}

// CreateFPCacheDBEx creates an in-memory template cache for 1:N.
func (e *Engine) CreateFPCacheDBEx() (int32, error) {
	v, err := oleutil.CallMethod(e.obj, "CreateFPCacheDBEx")
	if err != nil {
		return 0, err
	}
	defer v.Clear()
	return variantToInt32(v), nil
}

// FreeFPCacheDBEx releases the cache.
func (e *Engine) FreeFPCacheDBEx(handle int32) error {
	_, err := oleutil.CallMethod(e.obj, "FreeFPCacheDBEx", handle)
	return err
}

// AddRegTemplateStrToFPCacheDBEx adds a template to the cache.
func (e *Engine) AddRegTemplateStrToFPCacheDBEx(fpcHandle, fpID int32, template9, template10 string) (int32, error) {
	v, err := oleutil.CallMethod(e.obj, "AddRegTemplateStrToFPCacheDBEx", fpcHandle, fpID, template9, template10)
	if err != nil {
		return -1, err
	}
	defer v.Clear()
	return variantToInt32(v), nil
}

// IdentificationFromStrInFPCacheDB does 1:N match.
func (e *Engine) IdentificationFromStrInFPCacheDB(fpcHandle int32, verTemplateStr string) (fpID int32, score int32, processed int32, err error) {
	var scoreVar, processedVar int32
	v, err := oleutil.CallMethod(e.obj, "IdentificationFromStrInFPCacheDB", fpcHandle, verTemplateStr, &scoreVar, &processedVar)
	if err != nil {
		return -1, 0, 0, err
	}
	defer v.Clear()
	fpID = variantToInt32(v)
	return fpID, scoreVar, processedVar, nil
}

// variantToInt32 extracts int32 from ole.VARIANT using Value() so VT and Val are handled correctly.
func variantToInt32(v *ole.VARIANT) int32 {
	if v == nil {
		return 0
	}
	val := v.Value()
	if val == nil {
		return 0
	}
	switch x := val.(type) {
	case int32:
		return x
	case int64:
		return int32(x)
	case int:
		return int32(x)
	case uint32:
		return int32(x)
	case uint64:
		return int32(x)
	default:
		return 0
	}
}

func (e *Engine) CaptureTemplate(timeout time.Duration) (FPStringV9, FPStringV10 string, err error) {
	oldStringTemplate, _ := e.GetTemplateAsStringEx("9")

	if err := e.BeginCapture(); err != nil {
		return "", "", err
	}
	
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		runMessageLoop()
		FPStringV9, _ := e.GetTemplateAsStringEx("9")
		
		if FPStringV9 != "" && FPStringV9 != oldStringTemplate {
			FPStringV10, _ := e.GetTemplateAsStringEx("10")
			return FPStringV9, FPStringV10, nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	_ = e.CancelCapture()
	return "", "", fmt.Errorf("capture timeout after %v", timeout)
}

func (e *Engine) EnrollTemplate(enrollCount int, timeout time.Duration) (FPStringV9, FPStringV10 string, err error) {
	oldStringTemplate, _ := e.GetTemplateAsStringEx("9")

	e.SetEnrollCount(int32(enrollCount))
	if err := e.BeginEnroll(); err != nil {
		return "", "", err
	}
	
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		runMessageLoop()
		FPStringV9, _ := e.GetTemplateAsStringEx("9")
		
		if FPStringV9 != "" && FPStringV9 != oldStringTemplate {
			FPStringV10, _ := e.GetTemplateAsStringEx("10")
			return FPStringV9, FPStringV10, nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	_ = e.CancelEnroll()
	return "", "", fmt.Errorf("enroll timeout after %v", timeout)
}