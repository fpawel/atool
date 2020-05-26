package logfile

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/must"
	"os"
	"strings"
	"time"
)

func MustNewJournal(filenameSuffix string, lev func() int) Journal {
	f, err := New(filenameSuffix)
	must.PanicIf(err)
	return Journal{
		f:   f,
		lev: lev,
	}
}

type Journal struct {
	f   *os.File
	lev func() int
}

func (x Journal) Close() error {
	return x.f.Close()
}

func (x Journal) WriteError(err error) {
	x.WriteErr(err.Error() + "\n\t" + pkg.FormatStacktrace(merry.Stack(err), "\n\t"))
}

func (x Journal) WriteErr(text string) {
	x.Write("ERR " + text)
}

func (x Journal) Write(text string) {
	_, err := fmt.Fprintf(x.f, "%s%s\n", prefix(x.lev()), text)
	must.PanicIf(err)
}

func prefix(lev int) string {
	return fmt.Sprintf("%s [%d]%s", time.Now().Format("15:04:05"), lev, strings.Repeat("    ", lev))
}
