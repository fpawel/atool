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
	if _, err := GetCurrentParty(context.Background(), db); err != nil {
		return nil, err
	}
	return db, nil
}

type Party struct {
	PartyInfo
	Products []Product    `db:"-"`
	Params   []modbus.Var `db:"-"`
}

type PartyInfo struct {
	PartyID   int64     `db:"party_id"`
	CreatedAt time.Time `db:"created_at"`
}

type Product struct {
	ProductID      int64       `db:"product_id"`
	PartyID        int64       `db:"party_id"`
	PartyCreatedAt time.Time   `db:"created_at"`
	Comport        string      `db:"comport"`
	Addr           modbus.Addr `db:"addr"`
	Checked        bool        `db:"checked"`
	Device         string      `db:"device"`
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

	if err := db.SelectContext(ctx, &party.Params, `
SELECT DISTINCT var
FROM product
         INNER JOIN param USING (device)
WHERE party_id = ?
UNION
SELECT DISTINCT var
FROM measurement
         INNER JOIN product USING (product_id)
WHERE party_id = ?
ORDER BY var`, partyID, partyID); err != nil {
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
