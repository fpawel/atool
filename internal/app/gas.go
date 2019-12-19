package app

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/gui/comports"
	"github.com/fpawel/comm"
	"github.com/fpawel/hardware/gas"
)

func switchGas(ctx context.Context, valve byte) error {
	c := config.Get().Gas
	port, err := comports.GetComport(c.Comport, 9600)
	if err != nil {
		return merry.Append(err, "COM порт газового блока")
	}
	commCfg := comm.Config{
		TimeoutGetResponse: c.TimeoutGetResponse,
		TimeoutEndResponse: c.TimeoutEndResponse,
		MaxAttemptsRead:    c.MaxAttemptsRead,
	}
	return gas.Switch(log, ctx, c.Type, port, commCfg, c.Addr, valve)
}
