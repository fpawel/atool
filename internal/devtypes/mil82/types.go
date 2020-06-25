package mil82

import (
	"github.com/fpawel/atool/internal/config/devicecfg"
)

type productType struct {
	Name    string
	Gas     string
	Scale0  float64
	Scale   float64
	Kef     map[devicecfg.Coefficient]float64
	TempMin float64
	TempMax float64
	Index   int
}

var (
	prodTypesList = []productType{
		{
			Name:    "00.00",
			Gas:     "CO2",
			Scale:   4,
			TempMin: -40,
			TempMax: 80,
			Index:   0,
			Kef: KefValueMap{
				4:  5,
				14: 0.1,
				35: 5,
				45: 60,
				50: 0,
			},
		},
		{
			Name:    "00.01",
			Gas:     "CO2",
			Scale:   10,
			TempMin: -40,
			TempMax: 80,
			Index:   1,
			Kef: KefValueMap{
				4:  5,
				14: 0.1,
				35: 5,
				45: 60,
				50: 0,
			},
		},
		{
			Name:    "00.02",
			Gas:     "CO2",
			Scale:   20,
			TempMin: -40,
			TempMax: 80,
			Index:   2,
			Kef: KefValueMap{
				4:  5,
				14: 0.1,
				35: 5,
				45: 60,
				50: 0,
			},
		},
		{
			Name:    "01.00",
			Gas:     "CH4",
			Scale:   100,
			TempMin: -40,
			TempMax: 80,
			Index:   3,
			Kef: KefValueMap{
				4:  7.5,
				14: 0.5,
				35: 5,
				45: 60,
				50: 0,
			},
		},
		{
			Name:    "01.01",
			Gas:     "CH4",
			Scale:   100,
			TempMin: -60,
			TempMax: 60,
			Index:   4,
			Kef: KefValueMap{
				4:  7.5,
				14: 0.5,
				35: 5,
				45: 60,
				50: 0,
			},
		},
		{
			Name:    "02.00",
			Gas:     "C3H8",
			Scale:   50,
			TempMin: -40,
			TempMax: 60,
			Index:   5,
			Kef: KefValueMap{
				4:  12.5,
				14: 0.5,
				35: 5,
				45: 30,
				50: 0,
			},
		},
		{
			Name:    "02.01",
			Gas:     "C3H8",
			Scale:   50,
			TempMin: -60,
			TempMax: 60,
			Index:   6,
			Kef: KefValueMap{
				4:  12.5,
				14: 0.5,
				35: 5,
				45: 30,
				50: 0,
			},
		},
		{
			Name:    "03.00",
			Gas:     "C3H8",
			Scale:   100,
			TempMin: -40,
			TempMax: 60,
			Index:   7,
			Kef: KefValueMap{
				4:  12.5,
				14: 0.5,
				35: 5,
				45: 30,
				50: 0,
			},
		},
		{
			Name:    "03.01",
			Gas:     "C3H8",
			Scale:   100,
			TempMin: -60,
			TempMax: 60,
			Index:   8,
			Kef: KefValueMap{
				4:  12.5,
				14: 0.5,
				35: 5,
				45: 30,
				50: 0,
			},
		},
		{
			Name:    "04.00",
			Gas:     "CH4",
			Scale:   100,
			TempMin: -60,
			TempMax: 80,
			Index:   9,
			Kef: KefValueMap{
				4:  7.5,
				14: 0.5,
				35: 5,
				45: 60,
				50: 0,
			},
		},
		{
			Name:    "05.00",
			Gas:     "C6H14",
			Scale0:  5,
			Scale:   50,
			TempMin: 15,
			TempMax: 80,
			Index:   10,
			Kef: KefValueMap{
				4:  1,
				14: 30,
				35: 5,
				45: 30,
				50: 0,
			},
		},
		{
			Name:    "10.00",
			Gas:     "CO2",
			Scale:   4,
			TempMin: -40,
			TempMax: 80,
			Index:   11,
			Kef: KefValueMap{
				4:  1,
				14: 30,
				35: 1,
				45: 30,
				50: 1,
			},
		},
		{
			Name:    "10.01",
			Gas:     "CO2",
			Scale:   10,
			TempMin: -40,
			TempMax: 80,
			Index:   12,
			Kef: KefValueMap{
				4:  1,
				14: 30,
				35: 1,
				45: 30,
				50: 1,
			},
		},
		{
			Name:    "10.02",
			Gas:     "CO2",
			Scale:   20,
			TempMin: -40,
			TempMax: 80,
			Index:   13,
			Kef: KefValueMap{
				4:  1,
				14: 30,
				35: 1,
				45: 30,
				50: 1,
			},
		},
		{
			Name:    "10.03",
			Gas:     "CO2",
			Scale:   4,
			TempMin: -60,
			TempMax: 80,
			Index:   14,
			Kef: KefValueMap{
				4:  1,
				14: 30,
				35: 1,
				45: 30,
				50: 1,
			},
		},
		{
			Name:    "10.04",
			Gas:     "CO2",
			Scale:   10,
			TempMin: -60,
			TempMax: 80,
			Index:   15,
			Kef: KefValueMap{
				4:  1,
				14: 30,
				35: 1,
				45: 30,
				50: 1,
			},
		},
		{
			Name:    "10.05",
			Gas:     "CO2",
			Scale:   20,
			TempMin: -60,
			TempMax: 80,
			Index:   16,
			Kef: KefValueMap{
				4:  1,
				14: 30,
				35: 1,
				45: 30,
				50: 1,
			},
		},
		{
			Name:    "11.00",
			Gas:     "CH4",
			Scale:   100,
			TempMin: -40,
			TempMax: 80,
			Index:   17,
			Kef: KefValueMap{
				4:  1,
				14: 30,
				35: 1,
				45: 30,
				50: 1,
			},
		},
		{
			Name:    "11.01",
			Gas:     "CH4",
			Scale:   100,
			TempMin: -60,
			TempMax: 80,
			Index:   18,
			Kef: KefValueMap{
				4:  1,
				14: 30,
				35: 1,
				45: 30,
				50: 1,
			},
		},
		{
			Name:    "13.00",
			Gas:     "C3H8",
			Scale:   100,
			TempMin: -40,
			TempMax: 80,
			Index:   19,
			Kef: KefValueMap{
				4:  1,
				14: 30,
				35: 1,
				45: 30,
				50: 1,
			},
		},
		{
			Name:    "13.01",
			Gas:     "C3H8",
			Scale:   100,
			TempMin: -60,
			TempMax: 80,
			Index:   20,
			Kef: KefValueMap{
				4:  1,
				14: 30,
				35: 1,
				45: 30,
				50: 1,
			},
		},
		{
			Name:    "14.00",
			Gas:     "CH4",
			Scale:   100,
			TempMin: -60,
			TempMax: 80,
			Index:   21,
			Kef: KefValueMap{
				4:  1,
				14: 30,
				35: 1,
				45: 30,
				50: 1,
			},
		},
		{
			Name:    "16.00",
			Gas:     "C3H8",
			Scale:   100,
			TempMin: -60,
			TempMax: 80,
			Index:   22,
			Kef: KefValueMap{
				4:  1,
				14: 30,
				35: 1,
				45: 30,
				50: 1,
			},
		},
	}

	prodTypes, prodTypeNames = func() (m map[string]productType, xs []string) {
		m = map[string]productType{}
		for i := range prodTypesList {
			t := &prodTypesList[i]
			t.Index = i + 1
			m[t.Name] = *t
			xs = append(xs, t.Name)
		}
		return
	}()
)
