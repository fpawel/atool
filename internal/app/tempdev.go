package app

import (
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/pkg/comports"
	"github.com/fpawel/comm"
	"github.com/fpawel/gofins/fins"
	"github.com/fpawel/hardware/temp"
	"github.com/fpawel/hardware/temp/ktx500"
	"github.com/fpawel/hardware/temp/tempcomport"
)

var (
	ktx500Client *fins.Client
)

func getTemperatureDevice() (temp.TemperatureDevice, error) {
	comports.CloseComport(config.Get().Temperature.Comport)
	c := config.Get().Temperature
	if err := c.Validate(); err != nil {
		return nil, err
	}

	switch c.Type {
	case config.T800:
		return tempcomport.T800(getTemperatureComportReader()), nil
	case config.T2500:
		return tempcomport.T2500(getTemperatureComportReader()), nil
	default:
		if ktx500Client != nil {
			ktx500Client.Close()
		}
		var err error
		ktx500Client, err = c.Ktx500.NewFinsClient()
		if err != nil {
			return nil, err
		}
		return ktx500.NewClient(ktx500Client, c.MaxAttemptsRead), err
	}
}

func getTemperatureComportReader() comm.T {
	c := config.Get().Temperature
	return comm.New(
		comports.GetComport(c.Comport, 9600),
		comm.Config{
			TimeoutGetResponse: c.TimeoutGetResponse,
			TimeoutEndResponse: c.TimeoutEndResponse,
			MaxAttemptsRead:    c.MaxAttemptsRead,
			Pause:              0,
		})
}
