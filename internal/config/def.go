package config

import (
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/devtypes"
	"github.com/fpawel/hardware/gas"
	"github.com/fpawel/hardware/temp/ktx500"
	"time"
)

func defaultConfig() Config {

	hardware := Hardware{
		"default": devicecfg.Device{
			Baud:               9600,
			TimeoutGetResponse: time.Second,
			TimeoutEndResponse: time.Millisecond * 50,
			MaxAttemptsRead:    0,
			Pause:              0,
			Params: []devicecfg.Params{
				{
					Format:    "bcd",
					ParamAddr: 0,
					Count:     1,
				},
			},
			NetAddr: devicecfg.NetAddr{
				Cmd:    5,
				Format: "bcd",
			},
			PartyParams: defaultPartyParams(),
			Coefficients: []devicecfg.Coefficients{
				{
					Range:  [2]int{0, 50},
					Format: "float_big_endian",
				},
			},
		},
	}

	for name, d := range devtypes.DeviceTypes {
		hardware[name] = d.Config
	}

	return Config{
		LogComm: false,

		FloatPrecision: 6,

		Hardware: Hardware{
			"default": devicecfg.Device{
				Baud:               9600,
				TimeoutGetResponse: time.Second,
				TimeoutEndResponse: time.Millisecond * 50,
				MaxAttemptsRead:    0,
				Pause:              0,
				Params: []devicecfg.Params{
					{
						Format:    "bcd",
						ParamAddr: 0,
						Count:     1,
					},
				},
				NetAddr: devicecfg.NetAddr{
					Cmd:    5,
					Format: "bcd",
				},
				PartyParams: defaultPartyParams(),
				Coefficients: []devicecfg.Coefficients{
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
