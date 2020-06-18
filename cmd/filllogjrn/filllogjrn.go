package main

import (
	"fmt"
	"github.com/fpawel/atool/internal/pkg/logfile"
	"github.com/fpawel/atool/internal/pkg/logjrn"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/schollz/progressbar/v3"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	j := new(logjrn.J)
	must.PanicIf(j.Open(filepath.Join(filepath.Dir(os.Args[0]), "logjrn.sqlite")))

	recs := logfile.ReadAllJournal()
	bar := progressbar.NewOptions(len(recs), progressbar.OptionSetPredictTime(true))

	reStack := regexp.MustCompile(`⤥([^⤣]+)⤣`)

	for i, rec := range recs {
		ent := &logjrn.Entry{
			StoredAt: rec.Time,
			Text:     strings.TrimSpace(rec.Text),
			Ok:       rec.Ok,
			Indent:   rec.Level,
		}

		xs := reStack.FindStringSubmatch(rec.Text)
		if len(xs) == 2 {
			ent.Stack = xs[1]
		}
		must.PanicIf(j.AddEntry(ent))
		must.PanicIf(bar.Add(1))
		bar.Describe(fmt.Sprintf("%d from %d", i+1, len(recs)))
	}
	must.PanicIf(bar.Finish())
	must.PanicIf(j.Close())
}
