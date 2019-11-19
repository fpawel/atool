package gui

import (
	"fmt"
	"github.com/fpawel/atool/internal/pkg/winapi"
	"github.com/lxn/win"
	"strings"
	"sync/atomic"
)

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
	sendMessage(cmdCurrentPartyChanged, 0, 0)
}

const (
	cmdCurrentPartyChanged = win.WM_USER + 1 + iota
)

func sendMessage(msg uint32, wParam uintptr, lParam uintptr) uintptr {
	return win.SendMessage(hWndTargetSendMessage(), msg, wParam, lParam)
}

var (
	hWndTargetSendMessage, setHWndTargetSendMessage = func() (func() win.HWND, func(win.HWND)) {
		hWnd := int64(win.HWND_TOP)
		return func() win.HWND {
				return win.HWND(atomic.LoadInt64(&hWnd))
			}, func(x win.HWND) {
				atomic.StoreInt64(&hWnd, int64(x))
			}
	}()
	//hWndSourceCopyData = winapi.NewWindowWithClassName(os.Args[0]+strconv.FormatInt(time.Now().Unix(), 10))
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
