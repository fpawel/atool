package ikds4

import (
	"github.com/fpawel/atool/internal/config/devicecfg"
)

type productType struct {
	Name  string
	Gas   string
	Scale float64
	Index int
	Kef   map[devicecfg.Kef]float64
	limD  float64
}

var (
	prodTypesList = []productType{
		{
			Name:  "CO2-2",
			Gas:   "CO2",
			Scale: 2,
			Index: 1,
			limD:  0.1,
			Kef:   KefValueMap{},
		},
		{
			Name:  "CO2-4",
			Gas:   "CO2",
			Scale: 4,
			Index: 2,
			limD:  0.25,
			Kef:   KefValueMap{},
		},
		{
			Name:  "CO2-10",
			Gas:   "CO2",
			Scale: 10,
			Index: 3,
			limD:  0.5,
			Kef:   KefValueMap{},
		},
		{
			Name:  "CH4-100",
			Gas:   "CH4",
			Scale: 100,
			Index: 4,
			Kef:   KefValueMap{},
		},
		{
			Name:  "CH4-100НКПР",
			Gas:   "CH4",
			Scale: 100,
			Index: 5,
			Kef:   KefValueMap{},
		},
		{
			Name:  "C3H8-100",
			Gas:   "C3H8",
			Scale: 100,
			Index: 6,
			Kef:   KefValueMap{},
		},
	}
	prodTypes, prodTypeNames = initProductTypes(prodTypesList)
)

func initProductTypes(prodTypesList []productType) (m map[string]productType, xs []string) {
	m = map[string]productType{}
	for i := range prodTypesList {
		t := &prodTypesList[i]
		t.Index = i + 1
		m[t.Name] = *t
		xs = append(xs, t.Name)
	}
	return
}
