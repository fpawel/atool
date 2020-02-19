package data

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/comm/modbus"
	"github.com/jmoiron/sqlx"
	"time"
)

//go:generate go run github.com/fpawel/gotools/cmd/sqlstr/...

func Open(filename string) (*sqlx.DB, error) {
	db, err := pkg.OpenSqliteDBx(filename)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(SQLCreate); err != nil {
		return nil, err
	}
	if _, err := GetCurrentParty(db); err != nil {
		return nil, err
	}
	return db, nil
}

func GetCurrentPartyID(db *sqlx.DB) (int64, error) {
	var partyID int64
	err := db.Get(&partyID, `SELECT party_id FROM app_config`)
	if err != nil {
		return 0, err
	}
	return partyID, nil
}

// момент времени последнего обновления текущей партии
func GetCurrentPartyUpdatedAt(db *sqlx.DB) (time.Time, error) {
	var tm string
	err := db.Get(&tm, `
WITH q AS ( SELECT party_id FROM app_config )
SELECT STRFTIME('%Y-%m-%d %H:%M:%f', tm) AS tm
FROM measurement
INNER JOIN product USING (product_id)
WHERE party_id = (SELECT q.party_id FROM q)
ORDER BY tm DESC
LIMIT 1`)
	if err == sql.ErrNoRows {
		return time.Time{}, sql.ErrNoRows
	}
	return parseTime(tm), nil
}

func GetCurrentParty(db *sqlx.DB) (Party, error) {
	partyID, err := GetCurrentPartyID(db)
	if err != nil {
		return Party{}, err
	}
	return GetParty(db, partyID)
}

func GetParty(db *sqlx.DB, partyID int64) (Party, error) {
	var party Party
	err := db.Get(&party, `SELECT * FROM party WHERE party_id=?`, partyID)
	if err != nil {
		return Party{}, err
	}

	if err := db.Select(&party.Products,
		`SELECT * FROM product_enumerated WHERE party_id=? ORDER BY place`,
		partyID); err != nil {
		return Party{}, err
	}
	return party, nil
}

func SetCurrentPartyValues(db *sqlx.DB, p PartyValues) error {
	partyID, err := GetCurrentPartyID(db)
	if err != nil {
		return err
	}
	p.PartyID = partyID
	if err := setPartyValues(db, p); err != nil {
		return err
	}
	return nil
}

func GetPartyValues(db *sqlx.DB, partyID int64, party *PartyValues) error {

	err := db.Get(party, `SELECT * FROM party WHERE party_id=?`, partyID)
	if err != nil {
		return err
	}

	party.Values = make(map[string]float64)
	var values []struct {
		Key   string  `db:"key"`
		Value float64 `db:"value"`
	}
	if err := db.Select(&values, `SELECT key, value FROM party_value WHERE party_id=?`, partyID); err != nil {
		return err
	}
	for _, x := range values {
		party.Values[x.Key] = x.Value
	}

	var xs []struct {
		ProductID int64       `db:"product_id"`
		Serial    int         `db:"serial"`
		Addr      modbus.Addr `db:"addr"`
		Key       string      `db:"key"`
		Value     float64     `db:"value"`
	}
	if err := db.Select(&xs, `
SELECT product_id, serial, addr, key, value FROM product_value 
INNER JOIN product USING(product_id)
WHERE party_id= ?
ORDER BY created_at, created_order`, partyID); err != nil {
		return err
	}
	for _, x := range xs {
		var p *ProductValues
		for i := range party.Products {
			if party.Products[i].ProductID == x.ProductID {
				p = &party.Products[i]
				break
			}
		}
		if p == nil {
			p = &ProductValues{
				ProductID: x.ProductID,
				Place:     len(party.Products),
				Serial:    x.Serial,
				Addr:      x.Addr,
				Values:    make(map[string]float64),
			}
			party.Products = append(party.Products, *p)
		}
		p.Values[x.Key] = x.Value
	}
	return nil
}

func GetCurrentPartyValues(db *sqlx.DB, party *PartyValues) error {
	partyID, err := GetCurrentPartyID(db)
	if err != nil {
		return err
	}
	return GetPartyValues(db, partyID, party)
}

func CopyCurrentParty(db *sqlx.DB) error {
	prevParty, err := GetCurrentParty(db)
	if err != nil {
		return err
	}

	newPartyID, err := createNewParty(db, prevParty.Name, prevParty.DeviceType, prevParty.ProductType)
	if err != nil {
		return err
	}

	if err := setAppConfigPartyID(db, newPartyID); err != nil {
		return err
	}

	if _, err := db.Exec(`
INSERT INTO party_value
SELECT ?, key, value FROM product_value 
WHERE product_id = ?`, newPartyID, prevParty.PartyID); err != nil {
		return err
	}

	for i, p := range prevParty.Products {
		p.CreatedAt = time.Now()
		p.CreatedOrder = i
		p.PartyID = newPartyID

		r, err := db.NamedExec(`
INSERT INTO product( party_id, addr, active, comport, created_at, created_order ) 
VALUES (:party_id, :addr, :active, :comport, :created_at, :created_order);`, p)

		if err != nil {
			return err
		}
		newProductID, err := getNewInsertedID(r)
		if err != nil {
			return err
		}

		if _, err := db.Exec(`
INSERT INTO product_param
SELECT ?, param_addr, chart, series_active FROM product_param 
WHERE product_id = ?`, newProductID, p.ProductID); err != nil {
			return err
		}

		if _, err := db.Exec(`
INSERT INTO product_value
SELECT ?, key, value FROM product_value 
WHERE product_id = ?`, newProductID, p.ProductID); err != nil {
			return err
		}
	}

	return nil
}

