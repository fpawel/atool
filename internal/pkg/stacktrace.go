package pkg

import (
	"fmt"
	"github.com/ansel1/merry"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

func FormatStacktrace(skip int, sep string) string {
	return formatStack(captureStack(skip), sep)
}

func FormatMerryStacktrace(e error, sep string) string {
	return formatStack(merry.Stack(e), sep)
}

func formatStack(stack []uintptr, sep string) string {
	trace := ""
	for i, fp := range stack {
		fnc := runtime.FuncForPC(fp)
		if fnc == nil {
			continue
		}

		name := filepath.Base(fnc.Name())
		if name == "runtime.goexit" {
			continue
		}
		file, line := fnc.FileLine(fp)
		file = formatStackTraceFileName(file)

		if i != 0 {
			trace += sep
		}
		trace += fmt.Sprintf("%s:%d %s", file, line, name)
	}
	return trace
}

func formatStackTraceFileName(file string) string {
	file = strings.ReplaceAll(file, "\\", "/")
	file = excludeGoPathSrcRegexp.ReplaceAllString(file, "")
	file = excludeGoPathGithubFpawelSrcRegexp.ReplaceAllString(file, "")
	file = excludeGoPathPkgModRegexp.ReplaceAllString(file, "")
	file = excludeGoPathPkgMod2Regexp.ReplaceAllString(file, "")
	return file
}

var excludeGoPathSrcRegexp = regexp.MustCompile(`[A-Z]:[^/]*/GOPATH/src/`)
var excludeGoPathPkgModRegexp = regexp.MustCompile(`[A-Z]:[^/]*/GOPATH/pkg/mod/`)
var excludeGoPathGithubFpawelSrcRegexp = regexp.MustCompile(`github.com/fpawel/`)
var excludeGoPathPkgMod2Regexp = regexp.MustCompile(`@v[^/]+`)

func captureStack(skip int) []uintptr {
	const maxStackDepth = 100
	stack := make([]uintptr, maxStackDepth)
	length := runtime.Callers(3+skip, stack[:])
	return stack[:length]
}
