//go:build windows

package zkfp

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

type Engine struct {
	obj      *ole.IDispatch
	mu       sync.Mutex
	inited   bool
	taskChan chan func()
	quit     chan struct{}
}

func NewEngine(progID string) (*Engine, error) {
	if progID == "" {
		progID = defaultProgID
	}

	e := &Engine{
		taskChan: make(chan func()),
		quit:     make(chan struct{}),
	}

	initErrChan := make(chan error)

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		if err := ole.CoInitialize(0); err != nil {
			initErrChan <- fmt.Errorf("CoInitialize: %w", err)
			return
		}
		defer ole.CoUninitialize()

		clsid, err := ole.CLSIDFromProgID(progID)
		if err != nil {
			initErrChan <- fmt.Errorf("CLSIDFromProgID: %w", err)
			return
		}

		unknown, err := ole.CreateInstance(clsid, ole.IID_IDispatch)
		if err != nil {
			initErrChan <- fmt.Errorf("CreateInstance: %w", err)
			return
		}
		defer unknown.Release()

		disp, err := unknown.QueryInterface(ole.IID_IDispatch)
		if err != nil {
			initErrChan <- fmt.Errorf("QueryInterface: %w", err)
			return
		}
		
		e.obj = disp
		initErrChan <- nil // Lapor inisialisasi sukses!

		// 2. SISTEM ANTREAN TUGAS (Mencegah Mabuk Perintah)
		for {
			select {
			case <-e.quit:
				disp.Release()
				return
			case task := <-e.taskChan:
				task() // Eksekusi tugas dari HTTP Request
			default:
				// Pompa antrean pesan Windows agar hardware tidak nyangkut/deadlock
				runMessageLoop()
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	if err := <-initErrChan; err != nil {
		return nil, err
	}

	return e, nil
}

// execute mengirim tugas ke kamar isolasi dan menunggu sampai selesai
func (e *Engine) execute(fn func()) {
	done := make(chan struct{})
	e.taskChan <- func() {
		fn()
		close(done)
	}
	<-done
}

func (e *Engine) Init() (ret int32, err error) {
	e.execute(func() {
		v, eCall := oleutil.CallMethod(e.obj, "InitEngine")
		if eCall != nil {
			err = fmt.Errorf("InitEngine: %w", eCall)
			return
		}
		defer v.Clear()
		e.inited = true
		ret = variantToInt32(v)
	})
	return
}

func (e *Engine) Close() error {
	if e.obj != nil {
		e.execute(func() {
			_, _ = oleutil.CallMethod(e.obj, "EndEngine")
		})
		close(e.quit)
		e.obj = nil
	}
	e.inited = false
	return nil
}

func (e *Engine) SetFPEngineVersion(ver string) (err error) {
	e.execute(func() { _, err = oleutil.PutProperty(e.obj, "FPEngineVersion", ver) })
	return
}

func (e *Engine) BeginCapture() (err error) {
	e.execute(func() { _, err = oleutil.CallMethod(e.obj, "BeginCapture") })
	return
}

func (e *Engine) CancelCapture() (err error) {
	e.execute(func() { _, err = oleutil.CallMethod(e.obj, "CancelCapture") })
	return
}

func (e *Engine) BeginEnroll() (err error) {
	e.execute(func() { _, err = oleutil.CallMethod(e.obj, "BeginEnroll") })
	return
}

func (e *Engine) CancelEnroll() (err error) {
	e.execute(func() { _, err = oleutil.CallMethod(e.obj, "CancelEnroll") })
	return
}

func (e *Engine) SetEnrollCount(n int32) (err error) {
	e.execute(func() { _, err = oleutil.PutProperty(e.obj, "EnrollCount", n) })
	return
}

func (e *Engine) GetTemplateAsStringEx(engineVersion string) (ret string, err error) {
	e.execute(func() {
		v, eCall := oleutil.CallMethod(e.obj, "GetTemplateAsStringEx", engineVersion)
		if eCall != nil {
			err = eCall
			return
		}
		defer v.Clear()
		ret = v.ToString()
	})
	return
}

func (e *Engine) VerFingerFromStr(regTemplateStr, verTemplateStr string, doLearning bool) (match bool, err error) {
	e.execute(func() {
		regChanged := false
		v, eCall := oleutil.CallMethod(e.obj, "VerFingerFromStr", regTemplateStr, verTemplateStr, doLearning, &regChanged)
		if eCall != nil {
			err = eCall
			return
		}
		defer v.Clear()
		if b, ok := v.Value().(bool); ok {
			match = b
		}
	})
	return
}

func (e *Engine) CreateFPCacheDBEx() (ret int32, err error) {
	e.execute(func() {
		v, eCall := oleutil.CallMethod(e.obj, "CreateFPCacheDBEx")
		if eCall != nil {
			err = eCall
			return
		}
		defer v.Clear()
		ret = variantToInt32(v)
	})
	return
}

func (e *Engine) FreeFPCacheDBEx(handle int32) (err error) {
	e.execute(func() { _, err = oleutil.CallMethod(e.obj, "FreeFPCacheDBEx", handle) })
	return
}

func (e *Engine) AddRegTemplateStrToFPCacheDBEx(fpcHandle, fpID int32, template9, template10 string) (ret int32, err error) {
	e.execute(func() {
		v, eCall := oleutil.CallMethod(e.obj, "AddRegTemplateStrToFPCacheDBEx", fpcHandle, fpID, template9, template10)
		if eCall != nil {
			err = eCall
			return
		}
		defer v.Clear()
		ret = variantToInt32(v)
	})
	return
}

func (e *Engine) IdentificationFromStrInFPCacheDB(fpcHandle int32, verTemplateStr string) (fpID int32, score int32, processed int32, err error) {
	e.execute(func() {
		var scoreVar, processedVar int32
		v, eCall := oleutil.CallMethod(e.obj, "IdentificationFromStrInFPCacheDB", fpcHandle, verTemplateStr, &scoreVar, &processedVar)
		if eCall != nil {
			err = eCall
			return
		}
		defer v.Clear()
		fpID = variantToInt32(v)
		score = scoreVar
		processed = processedVar
	})
	return
}

func (e *Engine) CaptureTemplate(timeout time.Duration) (FPStringV9, FPStringV10 string, err error) {
	e.execute(func() {
		// Ambil patokan
		vOld, _ := oleutil.CallMethod(e.obj, "GetTemplateAsStringEx", "9")
		oldStringTemplate := ""
		if vOld != nil {
			oldStringTemplate = vOld.ToString()
			vOld.Clear()
		}

		// Nyalakan alat
		_, errBegin := oleutil.CallMethod(e.obj, "BeginCapture")
		if errBegin != nil {
			err = fmt.Errorf("BeginCapture: %w", errBegin)
			return
		}
		
		// Gembok otomatis
		defer func() {
			_, _ = oleutil.CallMethod(e.obj, "CancelCapture")
			runMessageLoop() 
		}()

		deadline := time.Now().Add(timeout)
		for time.Now().Before(deadline) {
			runMessageLoop()
			v9, _ := oleutil.CallMethod(e.obj, "GetTemplateAsStringEx", "9")
			if v9 != nil {
				FPStringV9 = v9.ToString()
				v9.Clear()
			}
			
			if FPStringV9 != "" && FPStringV9 != oldStringTemplate {
				v10, _ := oleutil.CallMethod(e.obj, "GetTemplateAsStringEx", "10")
				if v10 != nil {
					FPStringV10 = v10.ToString()
					v10.Clear()
				}
				return // Selesai! Fungsi HTTP akan di-resume
			}
			
			time.Sleep(100 * time.Millisecond)
		}

		err = fmt.Errorf("capture timeout after %v", timeout)
	})
	return
}

func (e *Engine) EnrollTemplate(enrollCount int, timeout time.Duration) (FPStringV9, FPStringV10 string, err error) {
	e.execute(func() {
		vOld, _ := oleutil.CallMethod(e.obj, "GetTemplateAsStringEx", "9")
		oldStringTemplate := ""
		if vOld != nil {
			oldStringTemplate = vOld.ToString()
			vOld.Clear()
		}

		_, _ = oleutil.PutProperty(e.obj, "EnrollCount", int32(enrollCount))
		_, errBegin := oleutil.CallMethod(e.obj, "BeginEnroll")
		if errBegin != nil {
			err = fmt.Errorf("BeginEnroll: %w", errBegin)
			return
		}
		
		defer func() {
			_, _ = oleutil.CallMethod(e.obj, "CancelEnroll")
			runMessageLoop()
		}()

		deadline := time.Now().Add(timeout)
		for time.Now().Before(deadline) {
			runMessageLoop()
			v9, _ := oleutil.CallMethod(e.obj, "GetTemplateAsStringEx", "9")
			if v9 != nil {
				FPStringV9 = v9.ToString()
				v9.Clear()
			}
			
			if FPStringV9 != "" && FPStringV9 != oldStringTemplate {
				v10, _ := oleutil.CallMethod(e.obj, "GetTemplateAsStringEx", "10")
				if v10 != nil {
					FPStringV10 = v10.ToString()
					v10.Clear()
				}
				return 
			}
			
			time.Sleep(100 * time.Millisecond)
		}

		err = fmt.Errorf("enroll timeout after %v", timeout)
	})
	return
}

func variantToInt32(v *ole.VARIANT) int32 {
	if v == nil { return 0 }
	val := v.Value()
	if val == nil { return 0 }
	switch x := val.(type) {
	case int32: return x
	case int64: return int32(x)
	case int: return int32(x)
	case uint32: return int32(x)
	case uint64: return int32(x)
	default: return 0
	}
}