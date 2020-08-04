package devdata

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"sort"
)

type Device struct {
	Name             string
	DataSections     DataSections
	ProductTypes     []string
	Config           Config
	PartyParams      []PartyParam
	Commands         []Cmd
	ProductTypesVars []ProductTypeVars
	InitParty        func() error
	Calc             func(data.PartyValues, *CalcSections) error
	Work             func(comm.Logger, context.Context) error
	OnReadProduct    func(comm.Logger, context.Context, Product) error
}

type Product struct {
	data.Product
	Party data.Party
}

type Config = devicecfg.Device

type ProductTypeVars struct {
	Names         []string
	ParamsRngList []devicecfg.Vars
}

type Cmd struct {
	Code modbus.DevCmd `yaml:"code"`
	Name string        `yaml:"name"`
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

func (d Device) VarName(paramAddr modbus.Var) string {
	for n, s := range d.Config.VarsNames {
		if n == paramAddr {
			return fmt.Sprintf("%d: %s", paramAddr, s)
		}
	}
	return fmt.Sprintf("%d", paramAddr)
}

func (d Device) VarsRng(prodType string) []devicecfg.Vars {
	xs := d.Config.Vars
	for _, y := range d.ProductTypesVars {
		for _, s := range y.Names {
			if s == prodType {
				xs = append(xs, y.ParamsRngList...)
			}
		}
	}
	return xs
}

func (d Device) BufferSize(prodType string) (r int) {
	for _, p := range d.VarsRng(prodType) {
		x := p.Var()*2 + p[1]*4
		if r < int(x) {
			r = int(x)
		}
	}
	return
}

func (d Device) Vars(prodType string) (ps []modbus.Var) {
	for _, p := range d.VarsRng(prodType) {
		for i := 0; i < p.Count(); i++ {
			ps = append(ps, p.Var()+modbus.Var(i)*2)
		}
	}
	sort.Slice(ps, func(i, j int) bool {
		return ps[i] < ps[j]
	})
	return
}

func (d Device) Validate() error {

	if err := d.Config.Validate(); err != nil {
		return err
	}

	for _, y := range d.ProductTypesVars {
		for _, s := range y.Names {
			m := make(map[modbus.Var]struct{})
			for _, x := range d.Vars(s) {
				if _, f := m[x]; f {
					return merry.Errorf(`дублирование адреса параметра %s: %d`, s, x)
				}
				m[x] = struct{}{}
			}
		}
	}
	return nil
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
