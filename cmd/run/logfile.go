package main

import (
	"fmt"
	"github.com/fpawel/atool/cmd/run/internal/ccolor"
	"github.com/powerman/structlog"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func execWithLogfile() error {
	if len(os.Args) < 2 {
		log.Fatal("file name to execute must be set")
	}
	name := os.Args[1]
	var args []string
	if len(os.Args) > 2 {
		args = os.Args[2:]
	}

	logFileOutput, err := newOutput("." + filepath.Base(name))
	if err != nil {
		return err
	}
	defer log.ErrIfFail(logFileOutput.Close)

	newWriter := func(f *os.File) io.Writer {
		return io.MultiWriter(logFileOutput, ccolor.NewWriter(f))
	}

	cmd := exec.Command(name, args...)
	cmd.Stderr = newWriter(os.Stderr)
	cmd.Stdout = newWriter(os.Stdout)

	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func newOutput(filenameSuffix string) (*os.File, error) {
	if err := ensureDir(); err != nil {
		return nil, err
	}
	filename := filename(daytime(time.Now()), filenameSuffix)
	return os.OpenFile(filename, os.O_CREATE|os.O_APPEND, 0666)
}

func filename(t time.Time, suffix string) string {
	return filepath.Join(logDir, fmt.Sprintf("%s%s.log", t.Format("2006-01-02"), suffix))
}

func daytime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

func ensureDir() error {
	_, err := os.Stat(logDir)
	if os.IsNotExist(err) { // создать каталог если его нет
		err = os.MkdirAll(logDir, os.ModePerm)
	}
	return err
}

func mustPanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

var (
	log    = structlog.New()
	logDir = filepath.Join(filepath.Dir(os.Args[0]), "logs")
)
