package workgui

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/logjrn"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/powerman/structlog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func OpenJournal() error {
	return Journal.Open(filepath.Join(filepath.Dir(os.Args[0]), "logjrn.sqlite"))
}

func CloseJournal() error {
	wgJrn.Wait()
	return Journal.Close()
}

func NotifyWorkSuspended(err error) {
	err = merry.Prepend(err, "произошла ошибка: выполнение приостановлено")
	go newEntryError(err).save()
	go gui.NotifyWorkSuspended(indentStr() + err.Error())
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
	newEntry(x).save()
	go notifyStatus(log, gui.Status{Text: indentStr() + x, Ok: true, PopupLevel: gui.LJournal})
}

func NotifyErr(log *structlog.Logger, err error) {
	newEntryError(err).save()
	go notifyStatus(log, gui.Status{Text: indentStr() + "⚠️ " + err.Error(), Ok: false, PopupLevel: gui.LJournal})
}

func NotifyWarn(log *structlog.Logger, warn string) {
	newEntryErrorText(warn).save()
	go notifyStatus(log, gui.Status{Text: indentStr() + "⚠️ " + warn, Ok: false, PopupLevel: gui.LJournal})
}

func NotifyWarnError(log *structlog.Logger, err error) {
	newEntryError(err).save()
	go notifyStatus(log, gui.Status{Text: indentStr() + "⚠️ " + err.Error(), Ok: false, PopupLevel: gui.LWarn})
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
	gui.NotifyStatus(x)
}

func indentStr() string {
	return strings.Repeat("    ", currentWorkLevel())
}

type jrnEntry struct {
	*logjrn.Entry
}

func (e jrnEntry) save() {
	wgJrn.Add(1)
	defer wgJrn.Done()
	must.PanicIf(Journal.AddEntry(e.Entry))
}

func newEntry(text string) jrnEntry {
	return jrnEntry{
		&logjrn.Entry{
			StoredAt: time.Now(),
			Text:     text,
			Ok:       true,
			Indent:   currentWorkLevel(),
		},
	}
}

func newEntryErrorText(text string) jrnEntry {
	return jrnEntry{
		&logjrn.Entry{
			StoredAt: time.Now(),
			Text:     text,
			Ok:       false,
			Indent:   currentWorkLevel(),
		},
	}
}

func newEntryError(err error) jrnEntry {
	return jrnEntry{
		&logjrn.Entry{
			StoredAt: time.Now(),
			Text:     err.Error(),
			Ok:       false,
			Indent:   currentWorkLevel(),
			Stack:    pkg.FormatStacktrace(merry.Stack(err), "\n\t"),
		},
	}
}

var (
	Journal = new(logjrn.J)
	wgJrn   sync.WaitGroup
)
