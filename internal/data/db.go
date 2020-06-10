package data

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/comm/modbus"
	"github.com/jmoiron/sqlx"
	"time"
)

//go:generate go run github.com/fpawel/gotools/cmd/sqlstr/...

var DB *sqlx.DB

var errNoInterrogateObjects = merry.New("не установлены объекты опроса")

func Open(filename string) error {
	var err error
	DB, err = pkg.OpenSqliteDBx(filename)
	if err != nil {
		return err
	}
	if _, err := DB.Exec(SQLCreate); err != nil {
		return err
	}
	if _, err := GetCurrentParty(); err != nil {
		return err
	}
	return nil
}

func GetCurrentPartyID() (int64, error) {
	var partyID int64
	err := DB.Get(&partyID, `SELECT party_id FROM app_config`)
	if err != nil {
		return 0, err
	}
	return partyID, nil
}

// момент времени последнего обновления текущей партии
//func GetCurrentPartyUpdatedAt() (time.Time, error) {
//	var tm string
//	err := DB.Get(&tm, `
//WITH q AS ( SELECT party_id FROM app_config )
//SELECT STRFTIME('%Y-%m-%d %H:%M:%f', tm) AS tm
//FROM measurement
//INNER JOIN product USING (product_id)
//WHERE party_id = (SELECT q.party_id FROM q)
//ORDER BY tm DESC
//LIMIT 1`)
//	if err == sql.ErrNoRows {
//		return time.Time{}, sql.ErrNoRows
//	}
//	return parseTime(tm), nil
//}

func GetCurrentParty() (Party, error) {
	partyID, err := GetCurrentPartyID()
	if err != nil {
		return Party{}, err
	}
	return GetParty(partyID)
}

func GetParty(partyID int64) (Party, error) {
	var party Party
	err := DB.Get(&party, `SELECT * FROM party WHERE party_id=?`, partyID)
	if err != nil {
		return Party{}, err
	}

	if err := DB.Select(&party.Products,
		`SELECT * FROM product_enumerated WHERE party_id=? ORDER BY place`,
		partyID); err != nil {
		return Party{}, err
	}
	return party, nil
}

func SetCurrentPartyValues(p PartyValues) error {
	partyID, err := GetCurrentPartyID()
	if err != nil {
		return err
	}
	p.PartyID = partyID
	if err := setPartyValues(p); err != nil {
		return err
	}
	return nil
}

func GetPartyValues1(partyID int64) (map[string]float64, error) {
	xs := make(map[string]float64)
	var values []struct {
		Key   string  `db:"key"`
		Value float64 `db:"value"`
	}
	if err := DB.Select(&values, `SELECT key, value FROM party_value WHERE party_id=?`, partyID); err != nil {
		return nil, err
	}

	for _, x := range values {
		xs[x.Key] = x.Value
	}
	return xs, nil
}

func GetPartyValues(partyID int64, party *PartyValues, filterSerial int64) error {

	err := DB.Get(party, `SELECT * FROM party WHERE party_id=?`, partyID)
	if err != nil {
		return err
	}

	party.Values, err = GetPartyValues1(partyID)
	if err != nil {
		return err
	}

	const (
		queryProducts1 = `SELECT * FROM product_enumerated WHERE party_id=? ORDER BY place`
		queryProducts2 = `
SELECT * FROM product_enumerated 
WHERE party_id=? AND serial = ?
ORDER BY place`
	)

	if filterSerial == -1 {
		err = DB.Select(&party.Products, queryProducts1, partyID)
	} else {
		err = DB.Select(&party.Products, queryProducts2, partyID, filterSerial)
	}

	if err != nil {
		return err
	}

	var xs []struct {
		ProductID int64       `db:"product_id"`
		Serial    int         `db:"serial"`
		Addr      modbus.Addr `db:"addr"`
		Key       string      `db:"key"`
		Value     float64     `db:"value"`
	}

	const (
		queryValues1 = `
SELECT product_id, serial, addr, key, value FROM product_value 
INNER JOIN product USING(product_id)
WHERE party_id= ?
ORDER BY created_at, created_order`
		queryValues2 = `
SELECT product_id, serial, addr, key, value FROM product_value 
INNER JOIN product USING(product_id)
WHERE party_id= ? AND serial = ?
ORDER BY created_at, created_order`
	)

	if filterSerial == -1 {
		err = DB.Select(&xs, queryValues1, partyID)
	} else {
		err = DB.Select(&xs, queryValues2, partyID, filterSerial)
	}

	if err != nil {
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
			panic("unexpected")
		}
		if p.Values == nil {
			p.Values = make(map[string]float64)
		}
		p.Values[x.Key] = x.Value
	}
	return nil
}

