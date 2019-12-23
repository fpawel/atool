package app

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/comm/modbus"
)

type helperSvc struct{}

var _ api.HelperService = &helperSvc{}

func (_ *helperSvc) FormatWrite32BCD(_ context.Context, s string) (string, error) {
	n, v, err := parseDevCmdAndFloat(s)
	if err != nil {
		return "", merry.Append(err, "ожидался номер и аргумент команды")
	}
	b := modbus.RequestWrite32{
		DeviceCmd: n,
		Format:    modbus.BCD,
		Value:     v,
	}.Request().Data
	return formatBytes(b), nil
}

func (_ *helperSvc) FormatWrite32FloatBE(_ context.Context, s string) (string, error) {
	n, v, err := parseDevCmdAndFloat(s)
	if err != nil {
		return "", merry.Append(err, "ожидался номер и аргумент команды")
	}
	b := modbus.RequestWrite32{
		DeviceCmd: n,
		Format:    modbus.FloatBigEndian,
		Value:     v,
	}.Request().Data
	return formatBytes(b), nil
}
func (_ *helperSvc) FormatWrite32FloatLE(_ context.Context, s string) (string, error) {
	n, v, err := parseDevCmdAndFloat(s)
	if err != nil {
		return "", merry.Append(err, "ожидался номер и аргумент команды")
	}
	b := modbus.RequestWrite32{
		DeviceCmd: n,
		Format:    modbus.FloatBigEndian,
		Value:     v,
	}.Request().Data
	return formatBytes(b), nil
}
