package luadata

import "github.com/fpawel/atool/internal/data"

type Device struct {
	Data DataSections
	Calc func(party data.PartyValues, calc *CalcSections) string
}

func (x *Device) AddDataSection(name string) *DataSection {
	s := &DataSection{Name: name}
	x.Data = append(x.Data, s)
	return s
}
