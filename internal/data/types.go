package data

import (
	"fmt"
	"github.com/fpawel/comm/modbus"
	"time"
)

type PartyInfo struct {
	PartyID     int64     `db:"party_id"`
	CreatedAt   time.Time `db:"created_at"`
	Name        string    `db:"name"`
	ProductType string    `db:"product_type"`
}

type Party struct {
	PartyInfo
	Products []Product
}

type Product struct {
	ProductID    int64       `db:"product_id"`
	PartyID      int64       `db:"party_id"`
	CreatedAt    time.Time   `db:"created_at"`
	CreatedOrder int         `db:"created_order"`
	Serial       int         `db:"serial"`
	Comport      string      `db:"comport"`
	Addr         modbus.Addr `db:"addr"`
	Device       string      `db:"device"`
	Active       bool        `db:"active"`
}

func (p Product) String() string {
	return fmt.Sprintf("%s,%s,адр=%d,сер№=%d,id=%d", p.Comport, p.Device, p.Addr, p.Serial, p.ProductID)
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

type Values map[string]float64

type PartyValues struct {
	Values   Values          `yaml:"values"`
	Products []ProductValues `yaml:"products"`
}

type ProductValues struct {
	ProductID int64              `yaml:"product_id"`
	Place     int                `yaml:"place"`
	Serial    int                `yaml:"serial"`
	Values    map[string]float64 `yaml:"values"`
}
