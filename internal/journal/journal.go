package journal

import (
	"fmt"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/logfile"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/powerman/structlog"
	"time"
)

func Close() error {
	return file.Close()
}

func Info(log *structlog.Logger, x string) {
	status(log, gui.Status{Text: x, Ok: true, PopupLevel: gui.LJournal})
}

func Err(log *structlog.Logger, err error) {
	status(log, gui.Status{Text: err.Error(), Ok: false, PopupLevel: gui.LJournal})
}

func Warn(log *structlog.Logger, x string) {
	status(log, gui.Status{Text: x, Ok: true, PopupLevel: gui.LWarn})
}

func WarnError(log *structlog.Logger, err error) {
	status(log, gui.Status{Text: err.Error(), Ok: false, PopupLevel: gui.LWarn})
}

func status(log *structlog.Logger, x gui.Status) {
	strTime := time.Now().Format("15:04:05")

	log = pkg.LogPrependSuffixKeys(log, "popup_level", x.PopupLevel, structlog.KeyTime, strTime)
	if x.Ok {
		log.Info(x.Text)
	} else {
		log.PrintErr(x.Text)
	}
	var err error
	if x.Ok {
		_, err = fmt.Fprintf(file, "%s %s\n", strTime, x.Text)
	} else {
		_, err = fmt.Fprintf(file, "%s ERR %s\n", strTime, x.Text)
	}
	must.PanicIf(err)

	go gui.NotifyStatus(x)
}

var file = logfile.MustNew(".journal")
