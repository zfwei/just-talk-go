//go:build windows

package overlay

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"github.com/c/just-talk-go/config"
	"golang.org/x/sys/windows"
)

const (
	basePillW  = 122
	basePillH  = 42
	baseMargin = 28
)

var (
	user32                     = windows.NewLazySystemDLL("user32.dll")
	procRegisterClassExW       = user32.NewProc("RegisterClassExW")
	procCreateWindowExW        = user32.NewProc("CreateWindowExW")
	procShowWindow             = user32.NewProc("ShowWindow")
	procUpdateWindow           = user32.NewProc("UpdateWindow")
	procGetMessageW            = user32.NewProc("GetMessageW")
	procTranslateMessage       = user32.NewProc("TranslateMessage")
	procDispatchMessageW       = user32.NewProc("DispatchMessageW")
	procDestroyWindow          = user32.NewProc("DestroyWindow")
	procDefWindowProcW         = user32.NewProc("DefWindowProcW")
	procPostQuitMessage        = user32.NewProc("PostQuitMessage")
	procSetWindowPos           = user32.NewProc("SetWindowPos")
	procSetLayeredWindowAttr   = user32.NewProc("SetLayeredWindowAttributes")
	procBeginPaint             = user32.NewProc("BeginPaint")
	procEndPaint               = user32.NewProc("EndPaint")
	procGetClientRect          = user32.NewProc("GetClientRect")
	procInvalidateRect         = user32.NewProc("InvalidateRect")
	procGetSystemMetrics       = user32.NewProc("GetSystemMetrics")
	procLoadCursorW            = user32.NewProc("LoadCursorW")
	procPostMessageW           = user32.NewProc("PostMessageW")
	procFillRect               = user32.NewProc("FillRect")
	procDrawTextW              = user32.NewProc("DrawTextW")
	procSystemParametersInfoW  = user32.NewProc("SystemParametersInfoW")

	kernel32             = windows.NewLazySystemDLL("kernel32.dll")
	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")

	gdi32                      = windows.NewLazySystemDLL("gdi32.dll")
	procCreateSolidBrush       = gdi32.NewProc("CreateSolidBrush")
	procDeleteObject           = gdi32.NewProc("DeleteObject")
	procRoundRect              = gdi32.NewProc("RoundRect")
	procCreateFontW            = gdi32.NewProc("CreateFontW")
	procSelectObject           = gdi32.NewProc("SelectObject")
	procSetTextColor           = gdi32.NewProc("SetTextColor")
	procSetBkMode              = gdi32.NewProc("SetBkMode")
	procGetStockObject         = gdi32.NewProc("GetStockObject")
)

type wndClassEx struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   uintptr
	Icon       windows.Handle
	Cursor     windows.Handle
	Background windows.Handle
	MenuName   *uint16
	ClassName  *uint16
	IconSm     windows.Handle
}

type rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type paintStruct struct {
	Hdc         windows.Handle
	Erase       int32
	Paint       rect
	Restore     int32
	IncUpdate   int32
	RgbReserved [32]byte
}

type msg struct {
	Hwnd    windows.HWND
	Message uint32
	Wparam  uintptr
	Lparam  uintptr
	Time    uint32
	Pt      point
}

type point struct {
	X int32
	Y int32
}

type windowsBackend struct {
	hwnd   windows.HWND
	label  string
	color  statusColor
	margin int
	w      int
	h      int
	pos    string
	mu     sync.Mutex
}

var activeBackend *windowsBackend

func newBackend(cfg config.OverlayConfig) (backend, error) {
	return newWindowsBackend(cfg)
}

