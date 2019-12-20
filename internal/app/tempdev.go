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
		rdr, err := getTemperatureComportReader()
		if err != nil {
			return nil, err
		}
		return tempcomport.NewT800(rdr), nil
	case config.T2500:
		rdr, err := getTemperatureComportReader()
		if err != nil {
			return nil, err
		}
		return tempcomport.NewT2500(rdr), nil
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

func getTemperatureComportReader() (tempcomport.ResponseReader, error) {
	c := config.Get().Temperature
	return tempcomport.ResponseReader{
		Wr: comports.GetComport(c.Comport, 9600),
		C: comm.Config{
			TimeoutGetResponse: c.TimeoutGetResponse,
			TimeoutEndResponse: c.TimeoutEndResponse,
			MaxAttemptsRead:    c.MaxAttemptsRead,
			Pause:              0,
		},
	}, nil
}
