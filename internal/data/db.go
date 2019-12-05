package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/comm/modbus"
	"github.com/jmoiron/sqlx"
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
	if _, err := GetCurrentParty(context.Background(), db); err != nil {
		return nil, err
	}
	return db, nil
}

func OpenAppConfig(db *sqlx.DB, ctx context.Context) (AppConfig, error) {
	var c AppConfig

	if err := db.GetContext(ctx, &c.LogComport, `SELECT log_comport FROM app_config WHERE id=1`); err != nil {
		return c, merry.Append(err, "get config from db")
	}
	if err := ListHardware(db, ctx, &c.Hardware); err != nil {
		return c, err
	}
	return c, nil
}

func SaveAppConfig(db *sqlx.DB, ctx context.Context, c AppConfig) error {
	if _, err := db.Exec(`UPDATE app_config SET log_comport=? WHERE id=1`, c.LogComport); err != nil {
		return merry.Append(err, "UPDATE app_config SET log_comport=?")
	}
	return SaveHardware(db, ctx, c.Hardware)
}

func SaveHardware(db *sqlx.DB, ctx context.Context, hardware []Hardware) error {

	if _, err := db.ExecContext(ctx, `DELETE FROM hardware WHERE TRUE`); err != nil {
		return merry.Append(err, "save config: DELETE FROM hardware WHERE TRUE")
	}

	if len(hardware) == 0 {
		hardware = []Hardware{
			{
				Device:             "default",
				Baud:               9600,
				TimeoutGetResponse: time.Second,
				TimeoutEndResponse: 50 * time.Millisecond,
			},
		}
	}

	for _, device := range hardware {
		if _, err := db.NamedExecContext(ctx, `
INSERT INTO hardware(device, baud, timeout_get_responses, timeout_end_response, pause, max_attempts_read)
VALUES (:device, :baud, :timeout_get_responses, :timeout_end_response, :pause, :max_attempts_read)`, device); err != nil {
			return merry.Appendf(err, "save config: INSERT INTO hardware: %+v", device)
		}
		if len(device.Params) == 0 {
			device.Params = []Param{{
				ParamAddr: 0,
				Format:    "bcd",
				SizeRead:  2,
			}}
		}
		for _, param := range device.Params {
			param.Device = device.Device
			if _, err := db.NamedExecContext(ctx, `
INSERT INTO param(device, param_addr, format, size_read) 
VALUES (:device, :param_addr, :format, :size_read)`, param); err != nil {
				return merry.Appendf(err, "save config: INSERT INTO param: device: %+v: param: %+v")
			}
		}
	}
	return nil
}

func GetCurrentPartyID(ctx context.Context, db *sqlx.DB) (int64, error) {
	var partyID int64
	err := db.GetContext(ctx, &partyID, `SELECT party_id FROM app_config`)
	if err != nil {
		return 0, err
	}
	return partyID, nil
}

func GetCurrentParty(ctx context.Context, db *sqlx.DB) (Party, error) {
	partyID, err := GetCurrentPartyID(ctx, db)
	if err != nil {
		return Party{}, err
	}
	return GetParty(ctx, db, partyID)
}

func GetLastMeasurementTime(db *sqlx.DB) (time.Time, error) {
	var s string
	err := db.Select(&s, `
SELECT STRFTIME('%Y-%m-%d %H:%M:%f', tm) AS tm
FROM measurement
ORDER BY tm DESC
LIMIT 1`)
	if err != nil {
		return time.Time{}, err
	}
	return parseTime(s), nil
}

func GetCurrentPartyChart(db *sqlx.DB) ([]Measurement, error) {
	var xs []struct {
		Tm        string     `db:"tm"`
		ProductID int64      `db:"product_id"`
		ParamAddr modbus.Var `db:"param_addr"`
		Value     float64    `db:"value"`
	}

	err := db.Select(&xs, `
WITH xs AS (
    SELECT product_id FROM product WHERE party_id = (SELECT party_id FROM app_config)
)
SELECT STRFTIME('%Y-%m-%d %H:%M:%f', tm) AS tm,
       product_id, param_addr, value
FROM measurement
WHERE product_id IN (SELECT * FROM xs)`)
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

func parseTime(sqlStr string) time.Time {
	t, err := time.ParseInLocation(TimeLayout, sqlStr, time.Now().Location())
	if err != nil {
		panic(err)
	}
	return t
}

func GetParty(ctx context.Context, db *sqlx.DB, partyID int64) (Party, error) {
	var party Party
	err := db.GetContext(ctx, &party, `SELECT * FROM party WHERE party_id=?`, partyID)
	if err != nil {
		return Party{}, err
	}
	if party.Products, err = ListProducts(ctx, db, party.PartyID); err != nil {
		return Party{}, err
	}

	if err := db.SelectContext(ctx, &party.ParamAddresses, `
SELECT DISTINCT param_addr
FROM product
         INNER JOIN param USING (device)
WHERE party_id = ?
ORDER BY param_addr`, partyID, partyID); err != nil {
		return Party{}, err
	}
	return party, err
}

func CopyCurrentParty(ctx context.Context, db *sqlx.DB) error {
	p, err := GetCurrentParty(ctx, db)
	if err != nil {
		return err
	}

	newPartyID, err := createNewParty(db)
	if err != nil {
		return err
	}

	for _, p := range p.Products {
		p.PartyID = newPartyID
		r, err := db.ExecContext(ctx, `
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
SELECT * FROM product_param 
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

func CreateNewParty(ctx context.Context, db *sqlx.DB, productsCount int) error {
	newPartyID, err := createNewParty(db)
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
	_, err = db.ExecContext(ctx, `UPDATE app_config SET party_id=? WHERE id=1`, newPartyID)
	if err != nil {
		return err
	}
	return nil
}

func createNewParty(db *sqlx.DB) (int64, error) {
	r, err := db.Exec(`INSERT INTO party (created_at) VALUES (?)`, time.Now())
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

func ListProducts(ctx context.Context, db *sqlx.DB, partyID int64) (products []Product, err error) {
	err = db.SelectContext(ctx, &products, `SELECT * FROM product WHERE party_id=? ORDER BY product_id`, partyID)
	return
}

func AddNewProduct(ctx context.Context, db *sqlx.DB) error {
	party, err := GetCurrentParty(ctx, db)
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
	r, err := db.ExecContext(ctx,
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

func ListHardware(db *sqlx.DB, ctx context.Context, hardware *[]Hardware) error {
	if err := db.SelectContext(ctx, hardware, `SELECT * FROM hardware`); err != nil {
		return merry.Append(err, "SELECT * FROM hardware")
	}
	for i := range *hardware {
		p := &(*hardware)[i]
		if err := db.SelectContext(ctx, &p.Params, `SELECT * FROM param WHERE device=?`, p.Device); err != nil {
			return merry.Appendf(err, "SELECT * FROM param WHERE device=? | %s", p.Device)
		}
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
