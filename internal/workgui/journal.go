package workgui

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/logfile"
	"github.com/powerman/structlog"
	"strings"
	"time"
)

func NotifyWorkSuspended(err error) {
	err = merry.Prepend(err, "произошла ошибка: выполнение приостановлено")
	File.WriteError(err)
	go gui.NotifyWorkSuspended(indentStr() + err.Error())
}

func WithNotifyResult(log *structlog.Logger, what string, work func() error) error {
	if err := work(); err != nil {
		NotifyErr(log, merry.Prepend(err, what))
		return err
	}
	NotifyInfo(log, what+" - успешно")
	return nil
}

func WithNotifyValue(log *structlog.Logger, what string, work func() (float64, error)) (float64, error) {
	value, err := work()
	if err != nil {
		NotifyErr(log, merry.Prepend(err, what))
		return 0, err
	}
	NotifyInfo(log, fmt.Sprintf("%s = %v", what, value))
	return value, nil
}

func NotifyInfo(log *structlog.Logger, x string) {
	notifyStatus(log, gui.Status{Text: x, Ok: true, PopupLevel: gui.LJournal})
	File.Write(x)
}

func NotifyErr(log *structlog.Logger, err error) {
	notifyStatus(log, gui.Status{Text: "⚠️ " + err.Error(), Ok: false, PopupLevel: gui.LJournal})
	File.WriteError(err)
}

func NotifyWarn(log *structlog.Logger, warn string) {
	notifyStatus(log, gui.Status{Text: "⚠️ " + warn, Ok: false, PopupLevel: gui.LJournal})
	File.WriteErrorText(warn)
}

func NotifyWarnError(log *structlog.Logger, err error) {
	notifyStatus(log, gui.Status{Text: "⚠️ " + err.Error(), Ok: false, PopupLevel: gui.LWarn})
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
