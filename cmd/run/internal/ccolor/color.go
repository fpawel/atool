// +build windows

package ccolor

import (
	"syscall"
	"unsafe"
)

// Color is the type of color to be set.
type Color int

const (
	// No change of color
	None = Color(iota)
	Black
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

// ResetColor resets the foreground and background to original colors
func ResetColor() {
	resetColor()
}

// ChangeColor sets the foreground and background colors. If the value of the color is None,
// the corresponding color keeps unchanged.
// If fgBright or bgBright is set true, corresponding color use bright color. bgBright may be
// ignored in some OS environment.
func ChangeColor(fg Color, fgBright bool, bg Color, bgBright bool) {
	changeColor(fg, fgBright, bg, bgBright)
}

// Foreground changes the foreground color.
func Foreground(cl Color, bright bool) {
	ChangeColor(cl, bright, None, false)
}

// Background changes the background color.
func Background(cl Color, bright bool) {
	ChangeColor(None, false, cl, bright)
}

var fgColors = []uint16{
	0,
	0,
	fgRed,
	fgGreen,
	fgRed | fgGreen,
	fgBlue,
	fgRed | fgBlue,
	fgGreen | fgBlue,
	fgRed | fgGreen | fgBlue}

var bgColors = []uint16{
	0,
	0,
	bgRed,
	bgGreen,
	bgRed | bgGreen,
	bgBlue,
	bgRed | bgBlue,
	bgGreen | bgBlue,
	bgRed | bgGreen | bgBlue}

const (
	fgBlue      = uint16(0x0001)
	fgGreen     = uint16(0x0002)
	fgRed       = uint16(0x0004)
	fgIntensity = uint16(0x0008)
	bgBlue      = uint16(0x0010)
	bgGreen     = uint16(0x0020)
	bgRed       = uint16(0x0040)
	bgIntensity = uint16(0x0080)

	fgMask = fgBlue | fgGreen | fgRed | fgIntensity
	bgMask = bgBlue | bgGreen | bgRed | bgIntensity
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procGetStdHandle               = kernel32.NewProc("GetStdHandle")
	procSetConsoleTextAttribute    = kernel32.NewProc("SetConsoleTextAttribute")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")

	hStdout        uintptr
	initScreenInfo *consoleScreenBufferInfo
)

func setConsoleTextAttribute(hConsoleOutput uintptr, wAttributes uint16) bool {
	ret, _, _ := procSetConsoleTextAttribute.Call(
		hConsoleOutput,
		uintptr(wAttributes))
	return ret != 0
}

type coordinate struct {
	X, Y int16
}

type smallRect struct {
	Left, Top, Right, Bottom int16
}

type consoleScreenBufferInfo struct {
	DwSize              coordinate
	DwCursorPosition    coordinate
	WAttributes         uint16
	SrWindow            smallRect
	DwMaximumWindowSize coordinate
}

func getConsoleScreenBufferInfo(hConsoleOutput uintptr) *consoleScreenBufferInfo {
	var csbi consoleScreenBufferInfo
	if ret, _, _ := procGetConsoleScreenBufferInfo.Call(hConsoleOutput, uintptr(unsafe.Pointer(&csbi))); ret == 0 {
		return nil
	}
	return &csbi
}

const (
	stdOutputHandle = uint32(-11 & 0xFFFFFFFF)
)

func init() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")

	procGetStdHandle = kernel32.NewProc("GetStdHandle")

	hStdout, _, _ = procGetStdHandle.Call(uintptr(stdOutputHandle))

	initScreenInfo = getConsoleScreenBufferInfo(hStdout)
}

func resetColor() {
	if initScreenInfo == nil { // No console info - Ex: stdout redirection
		return
	}
	setConsoleTextAttribute(hStdout, initScreenInfo.WAttributes)
}

func changeColor(fg Color, fgBright bool, bg Color, bgBright bool) {
	attr := uint16(0)
	if fg == None || bg == None {
		cBufInfo := getConsoleScreenBufferInfo(hStdout)
		if cBufInfo == nil { // No console info - Ex: stdout redirection
			return
		}
		attr = cBufInfo.WAttributes
	}
	if fg != None {
		attr = attr & ^fgMask | fgColors[fg]
		if fgBright {
			attr |= fgIntensity
		}
	}
	if bg != None {
		attr = attr & ^bgMask | bgColors[bg]
		if bgBright {
			attr |= bgIntensity
		}
	}
	setConsoleTextAttribute(hStdout, attr)
}
