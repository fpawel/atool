package ccolor

import (
	"bytes"
	"io"
	"os"
)

type Output struct {
	f *os.File
}

func NewWriter(f *os.File) io.Writer {
	return &Output{f: f}
}

func (x *Output) Write(p []byte) (int, error) {

	fields := bytes.Fields(p)
	if len(fields) > 1 {
		switch string(fields[1]) {
		case "ERR":
			Foreground(Red, true)
		case "WRN":
			Foreground(Yellow, true)
		case "inf":
			Foreground(White, true)
		default:
			Foreground(White, false)
		}
	}
	defer ResetColor()

	return x.f.Write(p)
}
