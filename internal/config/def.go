package config

import (
	"github.com/fpawel/hardware/gas"
	"github.com/fpawel/hardware/temp/ktx500"
	"time"
)

func defaultConfig() Config {
	return Config{
		LogComm:         false,
		BlowGas:         5 * time.Minute,
		HoldTemperature: 2 * time.Hour,
		FloatPrecision:  6,
		ProductTypes:    []string{"00.01", "00.02"},
		PartyParams:     defaultPartyParams(),
		ProductParams:   defaultProductParams(),
		Hardware: Hardware{
			"default": Device{
				Baud:               9600,
				TimeoutGetResponse: time.Second,
				TimeoutEndResponse: time.Millisecond * 50,
				MaxAttemptsRead:    0,
				Pause:              0,
				Params: []Params{
					{
						Format:    "bcd",
						ParamAddr: 0,
						Count:     1,
					},
				},
			},
		},
		Gas: Gas{
			Type:               gas.Mil82,
			Addr:               100,
			Comport:            "COM1",
			TimeoutGetResponse: time.Second,
			TimeoutEndResponse: time.Millisecond * 50,
			MaxAttemptsRead:    0,
		},
		Temperature: Temperature{
			Type:               T800,
			Comport:            "COM1",
			TimeoutGetResponse: time.Second,
			TimeoutEndResponse: time.Millisecond * 50,
			MaxAttemptsRead:    1,
		},
		Ktx500:               ktx500.NewDefaultConfig(),
		InactiveCoefficients: make(map[int]struct{}),
		Coefficients: []Coefficients{
			{
				Range:  [2]int{0, 50},
				Format: "float_big_endian",
			},
		},
	}
}

func defaultProductParams() ProductParams {
	type m = map[string]string
	return ProductParams{
		"Линеаризация": m{
			"lin1": "газ1",
			"lin2": "газ2",
			"lin3": "газ3",
			"lin4": "газ4",
			"lin5": "газ5",
		},
		"Пониженная температура, датчик температуры": m{
			"t1gas1t": "газ1",
			"t1gas2t": "газ2",
			"t1gas3t": "газ3",
		},
		"Пониженная температура, сигнал": m{
			"t1gas1v": "газ1",
			"t1gas2v": "газ2",
			"t1gas3v": "газ3",
		},
		"Нормальная температура, датчик температуры": m{
			"t2gas1t": "газ1",
			"t2gas2t": "газ2",
			"t2gas3t": "газ3",
		},
		"Нормальная температура, сигнал": m{
			"t2gas1v": "газ1",
			"t2gas2v": "газ2",
			"t2gas3v": "газ3",
		},
		"Снятие основной погрешности, концентрация": m{
			"с1": "1.ПГС1",
			"с2": "2.ПГС2",
			"с3": "3.ПГС3",
			"с4": "4.ПГС2",
			"с5": "5.ПГС1",
		},
	}
}

func defaultPartyParams() map[string]PartyParam {
	return map[string]PartyParam{
		"c1": {
			Name:     "ПГС1",
			Positive: true,
			Def:      0,
		},
		"c2": {
			Name:     "ПГС2",
			Positive: true,
			Def:      50,
		},
		"c3": {
			Name:     "ПГС3",
			Positive: true,
			Def:      100,
		},
	}
}
