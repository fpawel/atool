package data

import (
	"github.com/fpawel/comm/modbus"
	"time"
)

type AppConfig struct {
	LogComport bool       `yaml:"log_comport"`
	Hardware   []Hardware `yaml:"hardware"`
}

type Hardware struct {
	Device             string        `db:"device" yaml:"device"`
	Baud               int           `db:"baud" yaml:"baud"`
	Pause              time.Duration `db:"pause" yaml:"pause"`
	TimeoutGetResponse time.Duration `db:"timeout_get_responses" yaml:"timeout_get_responses"`
	TimeoutEndResponse time.Duration `db:"timeout_end_response" yaml:"timeout_end_response"`
	MaxAttemptsRead    int           `db:"max_attempts_read" yaml:"max_attempts_read"`
	Params             []Param       `db:"-" yaml:"params"`
}

type Param struct {
	Device    string     `db:"device" yaml:"-"`
	ParamAddr modbus.Var `db:"param_addr" yaml:"param_addr"`
	Format    string     `db:"format" yaml:"format"`
	SizeRead  uint16     `db:"size_read" yaml:"size_read"`
	ReadOnce  bool       `db:"read_once" yaml:"read_once"`
}

type PartyInfo struct {
	PartyID   int64     `db:"party_id"`
	CreatedAt time.Time `db:"created_at"`
}

type Party struct {
	PartyInfo
	Products       []Product
	ParamAddresses []modbus.Var
}

type Product struct {
	ProductID      int64       `db:"product_id"`
	PartyID        int64       `db:"party_id"`
	PartyCreatedAt time.Time   `db:"created_at"`
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
	ParamAddr modbus.Var
	Value     float64
}
