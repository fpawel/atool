package gui

import (
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/lxn/win"
	"github.com/powerman/structlog"
	"time"
)

func notifyStatus(log *structlog.Logger, x Status) {
	if log != nil && x.PopupLevel > 0 {
		log = pkg.LogPrependSuffixKeys(log, "popup_level", x.PopupLevel, structlog.KeyTime, time.Now().Format("15:04:05"))
		if x.Ok {
			log.Info(x.Text)
		} else {
			log.PrintErr(x.Text)
		}
	}
	go func() {
		w := copyData()
		if !w.SendJson(MsgStatus, x) {
			if log == nil {
				log = structlog.New()
			}
			err := merry.Errorf("can't send message %d: %+v: %+v", MsgStatus, x, w)
			if errno := win.GetLastError(); errno != win.ERROR_SUCCESS {
				err = err.Appendf("windows error code: %d", errno)
			}
			log.PrintErr(err)
		}
	}()
}

func Popup(log *structlog.Logger, x string) {
	notifyStatus(log, Status{Text: x, Ok: true, PopupLevel: LPopup})
}

func Journal(log *structlog.Logger, x string) {
	notifyStatus(log, Status{Text: x, Ok: true, PopupLevel: LJournal})
}

func Warn(log *structlog.Logger, x string) {
	notifyStatus(log, Status{Text: x, Ok: true, PopupLevel: LWarn})
}

func JournalError(log *structlog.Logger, err error) {
	notifyStatus(log, Status{Text: err.Error(), Ok: false, PopupLevel: LJournal})
}

func WarnError(log *structlog.Logger, err error) {
	notifyStatus(log, Status{Text: err.Error(), Ok: false, PopupLevel: LWarn})
}
