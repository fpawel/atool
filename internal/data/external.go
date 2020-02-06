package data

import (
	"encoding/json"
	"github.com/fpawel/comm/modbus"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
)

func LoadFile(db *sqlx.DB, filename string) error {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	var x externalParty

	if err = json.Unmarshal(b, &x); err != nil {
		return err
	}

	partyID, err := GetCurrentPartyID(db)
	if err != nil {
		return err
	}

	const q1 = `UPDATE party SET product_type=?, name=? WHERE party_id=?`
	if _, err := db.Exec(q1, x.ProductType, x.Name, partyID); err != nil {
		return err
	}

	if _, err := db.Exec(`DELETE FROM product WHERE party_id=?`, partyID); err != nil {
		return err
	}

	partyValues := PartyValues{
		Values: x.Values,
	}

	for i, p := range x.Products {
		r, err := db.Exec(
			`INSERT INTO product( party_id, serial, created_order) VALUES (?,?,?)`,
			partyID, p.Serial, i)
		if err != nil {
			return err
		}
		productID, err := getNewInsertedID(r)
		if err != nil {
			return err
		}

		_, _ = db.Exec(`UPDATE product SET addr=? WHERE product_id=?`, p.Addr, productID)

		partyValues.Products = append(partyValues.Products, ProductValues{
			ProductID: productID,
			Place:     i,
			Serial:    p.Serial,
			Values:    p.Values,
		})
	}

	if err := setPartyValues(db, partyID, partyValues); err != nil {
		return err
	}

	return nil
}

type externalParty struct {
	ProductType string
	Name        string
	Values      map[string]float64
	Products    []externalProduct
}

type externalProduct struct {
	Addr   modbus.Addr
	Serial int
	Values map[string]float64
}
