package app

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/comm/modbus"
	"github.com/powerman/structlog"
	"math"
	"strconv"
	"strings"
	"time"
)

type logger = *structlog.Logger

func getCurrentPartyDeviceConfig() (config.Device, error) {
	party, err := data.GetCurrentParty(db)
	if err != nil {
		return config.Device{}, err
	}
	return config.Get().Hardware.GetDevice(party.DeviceType)
}

func getActiveProducts() ([]data.Product, error) {

	var products []data.Product
	err := db.Select(&products,
		`SELECT * FROM product_enumerated WHERE party_id = (SELECT party_id FROM app_config) AND active`)
	if err != nil {
		return nil, err
	}
	if len(products) == 0 {
		return nil, errNoInterrogateObjects
	}
	return products, nil
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

func formatFloat(v float64) string {
	n := config.Get().FloatPrecision
	if n > 0 {
		k := math.Pow10(n)
		v = math.Round(v*k) / k
	}
	return strconv.FormatFloat(v, 'g', -1, 64)
}

func timeUnixMillis(t time.Time) apitypes.TimeUnixMillis {
	return apitypes.TimeUnixMillis(t.UnixNano() / int64(time.Millisecond))
}

func formatError1(err error) string {
	return strings.Replace(err.Error(), ":", "\n\t", -1)
}

func unixMillisToTime(m apitypes.TimeUnixMillis) time.Time {
	t := int64(time.Millisecond) * int64(m)
	sec := t / int64(time.Second)
	ns := t % int64(time.Second)
	return time.Unix(sec, ns)
}

const timeLayout = "2006-01-02 15:04:05.000"

func pause(chDone <-chan struct{}, d time.Duration) {
	timer := time.NewTimer(d)
	for {
		select {
		case <-timer.C:
			return
		case <-chDone:
			timer.Stop()
			return
		}
	}
}

func getCurrentPartyValues() (map[string]float64, error) {
	var xs []struct {
		Key   string  `db:"key"`
		Value float64 `db:"value"`
	}
	const q1 = `SELECT key, value FROM party_value WHERE party_id = (SELECT party_id FROM app_config)`
	if err := db.Select(&xs, q1); err != nil {
		return nil, merry.Append(err, q1)
	}
	m := map[string]float64{}
	for _, x := range xs {
		m[x.Key] = x.Value
	}
	return m, nil
}

func deleteProductKey(productID int64, key string) error {
	const q1 = `DELETE FROM product_value WHERE product_id = ? AND key = ?`
	_, err := db.Exec(q1, productID, key)
	return merry.Appendf(err, "%s, %s", q1, key)
}

func saveProductValue(productID int64, key string, value float64) error {
	const q1 = `
INSERT INTO product_value
VALUES (?, ?, ?)
ON CONFLICT (product_id,key) DO UPDATE
    SET value = ?`
	_, err := db.Exec(q1, productID, key, value, value)
	return merry.Appendf(err, "%s, %s: %v", q1, key, value)
}

func dbKeyCoefficient(k int) string {
	return fmt.Sprintf("K%02d", k)
}

func getProductValues(productID int64) (map[string]float64, error) {
	var xs []struct {
		Key   string  `db:"key"`
		Value float64 `db:"value"`
	}
	const q1 = `SELECT key, value FROM product_value WHERE product_id = ?`
	if err := db.Select(&xs, q1, productID); err != nil {
		return nil, merry.Append(err, q1)
	}
	m := map[string]float64{}
	for _, x := range xs {
		m[x.Key] = x.Value
	}
	return m, nil
}
