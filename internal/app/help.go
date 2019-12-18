package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/gui"
	"github.com/fpawel/atool/internal/cfg"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/powerman/structlog"
	"strconv"
	"strings"
	"time"
)

func formatBytes(xs []byte) string {
	return fmt.Sprintf("% X", xs)
}

func formatIDs(ids []int64) string {
	var ss []string
	for _, id := range ids {
		ss = append(ss, strconv.FormatInt(id, 10))
	}
	return strings.Join(ss, ",")
}

func requestWrite32Bytes(c uint16, v float64, f cfg.FloatBitsFormat) []byte {
	b := []byte{
		0, 32, 0, 3, 6,
		byte(c >> 8),
		byte(c),
		0, 0, 0, 0,
	}
	f.PutFloat(b[7:], v)
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

func formatFloat(v float64) string {
	//n := cfg.Get().FloatPrecision
	//if n > 0 {
	//	k := math.Pow10(n)
	//	v = math.Round(v*k)/k
	//}
	return strconv.FormatFloat(v, 'g', -1, 64)
}

func formatTimeAsQuery(t time.Time) string {
	return "julianday(STRFTIME('%Y-%m-%d %H:%M:%f','" +
		t.Format(timeLayout) + "'))"
}

func timeUnixMillis(t time.Time) apitypes.TimeUnixMillis {

	return apitypes.TimeUnixMillis(t.UnixNano() / int64(time.Millisecond))
}

func unixMillisToTime(m apitypes.TimeUnixMillis) time.Time {
	t := int64(time.Millisecond) * int64(m)
	sec := t / int64(time.Second)
	ns := t % int64(time.Second)
	return time.Unix(sec, ns)
}

const timeLayout = "2006-01-02 15:04:05.000"

type commTransaction struct {
	comportName string
	what        string
	device      cfg.Device
	req         modbus.Request
	prs         comm.ResponseParser
}

func (x commTransaction) getResponse(log *structlog.Logger, ctx context.Context) ([]byte, error) {
	startTime := time.Now()
	rdr, err := wrk.getResponseReader(x.comportName, x.device)
	if err != nil {
		return nil, err
	}
	response, err := x.req.GetResponse(log, ctx, rdr, x.prs)
	if merry.Is(err, context.Canceled) {
		return response, err
	}
	ct := gui.CommTransaction{
		Addr:     x.req.Addr,
		Comport:  x.comportName,
		Request:  formatBytes(x.req.Bytes()),
		Response: formatBytes(response) + " " + time.Since(startTime).String(),
		Ok:       err == nil,
	}
	if len(x.what) > 0 {
		ct.Request += " " + x.what
	}

	if err != nil {
		if len(response) == 0 {
			ct.Response = err.Error()
		} else {
			ct.Response += " " + err.Error()
		}
	}
	go gui.NotifyNewCommTransaction(ct)
	return response, err
}
