//go:build windows

package zkfp

import (
	"syscall"
	"unsafe"
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	procPeekMessageW = user32.NewProc("PeekMessageW")
	procGetMessageW  = user32.NewProc("GetMessageW")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procDispatchMessageW = user32.NewProc("DispatchMessageW")
)

type msg struct {
	Hwnd    syscall.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

// runMessageLoop processes pending Windows messages (one iteration). Call periodically so COM/ActiveX events can be delivered.
func runMessageLoop() {
	const PM_REMOVE = 1
	var m msg
	r, _, _ := procPeekMessageW.Call(
		uintptr(unsafe.Pointer(&m)),
		0, 0, 0, PM_REMOVE,
	)
	if r == 0 {
		return
	}
	procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
	procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
}
