package devdata

import (
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/devdata/devcalc"
)

var Devices = make(map[string]Device)

type Device struct {
	DataSections DataSections
	Calc         func(party data.PartyValues, calc *devcalc.CalcSections) error
}
