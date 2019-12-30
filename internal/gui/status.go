package gui

import (
	"github.com/fpawel/atool/internal/pkg"
	"github.com/powerman/structlog"
)

func status(log *structlog.Logger, x Status) {
	if x.PopupLevel > 0 {
		log = pkg.LogPrependSuffixKeys(log, "popup_level", x.PopupLevel)
		if x.Ok {
			log.Info(x.Text)
		} else {
			log.PrintErr(x.Text)
		}
	}
	go func() {
		if !copyData().SendJson(MsgStatus, x) {
			log.PrintErr("can't send message %d: %+v", MsgStatus, x)
		}
	}()
}

func Popup(log *structlog.Logger, x string) {
	status(log, Status{Text: x, Ok: true, PopupLevel: LPopup})
}

func Journal(log *structlog.Logger, x string) {
	status(log, Status{Text: x, Ok: true, PopupLevel: LJournal})
}

func Warn(log *structlog.Logger, x string) {
	status(log, Status{Text: x, Ok: true, PopupLevel: LWarn})
}

func JournalError(log *structlog.Logger, err error) {
	status(log, Status{Text: err.Error(), Ok: false, PopupLevel: LJournal})
}

func WarnError(log *structlog.Logger, err error) {
	status(log, Status{Text: err.Error(), Ok: false, PopupLevel: LWarn})
}
