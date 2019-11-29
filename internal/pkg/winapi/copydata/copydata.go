package copydata

import (
	"encoding/json"
	"fmt"
	"github.com/lxn/win"
	"reflect"
	"unicode/utf16"
	"unsafe"
)

func SendMessage(hWndSrc, hWndDest win.HWND, wParam uintptr, b []byte) uintptr {
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&b))
	cd := copyData{
		CbData: uint32(header.Len),
		LpData: header.Data,
		DwData: uintptr(hWndSrc),
	}
	return win.SendMessage(hWndDest, win.WM_COPYDATA, wParam, uintptr(unsafe.Pointer(&cd)))
}

type W struct {
	HWndSrc, HWndDest win.HWND
}

func New(HWndSrc, HWndDest win.HWND) W {
	return W{
		HWndSrc:  HWndSrc,
		HWndDest: HWndDest,
	}
}

func (x W) SendString(msg uintptr, s string) bool {
	return x.SendMessage(msg, utf16FromString(s))
}

func (x W) SendJson(msg uintptr, param interface{}) bool {
	b, err := json.Marshal(param)
	if err != nil {
		panic(err)
	}
	return x.SendString(msg, string(b))
}

func (x W) SendMessage(msg uintptr, b []byte) bool {
	return SendMessage(x.HWndSrc, x.HWndDest, msg, b) != 0
}

type copyData struct {
	DwData uintptr
	CbData uint32
	LpData uintptr
}

func utf16FromString(s string) (b []byte) {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			panic(fmt.Sprintf("%q[%d] is 0", s, i))
		}
	}
	for _, v := range utf16.Encode([]rune(s)) {
		b = append(b, byte(v), byte(v>>8))
	}
	return
}
