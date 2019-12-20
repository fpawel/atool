package gui

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/pkg/winapi"
	"github.com/fpawel/atool/internal/pkg/winapi/copydata"
	"github.com/lxn/win"
	"os"
	"strings"
	"sync/atomic"
)

type MsgCopyData = uintptr

const (
	MsgNewCommTransaction MsgCopyData = iota
	MsgNewProductParamValue
	MsgChart
	MsgPopup
	MsgCoefficients
)

func NotifyCoefficients(xs []CoefficientValue) bool {
	return copyData().SendJson(MsgCoefficients, xs)
}

func Popup(warning bool, x string) bool {
	return copyData().SendJson(MsgPopup, PopupMessage{Text: x, Ok: true, Warning: warning})
}

func PopupError(warning bool, x error) bool {
	return copyData().SendJson(MsgPopup, PopupMessage{Text: x.Error(), Ok: false, Warning: warning})
}

func NotifyNewCommTransaction(с CommTransaction) bool {
	return copyData().SendJson(MsgNewCommTransaction, с)
}

func NotifyNewProductParamValue(x ProductParamValue) bool {
	return copyData().SendJson(MsgNewProductParamValue, x)
}

func NotifyChart(xs []data.Measurement) bool {
	buf := new(bytes.Buffer)
	writeBinary(buf, int64(len(xs)))
	for _, m := range xs {
		writeBinary(buf, m.Time.UnixNano()/1000000) // количество миллисекунд метки времени
		writeBinary(buf, m.ProductID)
		writeBinary(buf, uint64(m.ParamAddr))
		writeBinary(buf, m.Value)
	}
	return copyData().SendMessage(MsgChart, buf.Bytes())
}

func MsgBox(title, message string, style int) int {
	hWnd := hWndTargetSendMessage()
	if hWnd == win.HWND_TOP {
		return 0
	}
	return int(win.MessageBox(
		hWnd,
		winapi.MustUTF16PtrFromString(strings.ReplaceAll(message, "\x00", "␀")),
		winapi.MustUTF16PtrFromString(strings.ReplaceAll(title, "\x00", "␀")),
		uint32(style)))
}

func SetHWndTargetSendMessage(hWnd win.HWND) {
	setHWndTargetSendMessage(hWnd)
}

func NotifyCurrentPartyChanged() {
	sendMessage(wmuCurrentPartyChanged, 0, 0)
}

func NotifyStartWork() {
	sendMessage(wmuStartWork, 0, 0)
}

func NotifyStopWork() {
	sendMessage(wmuStopWork, 0, 0)
}

const (
	wmuCurrentPartyChanged = win.WM_USER + 1 + iota
	wmuStartWork
	wmuStopWork
)

func sendMessage(msg uint32, wParam uintptr, lParam uintptr) uintptr {
	return win.SendMessage(hWndTargetSendMessage(), msg, wParam, lParam)
}

func writeBinary(buf *bytes.Buffer, data interface{}) {
	if err := binary.Write(buf, binary.LittleEndian, data); err != nil {
		panic(err)
	}
}

func copyData() copydata.W {
	return copydata.New(hWndSourceSendMessage, hWndTargetSendMessage())
}

var (
	hWndSourceSendMessage                           = winapi.NewWindowWithClassName(os.Args[0] + "33BCE8B6-E14D-4060-97C9-2B7E79719195")
	hWndTargetSendMessage, setHWndTargetSendMessage = func() (func() win.HWND, func(win.HWND)) {
		hWnd := int64(win.HWND_TOP)
		return func() win.HWND {
				return win.HWND(atomic.LoadInt64(&hWnd))
			}, func(x win.HWND) {
				atomic.StoreInt64(&hWnd, int64(x))
			}
	}()
)

func init() {

	go func() {
		for {
			var msg win.MSG
			if win.GetMessage(&msg, 0, 0, 0) == 0 {
				fmt.Println("выход из цикла оконных сообщений")
				return
			}
			fmt.Printf("%+v\n", msg)
			win.TranslateMessage(&msg)
			win.DispatchMessage(&msg)
		}
	}()
}