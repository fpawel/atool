package config

import (
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/hardware/gas"
	"github.com/fpawel/hardware/temp/ktx500"
	"time"
)

func defaultConfig() Config {
	c := Config{
		LogComm:        false,
		FloatPrecision: 6,
		Gas: Gas{
			Type:               gas.Mil82,
			Addr:               100,
			Comport:            "COM1",
			TimeoutGetResponse: time.Second,
			TimeoutEndResponse: time.Millisecond * 50,
			MaxAttemptsRead:    0,
			BlowDuration: [6]time.Duration{
				5 * time.Minute,
				5 * time.Minute,
				5 * time.Minute,
				5 * time.Minute,
				5 * time.Minute,
				5 * time.Minute,
			},
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
		InactiveCoefficients: make(map[Kef]struct{}),
		Hardware:             devicecfg.Hardware{},
	}
	return c
}
