//go:build windows

package autotype

import (
	"fmt"
	"log/slog"
	"time"
	"unsafe"

	"github.com/c/just-talk-go/internal/clipboard"
	"golang.org/x/sys/windows"
)

var (
	user32        = windows.NewLazySystemDLL("user32.dll")
	procSendInput = user32.NewProc("SendInput")
	procVkKeyScan = user32.NewProc("VkKeyScanW")
)

type keyboardInput struct {
	Wvk     uint16
	Wscan   uint16
	DwFlags uint32
	Time    uint32
	DwExtra uintptr
}

type input struct {
	Type uint32
	Ki   keyboardInput
	_    [8]byte
}

const (
	inputKeyboard  = 1
	keyeventfKeyUp = 2
)

func pastePlatform(text string, logger *slog.Logger) error {
	cb, err := clipboard.New()
	if err != nil {
		return fmt.Errorf("clipboard: %w", err)
	}
	if err := cb.Set(text); err != nil {
		return fmt.Errorf("set clipboard: %w", err)
	}

	time.Sleep(50 * time.Millisecond)
	if err := simulatePaste(); err != nil {
		return fmt.Errorf("simulate paste: %w", err)
	}
	logger.Debug("autotype done", "text_len", len(text), "method", pasteMethod())
	return nil
}

func simulatePaste() error {
	// Simulate Ctrl down → V down → V up → Ctrl up
	keys := []struct {
		code uint16
		up   bool
	}{
		{0x11, false}, // VK_CONTROL down
		{0x56, false}, // VK_V down
		{0x56, true},  // VK_V up
		{0x11, true},  // VK_CONTROL up
	}

	var inputs []input
	for _, k := range keys {
		flags := uint32(0)
		if k.up {
			flags = keyeventfKeyUp
		}
		inputs = append(inputs, input{
			Type: inputKeyboard,
			Ki: keyboardInput{
				Wvk:     k.code,
				DwFlags: flags,
			},
		})
	}

	cbSize := unsafe.Sizeof(input{})
	r1, _, err := procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(cbSize),
	)
	if r1 == 0 {
		return fmt.Errorf("SendInput returned 0 (all inputs blocked or invalid struct size), error: %w", err)
	}
	return nil
}

func pasteMethod() string { return "windows/SendInput+Ctrl+V" }

func isWaylandSession() bool { return false }
