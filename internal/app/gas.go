package app

import (
	"context"
	"fmt"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg/comports"
	"github.com/fpawel/comm"
	"github.com/fpawel/hardware/gas"
)

func switchGas(ctx context.Context, valve byte) error {
	c := config.Get().Gas
	port := comports.GetComport(c.Comport, 9600)
	commCfg := comm.Config{
		TimeoutGetResponse: c.TimeoutGetResponse,
		TimeoutEndResponse: c.TimeoutEndResponse,
		MaxAttemptsRead:    c.MaxAttemptsRead,
	}
	err := gas.Switch(log, ctx, c.Type, port, commCfg, c.Addr, valve)
	ct := gui.CommTransaction{
		Addr:    c.Addr,
		Comport: c.Comport,
		Request: fmt.Sprintf("пневмоблок: %d", valve),
	}
	if err == nil {
		ct.Response = "ok"
		ct.Ok = true
	} else {
		ct.Response = err.Error()
	}
	go gui.NotifyNewCommTransaction(ct)
	return err
}
