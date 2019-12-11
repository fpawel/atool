package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/comm/modbus"
	"github.com/jmoiron/sqlx"
	"strconv"
	"strings"
	"time"
)

//go:generate go run github.com/fpawel/gotools/cmd/sqlstr/...

const TimeLayout = "2006-01-02 15:04:05.000"

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
	if err != nil {
		return time.Time{}, err
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

func GetCurrentPartyChart(db *sqlx.DB, paramAddresses []int) ([]Measurement, error) {
	var xs []struct {
		Tm        string  `db:"tm"`
		ProductID int64   `db:"product_id"`
		ParamAddr int     `db:"param_addr"`
		Value     float64 `db:"value"`
	}

	var sQs []string
	for _, n := range paramAddresses {
		sQs = append(sQs, strconv.Itoa(n))
	}
	sQ := fmt.Sprintf("(%s)", strings.Join(sQs, ","))

	err := db.Select(&xs, `
WITH xs AS ( SELECT product_id FROM product WHERE party_id = (SELECT party_id FROM app_config))
SELECT STRFTIME('%Y-%m-%d %H:%M:%f', tm) AS tm,
       product_id, param_addr, value
FROM measurement
WHERE product_id IN (SELECT * FROM xs) AND param_addr IN `+sQ)
	if err != nil {
		return nil, err
	}
	var r []Measurement
	for _, x := range xs {
		r = append(r, Measurement{
			Time:      parseTime(x.Tm),
			ProductID: x.ProductID,
			ParamAddr: x.ParamAddr,
			Value:     x.Value,
		})
	}
	return r, nil
}

func GetParty(db *sqlx.DB, partyID int64) (Party, error) {
	var party Party
	err := db.Get(&party, `SELECT * FROM party WHERE party_id=?`, partyID)
	if err != nil {
		return Party{}, err
	}
	if party.Products, err = ListProducts(db, party.PartyID); err != nil {
		return Party{}, err
	}
	return party, nil
}

func CopyCurrentParty(db *sqlx.DB) error {
	p, err := GetCurrentParty(db)
	if err != nil {
		return err
	}

	newPartyID, err := createNewParty(db, p.Name)
	if err != nil {
		return err
	}

	for _, p := range p.Products {
		p.PartyID = newPartyID
		r, err := db.NamedExec(`
INSERT INTO product( party_id, addr, device, active, comport ) 
VALUES (:party_id, :addr, :device, :active, :comport);`, p)
		if err != nil {
			return err
		}
		newProductID, err := getNewInsertedID(r)
		if err != nil {
			return err
		}
		var pp []ProductParam
		if err := db.Select(&pp, `
SELECT product_id, param_addr, chart, series_active FROM product_param 
INNER JOIN product USING (product_id)
WHERE product_id = ?`, p.ProductID); err != nil {
			return err
		}
		for _, p := range pp {
			p.ProductID = newProductID
			if err := SetProductParam(db, p); err != nil {
				return err
			}
		}
	}

	if err := setAppConfigPartyID(db, newPartyID); err != nil {
		return err
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

func CreateNewParty(ctx context.Context, db *sqlx.DB, productsCount int, name string) error {
	newPartyID, err := createNewParty(db, name)
	if err != nil {
		return err
	}
	for i := 0; i < productsCount; i++ {
		r, err := db.ExecContext(ctx,
			`INSERT INTO product(party_id, addr) VALUES (?, ?);`,
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

func setAppConfigPartyID(db *sqlx.DB, partyID int64) error {
	_, err := db.Exec(`UPDATE app_config SET party_id=? WHERE id=1`, partyID)
	return err
}

func createNewParty(db *sqlx.DB, name string) (int64, error) {
	r, err := db.Exec(`INSERT INTO party (created_at, name) VALUES (?,?)`, time.Now(), name)
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

func ListProducts(db *sqlx.DB, partyID int64) (products []Product, err error) {
	err = db.Select(&products, `SELECT * FROM product WHERE party_id=? ORDER BY product_id`, partyID)
	return
}

func AddNewProduct(db *sqlx.DB) error {
	party, err := GetCurrentParty(db)
	if err != nil {
		return err
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
		`INSERT INTO product( party_id, addr) VALUES (?,?)`,
		party.PartyID, addr, time.Now())
	if err != nil {
		return err
	}

	if _, err = getNewInsertedID(r); err != nil {
		return err
	}
	return nil
}

func getNewInsertedID(r sql.Result) (int64, error) {
	id, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}
	if id <= 0 {
		return 0, errors.New("was not inserted")
	}
	return id, nil
}

func parseTime(sqlStr string) time.Time {
	t, err := time.ParseInLocation(TimeLayout, sqlStr, time.Now().Location())
	if err != nil {
		panic(err)
	}
	return t
}
