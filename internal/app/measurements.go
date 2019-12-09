package app

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/data"
	"strings"
	"time"
)

type measurements struct {
	xs []data.Measurement
}

func (x *measurements) add(ProductID int64, ParamAddr int, Value float64) {
	x.xs = append(x.xs, data.Measurement{
		Time:      time.Now(),
		ProductID: ProductID,
		ParamAddr: ParamAddr,
		Value:     Value,
	})
	if len(x.xs) >= 1000 {
		saveMeasurements(x.xs)
		x.xs = nil
	}
}

func saveMeasurements(measurements []data.Measurement) {
	if len(measurements) == 0 {
		return
	}
	var xs []string
	for _, m := range measurements {
		xs = append(xs, fmt.Sprintf("(%s,%d,%d,%v)", formatTimeAsQuery(m.Time), m.ProductID, m.ParamAddr, m.Value))
	}
	strQueryInsert := `INSERT INTO measurement(tm, product_id, param_addr, value) VALUES ` + "  " + strings.Join(xs, ",")
	if _, err := db.Exec(strQueryInsert); err != nil {
		err = merry.Appendf(err, "fail to insert measurements: %q", strQueryInsert)
		log.PrintErr(err)
	}
}

func formatTimeAsQuery(t time.Time) string {
	return "julianday(STRFTIME('%Y-%m-%d %H:%M:%f','" +
		t.Format(timeLayout) + "'))"
}

const timeLayout = "2006-01-02 15:04:05.000"