func SaveProductKefValue(productID int64, kef int, value float64) error {
	return SaveProductValue(productID, KeyCoefficient(kef), value)
}

func KeyCoefficient(k int) string {
	return fmt.Sprintf("K%02d", k)
}

func SaveProductValue(productID int64, key string, value float64) error {
	const q1 = `
INSERT INTO product_value
VALUES (?, ?, ?)
ON CONFLICT (product_id,key) DO UPDATE
    SET value = ?`
	_, err := DB.Exec(q1, productID, key, value, value)
	return merry.Appendf(err, "%s, %s: %v", q1, key, value)
}

func CopyParty(partyID int64) error {
	prevParty, err := GetParty(partyID)
	if err != nil {
		return err
	}

	newPartyID, err := createNewParty(prevParty.Name, prevParty.DeviceType, prevParty.ProductType)
	if err != nil {
		return err
	}

	if err := setAppConfigPartyID(newPartyID); err != nil {
		return err
	}

	if _, err := DB.Exec(`
INSERT INTO party_value
SELECT ?, key, value FROM product_value 
WHERE product_id = ?`, newPartyID, prevParty.PartyID); err != nil {
		return err
	}

	for i, p := range prevParty.Products {
		p.CreatedAt = time.Now()
		p.CreatedOrder = i
		p.PartyID = newPartyID

		r, err := DB.NamedExec(`
INSERT INTO product( party_id, addr, active, comport, created_at, created_order ) 
VALUES (:party_id, :addr, :active, :comport, :created_at, :created_order);`, p)

		if err != nil {
			return err
		}
		newProductID, err := getNewInsertedID(r)
		if err != nil {
			return err
		}

		if _, err := DB.Exec(`
INSERT INTO product_param
SELECT ?, param_addr, chart, series_active FROM product_param 
WHERE product_id = ?`, newProductID, p.ProductID); err != nil {
			return err
		}

		if _, err := DB.Exec(`
INSERT INTO product_value
SELECT ?, key, value FROM product_value 
WHERE product_id = ?`, newProductID, p.ProductID); err != nil {
			return err
		}
	}

	return nil
}

func SetProductParam(p ProductParam) error {
	_, err := DB.NamedExec(`
INSERT INTO product_param (product_id, param_addr, chart, series_active)
VALUES (:product_id, :param_addr, :chart, :series_active)
ON CONFLICT (product_id, param_addr) DO UPDATE SET series_active=:series_active,
                                                   chart=:chart`, p)
	return err
}

func SetNewCurrentParty(productsCount int) error {

	prevParty, err := GetCurrentParty()
	if err != nil {
		return err
	}
	t := time.Now()
	name := fmt.Sprintf("%d %s %s, %s",
		t.Day(), formatMonth(t), t.Format("2006"), t.Format("15:04"))

	var comport string
	if len(prevParty.Products) > 0 {
		comport = prevParty.Products[0].Comport
	}

	newPartyID, err := createNewParty(name, prevParty.DeviceType, prevParty.ProductType)
	if err != nil {
		return err
	}
	return setNewCurrentPartyProducts(newPartyID, productsCount, comport)
}

func GetActiveProducts() ([]Product, error) {

	var products []Product
	err := DB.Select(&products,
		`SELECT * FROM product_enumerated WHERE party_id = (SELECT party_id FROM app_config) AND active`)
	if err != nil {
		return nil, err
	}
	if len(products) == 0 {
		return nil, errNoInterrogateObjects
	}
	return products, nil
}

