package data

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
)

func LoadFile(db *sqlx.DB, filename string) error {
	jsonData, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	var x PartyValues

	if err = json.Unmarshal(jsonData, &x); err != nil {
		return err
	}
	return SetCurrentPartyValues(db, x)
}
