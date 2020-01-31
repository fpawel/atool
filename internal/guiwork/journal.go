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

func NotifyLuaSuspended(log *structlog.Logger, err error) {
	err = merry.New("произошла ошибка: выполнение приостановлено").WithCause(err)
	writeFileJournal(log, false, err.Error())
	go gui.NotifyLuaSuspended(indentWorkLevel() + err.Error())
}

func JournalInfo(log *structlog.Logger, x string) {
	notifyStatus(log, gui.Status{Text: x, Ok: true, PopupLevel: gui.LJournal})
}

func JournalErr(log *structlog.Logger, err error) {
	notifyStatus(log, gui.Status{Text: err.Error(), Ok: false, PopupLevel: gui.LJournal})
}

//func Warn(log *structlog.Logger, x string) {
//	status(log, gui.Status{Text: x, Ok: true, PopupLevel: gui.LWarn})
//}

func JournalWarnError(log *structlog.Logger, err error) {
	notifyStatus(log, gui.Status{Text: err.Error(), Ok: false, PopupLevel: gui.LWarn})
}

func writeFileJournal(log *structlog.Logger, ok bool, text string) {
	strTime := time.Now().Format("15:04:05")
	log = pkg.LogPrependSuffixKeys(log, structlog.KeyTime, strTime)
	var err error
	indent := indentWorkLevel()
	currentWorkLevel := currentWorkLevel()

	s1 := fmt.Sprintf("%s [%d]%s", strTime, currentWorkLevel, indent)

	if ok {
		log.Info(text)
		_, err = fmt.Fprintf(fileJournal, "%s%s\n", s1, text)
	} else {
		log.PrintErr(text)
		_, err = fmt.Fprintf(fileJournal, "%sERR %s\n", s1, text)
	}
	must.PanicIf(err)
}

func notifyStatus(log *structlog.Logger, x gui.Status) {
	log = pkg.LogPrependSuffixKeys(log, "popup_level", x.PopupLevel)
	writeFileJournal(log, x.Ok, x.Text)

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