func setNewCurrentPartyProducts(newPartyID int64, productsCount int, comport string) error {
	for i := 0; i < productsCount; i++ {
		r, err := DB.Exec(
			`INSERT INTO product(party_id, addr, created_order, created_at, comport) VALUES (?, ?, ?, ?, ?);`,
			newPartyID, i+1, i+1, time.Now().Add(time.Second*time.Duration(i)), comport)
		if err != nil {
			return err
		}
		if _, err = getNewInsertedID(r); err != nil {
			return err
		}
	}
	if err := setAppConfigPartyID(newPartyID); err != nil {
		return err
	}
	return nil
}

func UpdateProduct(p Product) error {
	_, err := DB.NamedExec(`
UPDATE product
SET addr         = :addr,
    serial       = :serial,
    comport      = :comport,
    created_at   = :created_at,
    created_order=:created_order
WHERE product_id = :product_id`, p)
	return err
}

func AddNewProduct(order int) (int64, error) {
	party, err := GetCurrentParty()
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
	r, err := DB.Exec(
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

func DeleteParty(partyID int64) error {

	currentPartyID, err := GetCurrentPartyID()
	if err != nil {
		return err
	}

	if currentPartyID == partyID {
		var newCurrentPartyID int64
		const q1 = `
SELECT party_id 
FROM party 
WHERE party_id != (SELECT party_id FROM app_config WHERE id=1)
ORDER BY created_at  DESC 
LIMIT 1`
		if err := DB.Get(&newCurrentPartyID, q1); err != nil {
			return err
		}

		if _, err := DB.Exec(`UPDATE app_config SET party_id = ? WHERE id=1`, newCurrentPartyID); err != nil {
			return err
		}

	}

	if _, err := DB.Exec(`DELETE FROM party WHERE party_id=?`, partyID); err != nil {
		return err
	}
	return nil
}

func DeleteProductKey(productID int64, key string) error {
	const q1 = `DELETE FROM product_value WHERE product_id = ? AND key = ?`
	_, err := DB.Exec(q1, productID, key)
	return merry.Appendf(err, "%s, %s", q1, key)
}

func setAppConfigPartyID(partyID int64) error {
	_, err := DB.Exec(`UPDATE app_config SET party_id=? WHERE id=1`, partyID)
	return err
}

func createNewParty(name, deviceType, productType string) (int64, error) {
	r, err := DB.Exec(`INSERT INTO party (created_at, name, device_type, product_type) VALUES (?,?,?,?)`,
		time.Now(), name, deviceType, productType)
	if err != nil {
		return 0, err
	}
	n, err := r.RowsAffected()
	if err != nil {
		return 0, err
	}
	if n != 1 {
		return 0, merry.Errorf("excpected 1 rows affected, got %d", n)
	}
	return getNewInsertedID(r)
}

func setPartyValues(p PartyValues) error {

	const q1 = `UPDATE party SET name=?, device_type = ?, product_type=? WHERE party_id=?`
	if _, err := DB.Exec(q1, p.Name, p.DeviceType, p.ProductType, p.PartyID); err != nil {
		return err
	}
	if _, err := DB.Exec(`DELETE FROM product WHERE party_id=?`, p.PartyID); err != nil {
		return err
	}
	if _, err := DB.Exec(`DELETE FROM party_value WHERE party_id=?`, p.PartyID); err != nil {
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
		if _, err := DB.Exec(`INSERT INTO party_value(party_id, key, value) VALUES ` + sqlStr); err != nil {
			return err
		}
	}

	sqlStr = ""
	for i := range p.Products {
		var err error
		p.Products[i].ProductID, err = AddNewProduct(i)
		if err != nil {
			return err
		}
		p := p.Products[i]

		if _, err = DB.Exec(`UPDATE product SET serial = ? WHERE product_id=?`, p.Serial, p.ProductID); err != nil {
			return err
		}

		_, _ = DB.Exec(`UPDATE product SET addr = ? WHERE product_id=?`, p.Addr, p.ProductID)

		for k, v := range p.Values {
			if len(sqlStr) > 0 {
				sqlStr += ","
			}
			sqlStr += fmt.Sprintf("(%d, '%s', %v)", p.ProductID, k, v)
		}
	}
	if len(sqlStr) > 0 {
		if _, err := DB.Exec(`INSERT INTO product_value(product_id, key, value) VALUES ` + sqlStr); err != nil {
			return err
		}
	}
	return nil
}
