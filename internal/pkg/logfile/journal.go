package logfile

import (
	"bytes"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/must"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"
)

type Journal struct {
	f   *os.File
	lev func() int
}

type JournalRecord struct {
	Time  time.Time
	Text  string
	Level int
	Ok    bool
}

func NewJournal(lev func() int) Journal {
	f, err := os.OpenFile(filename(daytime(time.Now()), ".journal"), os.O_CREATE|os.O_APPEND, 0666)
	must.PanicIf(err)
	return Journal{
		lev: lev,
		f:   f,
	}
}

func (x Journal) Close() error {
	return x.f.Close()
}

func (x Journal) WriteError(err error) {
	x.WriteErrorText(err.Error() + " ⚠️\n\t" + "⤥" + pkg.FormatStacktrace(merry.Stack(err), "\n\t") + "⤣")
}

func (x Journal) WriteErrorText(text string) {
	x.Write("⚠️ " + text)
}

func (x Journal) Write(text string) {
	_, err := fmt.Fprintf(x.f, "%s%s↩\n", prefix(x.lev()), text)
	must.PanicIf(err)
}

func prefix(lev int) string {
	return fmt.Sprintf("%s [%d] %s", time.Now().Format("15:04:05"), lev, strings.Repeat("    ", lev))
}

func ReadJournal() (result []JournalRecord) {
	for _, daytime := range listDays() {
		xs, err := readJournalFile(daytime)
		must.PanicIf(err)
		result = append(xs, result...)
		if len(result) > 1000 {
			break
		}
	}
	return
}

func parseRecord(s string, daytime time.Time, r *JournalRecord) bool {
	trimPrefix := func(prefix string) {
		s = strings.TrimPrefix(s, prefix)
		if len(s) > 0 {
			s = s[1:]
		}
	}

	words := strings.Fields(s)
	if len(words) < 3 {
		return false
	}

	t, err := time.Parse("15:04:05", words[0])
	if err != nil {
		return false
	}
	y, m, d := daytime.Date()
	r.Time = time.Date(y, m, d, t.Hour(), t.Minute(), t.Second(), 0, time.Local)

	trimPrefix(words[0])

	r.Level, err = strconv.Atoi(strings.Trim(words[1], "[]"))
	if err != nil {
		return false
	}
	trimPrefix(words[1])

	r.Ok = true
	if len(words) > 3 && words[2] == "⚠️" {
		r.Ok = false
		//trimPrefix(words[2])
	}
	r.Text = s
	return true
}

func readJournalFile(daytime time.Time) ([]JournalRecord, error) {
	filename := filename(daytime, ".journal")
	b, err := ioutil.ReadFile(filename)
	if err == syscall.ERROR_FILE_NOT_FOUND {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var recs []JournalRecord

	n := bytes.IndexRune(b, '↩')
	for ; n != -1; n = bytes.IndexRune(b, '↩') {
		s := string(b[:n])
		b = b[n+utf8.RuneLen('↩')+1:]
		var rec JournalRecord
		if parseRecord(s, daytime, &rec) {
			recs = append(recs, rec)
		}
	}
	return recs, nil
}

func listDays() []time.Time {
	r := regexp.MustCompile(`\d\d\d\d-\d\d-\d\d`)
	m := make(map[time.Time]struct{})
	_ = filepath.Walk(logDir, func(path string, f os.FileInfo, _ error) error {
		if f == nil || f.IsDir() {
			return nil
		}
		if !strings.HasSuffix(f.Name(), ".journal.log") {
			return nil
		}
		s := r.FindString(f.Name())
		if len(s) == 0 {
			return nil
		}
		t, err := time.Parse(layoutDate, s)
		if err != nil {
			return nil
		}
		m[daytime(t)] = struct{}{}
		return nil
	})
	var days []time.Time
	for t := range m {
		days = append(days, t)
	}
	sort.Slice(days, func(i, j int) bool {
		return days[i].After(days[j])
	})
	return days
}

const (
	layoutTime = "15:04:05"
	layoutDate = "2006-01-02"
)
