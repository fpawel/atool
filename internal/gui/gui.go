package gui

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/pkg/logfile"
	"github.com/fpawel/atool/internal/pkg/winapi"
	"github.com/fpawel/atool/internal/pkg/winapi/copydata"
	"github.com/lxn/win"
	"os"
	"sync/atomic"
	"time"
)

type MsgCopyData = uintptr

const (
	MsgNewCommTransaction MsgCopyData = iota
	MsgNewProductParamValue
	MsgChart
	MsgStatus
	MsgCoefficient
	MsgProductConnection
	MsgDelay
	MsgWorkSuspended
	MsgLuaSelectWorks
	MsgGas
	MsgTemperature
	MsgTemperatureSetPoint
	MsgProgress
	MsgJournal
)

const (
	wmuCurrentPartyChanged = win.WM_USER + 1 + iota
	wmuStartWork
	wmuStopWork
	wmuRequestConfigParamValues
)

func NotifyProgressShow(max int, what string) bool {
	return copyData().SendJson(MsgProgress, ProgressInfo{Cmd: ProgressShow, Max: max, What: what})
}

func NotifyProgressHide() bool {
	return copyData().SendJson(MsgProgress, ProgressInfo{Cmd: ProgressHide})
}

func NotifyProgress(value int, what string) bool {
	return copyData().SendJson(MsgProgress, ProgressInfo{Cmd: ProgressProgress, Progress: value, What: what})
}

func RequestConfigParamValues() {
	sendMessage(wmuRequestConfigParamValues, 0, 0)
}

func NotifyLuaSelectWorks(works []string) bool {
	return copyData().SendJson(MsgLuaSelectWorks, works)
}

func NotifyWorkSuspended(Text string) bool {
	return copyData().SendString(MsgWorkSuspended, Text)
}

func NotifyStatus(x Status) bool {
	return copyData().SendJson(MsgStatus, x)
}

func NotifyGas(gas int) bool {
	return copyData().SendString(MsgGas, fmt.Sprintf("ПГС: %d", gas))
}

func NotifyTemperature(t float64) bool {
	return copyData().SendString(MsgTemperature, fmt.Sprintf("температура: %v⁰C", t))
}

func NotifyTemperatureSetPoint(t float64) bool {
	return copyData().SendString(MsgTemperatureSetPoint, fmt.Sprintf("уставка: %v⁰C", t))
}

func Popup(x string) bool {
	return NotifyStatus(Status{Text: x, Ok: true, PopupLevel: LPopup})
}

func Popupf(format string, a ...interface{}) bool {
	return Popup(fmt.Sprintf(format, a...))
}

func NotifyBeginDelay(duration time.Duration, what string) bool {
	return copyData().SendJson(MsgDelay, struct {
		Delay          bool
		DurationMillis int64
		What           string
	}{true, int64(duration / time.Millisecond), what})
}

func NotifyEndDelay() bool {
	return copyData().SendJson(MsgDelay, struct {
		Delay bool
	}{false})
}

func NotifyCoefficient(xs CoefficientValue) bool {
	return copyData().SendJson(MsgCoefficient, xs)
}

func NotifyProductConnection(x ProductConnection) bool {
	return copyData().SendJson(MsgProductConnection, x)
}

func NotifyNewCommTransaction(c CommTransaction) bool {
	return copyData().SendJson(MsgNewCommTransaction, c)
}

func NotifyNewProductParamValue(x ProductParamValue) bool {
	return copyData().SendJson(MsgNewProductParamValue, x)
}

func NotifyChart(xs []data.Measurement) bool {

	for n := 0; n < len(xs); {
		p := xs[n:]
		offset := len(p)
		if offset > 100000 {
			offset = 100000
		}
		p = p[:offset]
		n += offset

		buf := bytes.NewBuffer(make([]byte, 0, 3300000))
		writeBinary(buf, int64(len(p)))
		for _, m := range p {
			writeBinary(buf, m.Time().UnixNano()/1000000) // количество миллисекунд метки времени
			writeBinary(buf, m.ProductID)
			writeBinary(buf, uint64(m.ParamAddr))
			writeBinary(buf, m.Value)
		}
		if !copyData().SendBytes(MsgChart, buf.Bytes()) {
			return false
		}
	}

	buf := bytes.NewBuffer(nil)
	writeBinary(buf, uint64(0))
	if !copyData().SendBytes(MsgChart, buf.Bytes()) {
		return false
	}

	return true
}

func NotifyJournal(xs []logfile.JournalRecord) bool {

	for n := 0; n < len(xs); {
		p := xs[n:]
		offset := len(p)
		if offset > 100000 {
			offset = 100000
		}
		p = p[:offset]
		n += offset

		buf := bytes.NewBuffer(make([]byte, 0, 3300000))
		writeBinary(buf, int64(len(p)))
		for _, m := range p {
			writeBinary(buf, m.Time.UnixNano()/1000000) // количество миллисекунд метки времени
			writeBinary(buf, byte(m.Level))
			writeBinary(buf, m.Ok)
			writeBinary(buf, uint64(len(m.Text)))
			writeBinary(buf, []byte(m.Text))
		}
		if !copyData().SendBytes(MsgJournal, buf.Bytes()) {
			return false
		}
	}

	buf := bytes.NewBuffer(nil)
	writeBinary(buf, uint64(0))
	if !copyData().SendBytes(MsgJournal, buf.Bytes()) {
		return false
	}

	return true
}

//func MsgBox(title, message string, style int) int {
//	hWnd := hWndTargetSendMessage()
//	if hWnd == win.HWND_TOP {
//		return 0
//	}
//	return int(win.MessageBox(
//		hWnd,
//		winapi.MustUTF16PtrFromString(strings.ReplaceAll(message, "\x00", "␀")),
//		winapi.MustUTF16PtrFromString(strings.ReplaceAll(title, "\x00", "␀")),
//		uint32(style)))
//}

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
	hWndSourceSendMessage = winapi.NewWindowWithClassName(os.Args[0] + "33BCE8B6-E14D-4060-97C9-2B7E79719195")

	hWndTargetSendMessage,
	setHWndTargetSendMessage = func() (func() win.HWND, func(win.HWND)) {
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
