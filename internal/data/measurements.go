package data

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
)

type Measurement struct {
	Tm        float64 `db:"tm"`
	ProductID int64   `db:"product_id"`
	ParamAddr int     `db:"param_addr"`
	Value     float64 `db:"value"`
}

func (x Measurement) Time() time.Time {
	return pkg.JulianToTime(x.Tm)
}

func NewMeasurement(ProductID int64, ParamAddr int, Value float64) Measurement {
	return Measurement{
		Tm:        pkg.TimeToJulian(time.Now()),
		ProductID: ProductID,
		ParamAddr: ParamAddr,
		Value:     Value,
	}
}

func SaveMeasurements(db *sqlx.DB, measurements []Measurement) error {
	if len(measurements) == 0 {
		return nil
	}
	var xs []string
	for _, m := range measurements {
		xs = append(xs, fmt.Sprintf("(%v,%d,%d,%v)", m.Tm, m.ProductID, m.ParamAddr, m.Value))
	}
	strQueryInsert := `INSERT INTO measurement(tm, product_id, param_addr, value) VALUES ` + "  " + strings.Join(xs, ",")
	if _, err := db.Exec(strQueryInsert); err != nil {
		return merry.Append(err, strQueryInsert)
	}
	return nil
}

func GetPartyChart(db *sqlx.DB, partyID int64, paramAddresses []int) ([]Measurement, error) {

	productsIDs, err := GetPartyProductsIDs(db, partyID)
	if err != nil {
		return nil, err
	}

	sQ := `SELECT tm, product_id, param_addr, value
FROM measurement
WHERE product_id IN ` +
		formatInt64SliceAsQuery(productsIDs) +
		` AND param_addr IN ` +
		formatIntSliceAsQuery(paramAddresses)

	var xs []Measurement

	if err = db.Select(&xs, sQ); err != nil {
		err = merry.Append(err, sQ)
	}
	return xs, err
}

func GetPartyProductsIDs(db *sqlx.DB, partyID int64) (xs []int64, err error) {
	err = db.Select(&xs, `SELECT product_id FROM product WHERE party_id = ?`, partyID)
	return
}
