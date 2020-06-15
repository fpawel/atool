package devdata

import (
	"context"
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/comm"
)

type Device struct {
	Name         string
	DataSections DataSections
	ProductTypes []string
	Config       devicecfg.Device
	PartyParams  []PartyParam
	InitParty    func() error
	Calc         func(data.PartyValues, *CalcSections) error
	Work         func(comm.Logger, context.Context) error
}

type PartyParam struct {
	Key, Name  string
	ValuesList []string
}

type DataSections []DataSection

type DataSection struct {
	Name   string
	Params []DataParam
}

type DataParam struct {
	Key, Name string
}

func (xs DataSections) Keys() map[string]struct{} {
	r := map[string]struct{}{}
	for _, x := range xs {
		for _, p := range x.Params {
			r[p.Key] = struct{}{}
		}
	}
	return r
}

func (xs DataSections) HasKey(key string) bool {
	for _, x := range xs {
		for _, p := range x.Params {
			if p.Key == key {
				return true
			}
		}
	}
	return false
}

type CalcSections []*CalcSection
type CalcParam = apitypes.CalcParam
type CalcValue = apitypes.CalcValue
type CalcSection = apitypes.CalcSection

func AddSect(x *CalcSections, name string) *CalcSection {
	c := &CalcSection{Name: name}
	*x = append(*x, c)
	return c
}

func AddParam(x *CalcSection, name string) *CalcParam {
	v := &CalcParam{Name: name}
	x.Params = append(x.Params, v)
	return v
}

func AddValue(x *CalcParam) *CalcValue {
	v := new(CalcValue)
	x.Values = append(x.Values, v)
	return v
}
