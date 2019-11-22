package data

import (
	"github.com/fpawel/comm/modbus"
	"time"
)

type Party struct {
	PartyInfo
	Products    []Product
	Vars        []modbus.Var
	ProductVars []ProductVar
}

type Product struct {
	ProductID      int64       `db:"product_id"`
	PartyID        int64       `db:"party_id"`
	PartyCreatedAt time.Time   `db:"created_at"`
	Comport        string      `db:"comport"`
	Addr           modbus.Addr `db:"addr"`
	Device         string      `db:"device"`
}

type ProductVar struct {
	DeviceVarID  int64      `db:"device_var_id"`
	ProductID    int64      `db:"product_id"`
	Var          modbus.Var `db:"var"`
	Chart        string     `db:"chart"`
	SeriesActive bool       `db:"series_active"`
	Read         bool       `db:"read"`
	SizeRead     int        `db:"size_read"`
}

type PartyInfo struct {
	PartyID   int64     `db:"party_id"`
	CreatedAt time.Time `db:"created_at"`
}
