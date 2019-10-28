package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	if _, err := GetLastParty(context.Background(), db); err != nil {
		return nil, err
	}
	return db, nil
}

type Party struct {
	PartyID   int64     `db:"party_id"`
	CreatedAt time.Time `db:"created_at"`
}

type Product struct {
	CreatedAt  time.Time   `db:"created_at"`
	ProductID  int64       `db:"product_id"`
	PartyID    int64       `db:"party_id"`
	Serial     int         `db:"serial"`
	Port       int         `db:"port"`
	Addr       modbus.Addr `db:"addr"`
	Checked    bool        `db:"checked"`
	DeviceName string      `db:"device_name"`
}

func GetLastParty(ctx context.Context, db *sqlx.DB) (party Party, err error) {
	if err = db.GetContext(ctx, &party, `SELECT * FROM last_party`); err == nil {
		return
	}
	if err != sql.ErrNoRows {
		return
	}
	if err := CreateNewParty(ctx, db); err != nil {
		return Party{}, err
	}
	err = db.GetContext(ctx, &party, `SELECT * FROM last_party`)
	return
}

func CreateNewParty(ctx context.Context, db *sqlx.DB) error {
	r, err := db.ExecContext(ctx, `INSERT INTO party DEFAULT VALUES`)
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
	if r, err = db.ExecContext(ctx, `INSERT INTO product(party_id, serial) VALUES (?, 1);`, newPartyID); err != nil {
		return err
	}
	if _, err = getNewInsertedID(r); err != nil {
		return err
	}
	return nil
}

func ListLastPartyProducts(ctx context.Context, db *sqlx.DB) ([]Product, error) {
	p, err := GetLastParty(ctx, db)
	if err != nil {
		return nil, err
	}
	return ListProducts(ctx, db, p.PartyID)
}

func ListProducts(ctx context.Context, db *sqlx.DB, partyID int64) (products []Product, err error) {
	err = db.SelectContext(ctx, &products, `SELECT * FROM product WHERE party_id=? ORDER BY created_at`, partyID)
	return
}

func AddNewProduct(ctx context.Context, db *sqlx.DB) error {
	xs, err := ListLastPartyProducts(ctx, db)
	if err != nil {
		return err
	}
	addresses := make(map[modbus.Addr]struct{})
	serials := make(map[int]struct{})
	for _, x := range xs {
		addresses[x.Addr] = struct{}{}
		serials[x.Serial] = struct{}{}
	}
	serial, addr := 1, modbus.Addr(1)
	for ; addr <= modbus.Addr(255); addr++ {
		if _, f := addresses[addr]; !f {
			break
		}
	}
	for serial = 1; serial < 100500; serial++ {
		if _, f := serials[serial]; !f {
			break
		}
	}
	r, err := db.ExecContext(ctx,
		`INSERT INTO product( party_id, serial, addr) VALUES ( (SELECT party_id FROM last_party),?,?)`,
		serial, addr)
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