func newWindowsBackend(cfg config.OverlayConfig) (backend, error) {
	scale := cfg.Scale
	if scale <= 0 {
		scale = 1.0
	}
	w := int(float64(basePillW) * scale)
	h := int(float64(basePillH) * scale)
	margin := int(float64(baseMargin) * scale)

	b := &windowsBackend{
		margin: margin,
		w:      w,
		h:      h,
		pos:    cfg.Position,
	}
	if b.pos == "" {
		b.pos = "bottom-center"
	}

	initChan := make(chan error, 1)

	go func() {
		runtime.LockOSThread()

		instance, _, _ := procGetModuleHandleW.Call(0)

		className := "JustTalkOverlayClass"
		classW := windows.StringToUTF16Ptr(className)

		cursor, _, _ := procLoadCursorW.Call(0, 32512) // IDC_ARROW = 32512

		wcex := wndClassEx{
			Size:      uint32(unsafe.Sizeof(wndClassEx{})),
			Style:     3, // CS_HREDRAW | CS_VREDRAW = 3
			WndProc:   syscall.NewCallback(wndProc),
			Instance:  instance,
			Cursor:    windows.Handle(cursor),
			ClassName: classW,
		}

		ret, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wcex)))
		if ret == 0 {
			initChan <- fmt.Errorf("RegisterClassExW failed: %w", err)
			return
		}

		x, y := getPosition(w, h, b.pos, margin)

		// ExStyle: WS_EX_TOPMOST | WS_EX_TOOLWINDOW | WS_EX_LAYERED | WS_EX_TRANSPARENT | WS_EX_NOACTIVATE
		exStyle := uintptr(0x00000008 | 0x00000080 | 0x00080000 | 0x00000020 | 0x08000000)
		// Style: WS_POPUP
		style := uintptr(0x80000000)

		titleW := windows.StringToUTF16Ptr("Just Talk Status Overlay")
		hwnd, _, err := procCreateWindowExW.Call(
			exStyle,
			uintptr(unsafe.Pointer(classW)),
			uintptr(unsafe.Pointer(titleW)),
			style,
			uintptr(x), uintptr(y),
			uintptr(w), uintptr(h),
			0, 0,
			uintptr(instance),
			0,
		)
		if hwnd == 0 {
			initChan <- fmt.Errorf("CreateWindowExW failed: %w", err)
			return
		}

		// Set color key: RGB(1, 1, 1) -> transparent (LWA_COLORKEY = 1)
		procSetLayeredWindowAttr.Call(hwnd, 0x010101, 255, 1)

		b.hwnd = windows.HWND(hwnd)
		activeBackend = b

		initChan <- nil

		var msg msg
		for {
			ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
			if int32(ret) <= 0 {
				break
			}
			procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
			procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
		}
	}()

	if err := <-initChan; err != nil {
		return nil, err
	}

	return b, nil
}

func translateLabel(label string) string {
	switch label {
	case "CON":
		return "连接中"
	case "REC":
		return "录音中"
	case "STP":
		return "延迟中"
	case "WAI":
		return "识别中"
	case "ERR":
		return "出错了"
	case "IDL":
		return "已就绪"
	default:
		return label
	}
}

func (b *windowsBackend) Show(label string, color statusColor) error {
	b.mu.Lock()
	b.label = translateLabel(label)
	b.color = color
	b.mu.Unlock()

	if b.hwnd == 0 {
		return nil
	}

	x, y := getPosition(b.w, b.h, b.pos, b.margin)
	// SWP_NOSIZE = 1, SWP_NOACTIVATE = 0x0010, SWP_NOZORDER = 4
	procSetWindowPos.Call(uintptr(b.hwnd), 0, uintptr(x), uintptr(y), 0, 0, 1|0x0010|4)

	procInvalidateRect.Call(uintptr(b.hwnd), 0, 1)

	// SW_SHOWNOACTIVATE = 4
	procShowWindow.Call(uintptr(b.hwnd), 4)
	return nil
}

func (b *windowsBackend) Hide() error {
	if b.hwnd == 0 {
		return nil
	}
	// SW_HIDE = 0
	procShowWindow.Call(uintptr(b.hwnd), 0)
	return nil
}

func (b *windowsBackend) Close() error {
	if b.hwnd == 0 {
		return nil
	}
	procPostMessageW.Call(uintptr(b.hwnd), 0x0010, 0, 0) // WM_CLOSE = 0x0010
	b.hwnd = 0
	activeBackend = nil
	return nil
}

