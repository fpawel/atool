package guiwork

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/logfile"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/powerman/structlog"
	"strings"
	"time"
)

func CloseJournal() error {
	return fileJournal.Close()
}

func NotifyLuaSuspended(err error) {
	err = merry.Prepend(err, "произошла ошибка: выполнение приостановлено")
	writeFileJournalError(err)
	go gui.NotifyLuaSuspended(indentWorkLevel() + err.Error())
}

func JournalInfo(log *structlog.Logger, x string) {
	notifyStatus(log, gui.Status{Text: x, Ok: true, PopupLevel: gui.LJournal})
	writeFileJournal(true, x)
}

func JournalErr(log *structlog.Logger, err error) {
	notifyStatus(log, gui.Status{Text: err.Error(), Ok: false, PopupLevel: gui.LJournal})
	writeFileJournalError(err)
}

//func Warn(log *structlog.Logger, x string) {
//	status(log, gui.Status{Text: x, Ok: true, PopupLevel: gui.LWarn})
//}

func JournalWarnError(log *structlog.Logger, err error) {
	notifyStatus(log, gui.Status{Text: err.Error(), Ok: false, PopupLevel: gui.LWarn})
	writeFileJournalError(err)
}

func writeFileJournalError(err error) {
	writeFileJournal(false, err.Error()+"\n\t"+pkg.FormatMerryStacktrace(err, "\n\t"))
}

func writeFileJournal(ok bool, text string) {
	strTime := time.Now().Format("15:04:05")

	var err error
	indent := indentWorkLevel()
	currentWorkLevel := currentWorkLevel()
	s1 := fmt.Sprintf("%s [%d]%s", strTime, currentWorkLevel, indent)
	if ok {
		_, err = fmt.Fprintf(fileJournal, "%s%s\n", s1, text)
		must.PanicIf(err)
		return
	}
	_, err = fmt.Fprintf(fileJournal, "%sERR %s\n", s1, text)
	must.PanicIf(err)
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
	x.Text = indentWorkLevel() + x.Text
	go gui.NotifyStatus(x)
}

func indentWorkLevel() string {
	currentWorkLevel := currentWorkLevel()
	if currentWorkLevel > 1 {
		return strings.Repeat("    ", currentWorkLevel-1)
	}
	return ""
}

var (
	fileJournal = logfile.MustNew(".journal")
)
