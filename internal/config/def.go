package config

import (
	"github.com/fpawel/hardware/gas"
	"github.com/fpawel/hardware/temp/ktx500"
	"time"
)

func defaultConfig() Config {
	return Config{
		LogComm: false,

		FloatPrecision: 6,

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
				NetAddr: NetAddr{
					Cmd:    5,
					Format: "bcd",
				},
				ProductTypes: []string{"00.01", "00.02"},
				PartyParams:  defaultPartyParams(),
				Coefficients: []Coefficients{
					{
						Range:  [2]int{0, 50},
						Format: "float_big_endian",
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
			BlowDuration:       5 * time.Minute,
		},
		Temperature: Temperature{
			Type:               T800,
			Comport:            "COM1",
			TimeoutGetResponse: time.Second,
			TimeoutEndResponse: time.Millisecond * 50,
			MaxAttemptsRead:    1,
			HoldDuration:       2 * time.Hour,
		},
		WarmSheets: Mil82WarmSheets{
			Enable: false,
			Addr:   99,
		},
		Ktx500:               ktx500.NewDefaultConfig(),
		InactiveCoefficients: make(map[int]struct{}),
	}
}

func defaultPartyParams() map[string]string {
	return map[string]string{
		"c1": "ПГС1",
		"c2": "ПГС2",
		"c3": "ПГС3",
		"c4": "ПГС2",
	}
}
