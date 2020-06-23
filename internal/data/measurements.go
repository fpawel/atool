package data

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/comm/modbus"
	"strings"
	"time"
)

type Measurement struct {
	//Tm        float64 `db:"tm"`

	Tm        int64      `db:"tm"`
	ProductID int64      `db:"product_id"`
	ParamAddr modbus.Var `db:"param_addr"`
	Value     float64    `db:"value"`
}

func (x Measurement) Time() time.Time {
	//return pkg.JulianToTime(x.Tm)
	return time.Unix(0, x.Tm)
}

func NewMeasurement(ProductID int64, ParamAddr modbus.Var, Value float64) Measurement {
	return Measurement{
		Tm:        time.Now().UnixNano(),
		ProductID: ProductID,
		ParamAddr: ParamAddr,
		Value:     Value,
	}
}

func SaveMeasurements(measurements []Measurement) error {
	if len(measurements) == 0 {
		return nil
	}
	var xs []string
	for _, m := range measurements {
		time.Now().UnixNano()
		xs = append(xs, fmt.Sprintf("(%v,%d,%d,%v)", m.Tm, m.ProductID, m.ParamAddr, m.Value))
	}
	strQueryInsert := `INSERT INTO measurement(tm, product_id, param_addr, value) VALUES ` + "  " + strings.Join(xs, ",")
	if _, err := DB.Exec(strQueryInsert); err != nil {
		return merry.Append(err, strQueryInsert)
	}
	return nil
}

type MeasurementCache []Measurement

func (xs *MeasurementCache) Add(ProductID int64, ParamAddr modbus.Var, Value float64) {
	*xs = append(*xs, NewMeasurement(ProductID, ParamAddr, Value))
	if len(*xs) >= 1000 {
		xs.Save()
	}
}

func (xs *MeasurementCache) Save() {
	must.PanicIf(SaveMeasurements(*xs))
	*xs = nil
}

func GetPartyChart(partyID int64, paramAddresses []modbus.Var) ([]Measurement, error) {

	productsIDs, err := GetPartyProductsIDs(partyID)
	if err != nil {
		return nil, err
	}

	sQ := `SELECT tm, product_id, param_addr, value
FROM measurement
WHERE product_id IN ` +
		formatInt64SliceAsQuery(productsIDs) +
		` AND param_addr IN ` +
		formatIntSliceAsQuery2(paramAddresses)

	var xs []Measurement

	if err = DB.Select(&xs, sQ); err != nil {
		err = merry.Append(err, sQ)
	}
	return xs, err
}

func GetPartyProductsIDs(partyID int64) (xs []int64, err error) {
	err = DB.Select(&xs, `SELECT product_id FROM product WHERE party_id = ?`, partyID)
	return
}