func SetProductParam(db *sqlx.DB, p ProductParam) error {
	_, err := db.NamedExec(`
INSERT INTO product_param (product_id, param_addr, chart, series_active)
VALUES (:product_id, :param_addr, :chart, :series_active)
ON CONFLICT (product_id, param_addr) DO UPDATE SET series_active=:series_active,
                                                   chart=:chart`, p)
	return err
}

func SetNewCurrentParty(ctx context.Context, db *sqlx.DB, productsCount int) error {

	prevParty, err := GetCurrentParty(db)
	if err != nil {
		return err
	}
	t := time.Now()
	name := fmt.Sprintf("%d %s %s, %s",
		t.Day(), formatMonth(t), t.Format("2006"), t.Format("15:04"))

	newPartyID, err := createNewParty(db, name, prevParty.DeviceType, prevParty.ProductType)
	if err != nil {
		return err
	}
	for i := 0; i < productsCount; i++ {
		r, err := db.ExecContext(ctx,
			`INSERT INTO product(party_id, addr, created_order, created_at) VALUES (?, ?, ?, ?);`,
			newPartyID, i+1, i+1, time.Now().Add(time.Second*time.Duration(i)))
		if err != nil {
			return err
		}
		if _, err = getNewInsertedID(r); err != nil {
			return err
		}
	}
	if err := setAppConfigPartyID(db, newPartyID); err != nil {
		return err
	}
	return nil
}

func AddNewProduct(db *sqlx.DB, order int) (int64, error) {
	party, err := GetCurrentParty(db)
	if err != nil {
		return 0, err
	}
	addresses := make(map[modbus.Addr]struct{})
	for _, x := range party.Products {
		addresses[x.Addr] = struct{}{}
	}
	addr := modbus.Addr(1)
	for ; addr <= modbus.Addr(255); addr++ {
		if _, f := addresses[addr]; !f {
			break
		}
	}
	r, err := db.Exec(
		`INSERT INTO product( party_id, addr, created_order) VALUES (?,?,?)`,
		party.PartyID, addr, order)
	if err != nil {
		return 0, err
	}

	productID, err := getNewInsertedID(r)
	if err != nil {
		return 0, err
	}
	return productID, nil
}

func DeleteCurrentParty(db *sqlx.DB) error {
	var newCurrentPartyID int64
	const q1 = `
SELECT party_id 
FROM party 
WHERE party_id != (SELECT party_id FROM app_config WHERE id=1)
ORDER BY created_at  DESC 
LIMIT 1`
	if err := db.Get(&newCurrentPartyID, q1); err != nil {
		return err
	}

	partyID, err := GetCurrentPartyID(db)
	if err != nil {
		return err
	}

	if _, err := db.Exec(`UPDATE app_config SET party_id = ? WHERE id=1`, newCurrentPartyID); err != nil {
		return err
	}

	if _, err := db.Exec(`DELETE FROM party WHERE party_id=?`, partyID); err != nil {
		return err
	}
	return nil
}

func setAppConfigPartyID(db *sqlx.DB, partyID int64) error {
	_, err := db.Exec(`UPDATE app_config SET party_id=? WHERE id=1`, partyID)
	return err
}

func createNewParty(db *sqlx.DB, name, deviceType, productType string) (int64, error) {
	r, err := db.Exec(`INSERT INTO party (created_at, name, device_type, product_type) VALUES (?,?,?,?)`,
		time.Now(), name, deviceType, productType)
	if err != nil {
		return 0, err
	}
	n, err := r.RowsAffected()
	if err != nil {
		return 0, err
	}
	if n != 1 {
		return 0, fmt.Errorf("excpected 1 rows affected, got %d", n)
	}
	return getNewInsertedID(r)
}

func setPartyValues(db *sqlx.DB, p PartyValues) error {

	const q1 = `UPDATE party SET name=?, device_type = ?, product_type=? WHERE party_id=?`
	if _, err := db.Exec(q1, p.Name, p.DeviceType, p.ProductType, p.PartyID); err != nil {
		return err
	}
	if _, err := db.Exec(`DELETE FROM product WHERE party_id=?`, p.PartyID); err != nil {
		return err
	}
	if _, err := db.Exec(`DELETE FROM party_value WHERE party_id=?`, p.PartyID); err != nil {
		return err
	}

	var sqlStr string
	for k, v := range p.Values {
		if len(sqlStr) > 0 {
			sqlStr += ","
		}
		sqlStr += fmt.Sprintf("(%d, '%s', %v)", p.PartyID, k, v)
	}
	if len(sqlStr) > 0 {
		if _, err := db.Exec(`INSERT INTO party_value(party_id, key, value) VALUES ` + sqlStr); err != nil {
			return err
		}
	}

	sqlStr = ""
	for i := range p.Products {
		var err error
		p.Products[i].ProductID, err = AddNewProduct(db, i)
		if err != nil {
			return err
		}
		p := p.Products[i]

		if _, err = db.Exec(`UPDATE product SET serial = ? WHERE product_id=?`, p.Serial, p.ProductID); err != nil {
			return err
		}

		_, _ = db.Exec(`UPDATE product SET addr = ? WHERE product_id=?`, p.Addr, p.ProductID)

		for k, v := range p.Values {
			if len(sqlStr) > 0 {
				sqlStr += ","
			}
			sqlStr += fmt.Sprintf("(%d, '%s', %v)", p.ProductID, k, v)
		}
	}
	if len(sqlStr) > 0 {
		if _, err := db.Exec(`INSERT INTO product_value(product_id, key, value) VALUES ` + sqlStr); err != nil {
			return err
		}
	}
	return nil
}
