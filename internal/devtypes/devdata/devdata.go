package devdata

import (
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"log"
	"sort"
)

type Device struct {
	DataSections        DataSections
	GetCalcSectionsFunc func(data.PartyValues, *CalcSections) error
	ProductTypes        map[string]interface{}
	Config              devicecfg.Device
	Name                string
}

type DataSections []DataSection

type DataSection struct {
	Name   string
	Params []DataParam
}

type DataParam struct {
	Key, Name string
}

func (x Device) ProductTypeIndex(productType string) int {
	t, ok := x.ProductTypes[productType]
	if !ok {
		log.Fatalf("%q: %q: product type not defined", x.Name, productType)
	}
	m, ok := t.(map[string]interface{})
	if !ok {
		log.Fatalf("%q: %q: wrong data type: %+v", x.Name, productType, m)
	}
	i, ok := m["index"]
	if !ok {
		log.Fatalf("%q: %q: index not defined: %+v", x.Name, productType, m)
	}
	n, ok := i.(float64)
	if !ok {
		log.Fatalf("%q: %q: index: wrong type %T, int excpected: %+v", x.Name, productType, n, m)
	}
	if float64(int(n)) != n {
		log.Fatalf("%q: %q: index: wrong value %v, int excpected: %+v", x.Name, productType, n, m)
	}
	return int(n)
}

func (x Device) ListProductTypes() []string {
	r := make([]string, 0)

	for k := range x.ProductTypes {
		r = append(r, k)
	}
	sort.Slice(r, func(i, j int) bool {
		return x.ProductTypeIndex(r[i]) < x.ProductTypeIndex(r[j])
	})
	return r
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
