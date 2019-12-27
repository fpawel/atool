package pkg

import (
	"bytes"
	"fmt"
	"github.com/ansel1/merry"
	"path/filepath"
	"runtime"
)

// FormatMerryStacktrace returns the error's stacktrace as a string formatted
// the same way as golangs runtime package.
// If e has no stacktrace, returns an empty string.
func FormatMerryStacktrace(e error) string {
	s := merry.Stack(e)
	if len(s) == 0 {
		return ""
	}
	buf := bytes.Buffer{}
	for i, fp := range s {
		fnc := runtime.FuncForPC(fp)
		if fnc != nil {
			f, l := fnc.FileLine(fp)
			name := filepath.Base(fnc.Name())
			ident := " "
			if i > 0 {
				ident = "\t"
			}
			buf.WriteString(fmt.Sprintf("%s%s:%d %s\n", ident, f, l, name))
		}
	}
	return buf.String()

}
