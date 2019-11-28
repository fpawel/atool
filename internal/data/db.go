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

//func OpenDev() (*sqlx.DB,error){
//	return Open(filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "fpawel", "daf", "build", "daf.sqlite"))
//}

//func OpenProd() (*sqlx.DB,error){
//	return Open(filepath.Join(filepath.Dir(os.Args[0]), "atool.sqlite"))
//}

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
INSERT INTO param(device, param_addr, format, size_read, read_once) 
VALUES (:device, :param_addr, :format, :size_read, :read_once)`, param); err != nil {
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

func CreateNewParty(ctx context.Context, db *sqlx.DB, productsCount int) error {
	r, err := db.ExecContext(ctx, `INSERT INTO party (created_at) VALUES (?)`, time.Now())
	if err != nil {
		return err
	}
	n, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("excpected 1 rows affected, got %d", n)
	}
	newPartyID, err := getNewInsertedID(r)
	if err != nil {
		return err
	}
	for i := 0; i < productsCount; i++ {
		if r, err = db.ExecContext(ctx,
			`INSERT INTO product(party_id, addr) VALUES (?, ?);`,
			newPartyID, i+1, i+1, time.Now().Add(time.Second*time.Duration(i))); err != nil {
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
