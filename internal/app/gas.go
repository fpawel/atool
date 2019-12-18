package app

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/hardware/gas"
	"time"
)

func switchGas(ctx context.Context, valve byte) error {
	c := wrk.cfg.Gas
	port, err := wrk.getComport(comport.Config{
		Name:        c.Comport,
		Baud:        9600,
		ReadTimeout: time.Millisecond,
	})
	if err != nil {
		return merry.Append(err, "COM порт газового блока")
	}
	rdr := modbus.NewResponseReader(port, comm.Config{
		TimeoutGetResponse: c.TimeoutGetResponse,
		TimeoutEndResponse: c.TimeoutEndResponse,
		MaxAttemptsRead:    c.MaxAttemptsRead,
		Pause:              0,
	})
	return gas.Switch(log, ctx, c.Type, rdr, c.Addr, valve)
}
