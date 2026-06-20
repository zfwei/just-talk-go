//go:build windows

package clipboard

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modUser32 = windows.NewLazySystemDLL("user32.dll")
	modKernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procOpenClipboard   = modUser32.NewProc("OpenClipboard")
	procCloseClipboard  = modUser32.NewProc("CloseClipboard")
	procEmptyClipboard  = modUser32.NewProc("EmptyClipboard")
	procSetClipboardData = modUser32.NewProc("SetClipboardData")
	procGetClipboardData = modUser32.NewProc("GetClipboardData")

	procGlobalAlloc = modKernel32.NewProc("GlobalAlloc")
	procGlobalFree  = modKernel32.NewProc("GlobalFree")
	procGlobalLock  = modKernel32.NewProc("GlobalLock")
	procGlobalUnlock = modKernel32.NewProc("GlobalUnlock")
)

const (
	gmemMoveable  = 0x0002
	cfUnicodeText = 13
)

func newPlatformClipboard() (*Clipboard, error) {
	return &Clipboard{
		getFunc: winGet,
		setFunc: winSet,
	}, nil
}

func winSet(text string) error {
	utf16, err := windows.UTF16FromString(text)
	if err != nil {
		return err
	}

	r, _, err := procOpenClipboard.Call(0)
	if r == 0 {
		return fmt.Errorf("OpenClipboard failed: %w", err)
	}
	defer procCloseClipboard.Call()

	r, _, err = procEmptyClipboard.Call()
	if r == 0 {
		return fmt.Errorf("EmptyClipboard failed: %w", err)
	}

	size := uintptr(len(utf16)) * unsafe.Sizeof(utf16[0])
	hMem, _, err := procGlobalAlloc.Call(gmemMoveable, size)
	if hMem == 0 {
		return fmt.Errorf("GlobalAlloc failed: %w", err)
	}

	p, _, err := procGlobalLock.Call(hMem)
	if p == 0 {
		procGlobalFree.Call(hMem)
		return fmt.Errorf("GlobalLock failed: %w", err)
	}

	slice := unsafe.Slice((*uint16)(unsafe.Pointer(p)), len(utf16))
	copy(slice, utf16)

	procGlobalUnlock.Call(hMem)

	r, _, err = procSetClipboardData.Call(cfUnicodeText, hMem)
	if r == 0 {
		procGlobalFree.Call(hMem)
		return fmt.Errorf("SetClipboardData failed: %w", err)
	}

	return nil
}

func winGet() (string, error) {
	r, _, err := procOpenClipboard.Call(0)
	if r == 0 {
		return "", fmt.Errorf("OpenClipboard failed: %w", err)
	}
	defer procCloseClipboard.Call()

	hMem, _, err := procGetClipboardData.Call(cfUnicodeText)
	if hMem == 0 {
		return "", nil
	}

	p, _, err := procGlobalLock.Call(hMem)
	if p == 0 {
		return "", fmt.Errorf("GlobalLock failed: %w", err)
	}
	defer procGlobalUnlock.Call(hMem)

	// Traverse UTF-16 null-terminated string to find length
	var length int
	for {
		ptr := unsafe.Pointer(p + uintptr(length)*2)
		val := *(*uint16)(ptr)
		if val == 0 {
			break
		}
		length++
	}

	slice := unsafe.Slice((*uint16)(unsafe.Pointer(p)), length)
	return windows.UTF16ToString(slice), nil
}
