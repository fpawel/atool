package logfile

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/powerman/structlog"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func New(filenameSuffix string) (io.WriteCloser, error) {
	if err := ensureDir(); err != nil {
		return nil, err
	}
	filename := filename(daytime(time.Now()), filenameSuffix)
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND, 0666)
	return &output{f: f}, err
}

type output struct {
	f    *os.File
	mu   sync.Mutex
	done bool
}

func (x *output) Close() error {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.done = true
	return x.f.Close()
}

func (x *output) Write(p []byte) (int, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	if x.done {
		return 0, errors.New("file was closed")
	}
	go func() {
		x.mu.Lock()
		defer x.mu.Unlock()
		for _, p := range bytes.Split(p, []byte{'\n'}) {
			if len(p) == 0 {
				continue
			}
			if _, err := fmt.Fprint(x.f, time.Now().Format(layoutDatetime), " "); err != nil {
				panic(err)
			}
			if _, err := x.f.Write(p); err != nil {
				panic(err)
			}
			if !bytes.HasSuffix(p, []byte("\n")) {
				if _, err := x.f.WriteString("\n"); err != nil {
					panic(err)
				}
			}
		}
	}()
	return len(p), nil
}

func daytime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

func filename(t time.Time, suffix string) string {
	return filepath.Join(logDir, fmt.Sprintf("%s%s.log", t.Format(layoutDate), suffix))
}

func ensureDir() error {
	_, err := os.Stat(logDir)
	if os.IsNotExist(err) { // создать каталог если его нет
		err = os.MkdirAll(logDir, os.ModePerm)
	}
	return err
}

const (
	layoutDatetime = "2006-01-02-15:04:05.000"
	layoutDate     = "2006-01-02"
)

var (
	log    = structlog.New()
	logDir = filepath.Join(filepath.Dir(os.Args[0]), "logs")
)
