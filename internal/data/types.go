package data

import (
	"github.com/fpawel/comm/modbus"
	"time"
)

type PartyInfo struct {
	PartyID   int64     `db:"party_id"`
	CreatedAt time.Time `db:"created_at"`
	Name      string    `db:"name"`
}

type Party struct {
	PartyInfo
	Products []Product
}

type Product struct {
	ProductID      int64       `db:"product_id"`
	PartyID        int64       `db:"party_id"`
	PartyCreatedAt time.Time   `db:"created_at"`
	Serial         int         `db:"serial"`
	Comport        string      `db:"comport"`
	Addr           modbus.Addr `db:"addr"`
	Device         string      `db:"device"`
	Active         bool        `db:"active"`
}

type ProductParam struct {
	ProductID    int64      `db:"product_id"`
	ParamAddr    modbus.Var `db:"param_addr"`
	Chart        string     `db:"chart"`
	SeriesActive bool       `db:"series_active"`
}

type Measurement struct {
	Time      time.Time
	ProductID int64
	ParamAddr int
	Value     float64
}
