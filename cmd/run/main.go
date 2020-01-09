package main

import (
	"github.com/fpawel/atool/cmd/run/internal/ccolor"
	"github.com/fpawel/atool/internal/pkg/logfile"
	"github.com/powerman/structlog"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	structlog.DefaultLogger.
		SetPrefixKeys(
			structlog.KeyApp,
			structlog.KeyPID, structlog.KeyLevel, structlog.KeyUnit, structlog.KeyTime,
		).
		SetDefaultKeyvals(
			structlog.KeyApp, filepath.Base(os.Args[0]),
			structlog.KeySource, structlog.Auto,
		).
		SetSuffixKeys(
			structlog.KeyStack,
		).
		SetSuffixKeys(structlog.KeySource).
		SetKeysFormat(map[string]string{
			structlog.KeyTime:   " %[2]s",
			structlog.KeySource: " %6[2]s",
			structlog.KeyUnit:   " %6[2]s",
		})

	log.ErrIfFail(func() error {
		return execWithLogfile()
	})
}

func execWithLogfile() error {
	if len(os.Args) < 2 {
		log.Fatal("file name to execute must be set")
	}
	name := os.Args[1]
	var args []string
	if len(os.Args) > 2 {
		args = os.Args[2:]
	}

	logFileOutput, err := logfile.New("." + filepath.Base(name))
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

var log = structlog.New()
