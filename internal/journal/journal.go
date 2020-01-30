package journal

import (
	"fmt"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/logfile"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/powerman/structlog"
	"strings"
	"time"
)

func Close() error {
	return file.Close()
}

func ScriptSuspended(log *structlog.Logger, err error) {
	writeFile(log, false, err.Error())
	go gui.NotifyLuaSuspended(err.Error())
}

func Info(log *structlog.Logger, x string) {
	status(log, gui.Status{Text: x, Ok: true, PopupLevel: gui.LJournal})
}

func Err(log *structlog.Logger, err error) {
	status(log, gui.Status{Text: err.Error(), Ok: false, PopupLevel: gui.LJournal})
}

//func Warn(log *structlog.Logger, x string) {
//	status(log, gui.Status{Text: x, Ok: true, PopupLevel: gui.LWarn})
//}

func WarnError(log *structlog.Logger, err error) {
	status(log, gui.Status{Text: err.Error(), Ok: false, PopupLevel: gui.LWarn})
}

func writeFile(log *structlog.Logger, ok bool, text string) {
	strTime := time.Now().Format("15:04:05")
	log = pkg.LogPrependSuffixKeys(log, structlog.KeyTime, strTime)
	var err error
	var indent string
	if CurrentWorkLevel > 1 {
		indent = strings.Repeat("\t", CurrentWorkLevel-1)
	}
	s1 := fmt.Sprintf("%s [%d]%s", strTime, CurrentWorkLevel, indent)

	if ok {
		log.Info(text)
		_, err = fmt.Fprintf(file, "%s%s\n", s1, text)
	} else {
		log.PrintErr(text)
		_, err = fmt.Fprintf(file, "%sERR %s\n", s1, text)
	}
	must.PanicIf(err)
}

func status(log *structlog.Logger, x gui.Status) {
	log = pkg.LogPrependSuffixKeys(log, "popup_level", x.PopupLevel)
	writeFile(log, x.Ok, x.Text)
	if CurrentWorkLevel > 1 {
		x.Text = strings.Repeat("    ", CurrentWorkLevel-1) + x.Text
	}
	go gui.NotifyStatus(x)
}

var (
	file             = logfile.MustNew(".journal")
	CurrentWorkLevel int
)
