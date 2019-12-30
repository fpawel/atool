package pkg

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/powerman/structlog"
	"path/filepath"
	"runtime"
)

// PrintMerryStacktrace returns the error's stacktrace as a string formatted
// the same way as golangs runtime package.
// If e has no stacktrace, returns an empty string.
func PrintMerryStacktrace(log *structlog.Logger, e error) {
	for i, fp := range merry.Stack(e) {
		fnc := runtime.FuncForPC(fp)
		if fnc != nil {
			f, l := fnc.FileLine(fp)
			name := filepath.Base(fnc.Name())
			ident := " "
			if i > 0 {
				ident = "\t"
			}
			log.PrintErr(fmt.Sprintf("%s%s:%d %s", ident, f, l, name))
		}
	}
}
