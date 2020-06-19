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
	"unicode/utf8"
)

func main() {
	j := new(logjrn.J)
	must.PanicIf(j.Open(filepath.Join(filepath.Dir(os.Args[0]), "logjrn.sqlite")))

	recs := logfile.ReadAllJournal()
	bar := progressbar.NewOptions(len(recs), progressbar.OptionSetPredictTime(true))

	reStack := regexp.MustCompile(`⤥([^⤣]+)⤣`)

	totalCount := len(recs)

	for n := 0; n < len(recs); {
		recs := recs[n:]
		offset := len(recs)
		if offset > 10000 {
			offset = 10000
		}
		recs = recs[:offset]
		n += offset

		var entList []*logjrn.Entry
		for _, rec := range recs {
			ent := &logjrn.Entry{
				StoredAt: rec.Time,
				Text:     strings.TrimSpace(rec.Text),
				Ok:       rec.Ok,
				Indent:   rec.Level,
			}

			xs := reStack.FindStringSubmatch(rec.Text)
			if len(xs) == 2 {
				ent.Stack = xs[1]
				ent.Text = reStack.ReplaceAllString(ent.Text, "")
			}
			ent.Text = removeInvalidRunes(ent.Text)
			ent.Stack = removeInvalidRunes(ent.Stack)
			entList = append(entList, ent)
		}

		must.PanicIf(j.AddEntries(entList))
		must.PanicIf(bar.Add(len(entList)))

		bar.Describe(fmt.Sprintf("%d из %d", n, totalCount))
	}

	must.PanicIf(bar.Finish())
	must.PanicIf(j.Close())
}

func removeInvalidRunes(s string) string {
	var xs []rune
	for _, r := range s {
		if utf8.ValidRune(r) {
			xs = append(xs, r)
		}
	}
	return string(xs)
}
