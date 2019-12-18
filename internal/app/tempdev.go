package app

import (
	"github.com/fpawel/atool/internal/cfg"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/gofins/fins"
	"github.com/fpawel/hardware/temp"
	"github.com/fpawel/hardware/temp/ktx500"
	"github.com/fpawel/hardware/temp/tempmil82"
	"time"
)

var (
	ktx500Client *fins.Client
)

func getTemperatureDevice() (temp.TemperatureDevice, error) {
	wrk.closeTemperatureDevice()
	c := wrk.cfg.Temperature
	if err := c.Validate(); err != nil {
		return nil, err
	}

	getComportReader := func() (comm.ResponseReader, error) {
		port, err := wrk.getComport(comport.Config{
			Name:        c.Comport,
			Baud:        9600,
			ReadTimeout: time.Millisecond,
		})
		if err != nil {
			return comm.ResponseReader{}, err
		}
		return comm.NewResponseReader(port, comm.Config{
			TimeoutGetResponse: c.TimeoutGetResponse,
			TimeoutEndResponse: c.TimeoutEndResponse,
			MaxAttemptsRead:    c.MaxAttemptsRead,
			Pause:              0,
		}, nil), nil
	}
	switch c.Type {
	case cfg.T800:
		r, err := getComportReader()
		if err != nil {
			return nil, err
		}
		return tempmil82.NewT800(r), nil
	case cfg.T2500:
		r, err := getComportReader()
		if err != nil {
			return nil, err
		}
		return tempmil82.NewT2500(r), nil
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