func getPosition(w, h int, pos string, margin int) (x, y int) {
	var workArea rect
	// SPI_GETWORKAREA = 48
	ret, _, _ := procSystemParametersInfoW.Call(48, 0, uintptr(unsafe.Pointer(&workArea)), 0)

	var sw, sh int
	var left, top, right, bottom int

	if ret != 0 {
		left = int(workArea.Left)
		top = int(workArea.Top)
		right = int(workArea.Right)
		bottom = int(workArea.Bottom)
		sw = right - left
		sh = bottom - top
	} else {
		scrW, _, _ := procGetSystemMetrics.Call(0) // SM_CXSCREEN = 0
		scrH, _, _ := procGetSystemMetrics.Call(1) // SM_CYSCREEN = 1
		sw = int(scrW)
		sh = int(scrH)
		left = 0
		top = 0
		right = sw
		bottom = sh
	}

	x = right - w - margin
	y = top + margin

	switch strings.ToLower(pos) {
	case "top-left":
		x = left + margin
		y = top + margin
	case "top-center":
		x = left + (sw-w)/2
		y = top + margin
	case "top-right":
		x = right - w - margin
		y = top + margin
	case "bottom-left":
		x = left + margin
		y = bottom - h - margin
	case "bottom-center":
		x = left + (sw-w)/2
		y = bottom - h - margin
	case "bottom-right":
		x = right - w - margin
		y = bottom - h - margin
	}
	return x, y
}

func wndProc(hwnd uintptr, message uint32, wparam uintptr, lparam uintptr) uintptr {
	switch message {
	case 0x000F: // WM_PAINT
		var ps paintStruct
		hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		if hdc != 0 && activeBackend != nil {
			activeBackend.paint(hdc)
		}
		procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0

	case 0x0014: // WM_ERASEBKGND
		return 1

	case 0x0002: // WM_DESTROY
		procPostQuitMessage.Call(0)
		return 0
	}
	r1, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wparam, lparam)
	return r1
}

func (b *windowsBackend) paint(hdc uintptr) {
	b.mu.Lock()
	label := b.label
	color := b.color
	w := b.w
	h := b.h
	b.mu.Unlock()

	var r rect
	procGetClientRect.Call(uintptr(b.hwnd), uintptr(unsafe.Pointer(&r)))

	colorKeyBrush, _, _ := procCreateSolidBrush.Call(0x010101) // RGB(1,1,1) -> 0x010101
	defer procDeleteObject.Call(colorKeyBrush)

	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), colorKeyBrush)

	rVal := byte(color.R >> 8)
	gVal := byte(color.G >> 8)
	bVal := byte(color.B >> 8)
	pillColorRef := uint32(rVal) | (uint32(gVal) << 8) | (uint32(bVal) << 16)

	pillBrush, _, _ := procCreateSolidBrush.Call(uintptr(pillColorRef))
	defer procDeleteObject.Call(pillBrush)

	oldBrush, _, _ := procSelectObject.Call(hdc, pillBrush)
	defer procSelectObject.Call(hdc, oldBrush)

	nullPen, _, _ := procGetStockObject.Call(8) // NULL_PEN = 8
	oldPen, _, _ := procSelectObject.Call(hdc, nullPen)
	defer procSelectObject.Call(hdc, oldPen)

	corner := int32(h / 2)
	procRoundRect.Call(hdc, 0, 0, uintptr(w), uintptr(h), uintptr(corner), uintptr(corner))

	procSetTextColor.Call(hdc, 0xFFFFFF) // White
	procSetBkMode.Call(hdc, 1)           // TRANSPARENT = 1

	fontNameW := windows.StringToUTF16Ptr("Segoe UI")
	fontH := -int32(14.0 * float64(h) / 42.0)
	if fontH > -10 {
		fontH = -12
	}
	font, _, _ := procCreateFontW.Call(
		uintptr(fontH), 0, 0, 0,
		700, // FW_BOLD = 700
		0, 0, 0, 1, 0, 0, 2, 0,
		uintptr(unsafe.Pointer(fontNameW)),
	)
	if font != 0 {
		oldFont, _, _ := procSelectObject.Call(hdc, font)
		defer func() {
			procSelectObject.Call(hdc, oldFont)
			procDeleteObject.Call(font)
		}()
	}

	labelW := windows.StringToUTF16Ptr(label)
	minusOne := int32(-1)
	procDrawTextW.Call(hdc, uintptr(unsafe.Pointer(labelW)), uintptr(minusOne), uintptr(unsafe.Pointer(&r)), 1|4|32) // DT_CENTER | DT_VCENTER | DT_SINGLELINE
}
