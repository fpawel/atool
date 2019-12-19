package app

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/thriftgen/api"
)

type helperSvc struct{}

var _ api.HelperService = &helperSvc{}

func (_ *helperSvc) FormatWrite32BCD(_ context.Context, s string) (string, error) {
	n, v, err := parseDevCmdAndFloat(s)
	if err != nil {
		return "", merry.Append(err, "ожидался номер и аргумент команды")
	}
	return formatBytes(requestWrite32Bytes(n, v, config.BCD)), nil
}

func (_ *helperSvc) FormatWrite32FloatBE(_ context.Context, s string) (string, error) {
	n, v, err := parseDevCmdAndFloat(s)
	if err != nil {
		return "", merry.Append(err, "ожидался номер и аргумент команды")
	}
	return formatBytes(requestWrite32Bytes(n, v, config.FloatBigEndian)), nil
}
func (_ *helperSvc) FormatWrite32FloatLE(_ context.Context, s string) (string, error) {
	n, v, err := parseDevCmdAndFloat(s)
	if err != nil {
		return "", merry.Append(err, "ожидался номер и аргумент команды")
	}
	return formatBytes(requestWrite32Bytes(n, v, config.FloatLittleEndian)), nil
}
