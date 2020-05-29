package guiwork

import (
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/logfile"
	"github.com/powerman/structlog"
	"strings"
	"time"
)

func NotifyLuaSuspended(err error) {
	err = merry.Prepend(err, "произошла ошибка: выполнение приостановлено")
	File.WriteError(err)
	go gui.NotifyLuaSuspended(indentStr() + err.Error())
}

func NotifyInfo(log *structlog.Logger, x string) {
	notifyStatus(log, gui.Status{Text: x, Ok: true, PopupLevel: gui.LJournal})
	File.Write(x)
}

func NotifyErr(log *structlog.Logger, err error) {
	notifyStatus(log, gui.Status{Text: err.Error(), Ok: false, PopupLevel: gui.LJournal})
	File.WriteError(err)
}

func NotifyWarnError(log *structlog.Logger, err error) {
	notifyStatus(log, gui.Status{Text: err.Error(), Ok: false, PopupLevel: gui.LWarn})
	File.WriteError(err)
}

func NotifyJournal() {
	_ = File.Close()
	journalRecords := logfile.ReadJournal()
	File = logfile.NewJournal(currentWorkLevel)
	go gui.NotifyJournal(journalRecords)
}

func notifyStatus(log *structlog.Logger, x gui.Status) {
	log = pkg.LogPrependSuffixKeys(log,
		structlog.KeyTime, time.Now().Format("15:04:05"),
		"popup_level", x.PopupLevel,
	)
	if x.Ok {
		log.Info(x.Text)
	} else {
		log.PrintErr(x.Text)
	}
	x.Text = indentStr() + x.Text
	go gui.NotifyStatus(x)
}

func indentStr() string {
	return strings.Repeat("    ", currentWorkLevel())
}

var (
	File = logfile.NewJournal(currentWorkLevel)
)
