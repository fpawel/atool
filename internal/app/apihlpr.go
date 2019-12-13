package app

import (
	"context"
	"encoding/binary"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/comm/modbus"
	"math"
	"strconv"
	"strings"
)

type helperSvc struct{}

var _ api.HelperService = &helperSvc{}

func (_ *helperSvc) FormatWrite32BCD(_ context.Context, s string) (string, error) {
	n, v, err := parseDevCmdAndFloat(s)
	if err != nil {
		return "", merry.Append(err, "ожидался номер и аргумент команды")
	}
	return formatBytes(requestWrite32Bytes(n, v, ffBCD)), nil
}

func (_ *helperSvc) FormatWrite32FloatBE(_ context.Context, s string) (string, error) {
	n, v, err := parseDevCmdAndFloat(s)
	if err != nil {
		return "", merry.Append(err, "ожидался номер и аргумент команды")
	}
	return formatBytes(requestWrite32Bytes(n, v, ffBE)), nil
}
func (_ *helperSvc) FormatWrite32FloatLE(_ context.Context, s string) (string, error) {
	n, v, err := parseDevCmdAndFloat(s)
	if err != nil {
		return "", merry.Append(err, "ожидался номер и аргумент команды")
	}
	return formatBytes(requestWrite32Bytes(n, v, ffLE)), nil
}

type floatFormat int

const (
	ffBCD floatFormat = iota
	ffBE
	ffLE
)

func requestWrite32Bytes(c uint16, v float64, f floatFormat) []byte {
	b := []byte{
		0, 32, 0, 3, 6,
		byte(c >> 8),
		byte(c),
		0, 0, 0, 0,
	}

	d := b[7:]
	switch f {
	case ffBCD:
		modbus.PutBCD6(d, v)
	case ffBE:
		n := math.Float32bits(float32(v))
		binary.BigEndian.PutUint32(d, n)
	case ffLE:
		n := math.Float32bits(float32(v))
		binary.LittleEndian.PutUint32(d, n)
	default:
		panic(f)
	}
	return b
}

func parseDevCmdAndFloat(s string) (uint16, float64, error) {
	s1, s2, err := parseTwoWords(s)
	if err != nil {
		return 0, 0, merry.Append(err, "ожидалось два слова")
	}
	n, err := strconv.ParseUint(s1, 10, 16)
	if err != nil {
		return 0, 0, merry.Append(err, "код команды прибору")
	}
	v, err := parseFloat(s2)
	if err != nil {
		return 0, 0, merry.Append(err, "значение аргумента")
	}
	return uint16(n), v, nil
}

func parseTwoWords(s string) (s1, s2 string, err error) {
	xs := strings.Fields(s)
	if len(xs) == 0 {
		return "", "", merry.New("пустая строка")
	}
	if len(xs) == 1 {
		return "", "", merry.New("одно слово")
	}
	return xs[0], xs[1], nil
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(strings.Replace(s, ",", ".", -1), 64)
}

func parseHexBytes(s string) ([]byte, error) {
	var xs []byte
	s = strings.TrimSpace(s)
	for i, strB := range strings.Split(s, " ") {
		v, err := strconv.ParseUint(strB, 16, 8)
		if err != nil {
			return nil, merry.Appendf(err, "поз.%d: %q", i+1, strB)
		}
		if v < 0 || v > 0xff {
			return nil, merry.Errorf("поз.%d: %q: ожижалось шестнадцатиричное число от 0 до FF", i+1, strB)
		}
		xs = append(xs, byte(v))
	}
	return xs, nil
}
