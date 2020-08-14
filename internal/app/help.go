package app

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/comm/modbus"
	"github.com/powerman/structlog"
	"strconv"
	"strings"
	"time"
)

type logger = *structlog.Logger

func selectProductParamsChart(chart string) (string, string, error) {
	var xs []struct {
		ProductID int64      `db:"product_id"`
		ParamAddr modbus.Var `db:"param_addr"`
	}
	if err := data.DB.Select(&xs,
		`SELECT DISTINCT product_id,param_addr
FROM product_param
WHERE chart = ?
  AND series_active = TRUE
  AND product_id IN (SELECT product_id FROM product WHERE party_id = (SELECT party_id FROM app_config))`, chart); err != nil {
		return "", "", err
	}
	var qProductsXs, qParamsXs []string
	mProducts := map[int64]struct{}{}
	mParams := map[modbus.Var]struct{}{}
	for _, p := range xs {
		if _, f := mProducts[p.ProductID]; !f {
			mProducts[p.ProductID] = struct{}{}
			qProductsXs = append(qProductsXs, fmt.Sprintf("%d", p.ProductID))
		}
		if _, f := mParams[p.ParamAddr]; !f {
			mParams[p.ParamAddr] = struct{}{}
			qParamsXs = append(qParamsXs, fmt.Sprintf("%d", p.ParamAddr))
		}
	}
	return strings.Join(qProductsXs, ","), strings.Join(qParamsXs, ","), nil
}

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

func parseDevCmdAndFloat(s string) (modbus.DevCmd, float64, error) {
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
	return modbus.DevCmd(n), v, nil
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

func timeUnixMillis(t time.Time) apitypes.TimeUnixMillis {
	return apitypes.TimeUnixMillis(t.UnixNano() / int64(time.Millisecond))
}

func unixMillisToTime(m apitypes.TimeUnixMillis) time.Time {
	t := int64(time.Millisecond) * int64(m)
	sec := t / int64(time.Second)
	ns := t % int64(time.Second)
	return time.Unix(sec, ns)
}
