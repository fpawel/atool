package app

import (
	"context"
	"fmt"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/guiwork"
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
	guiwork.NotifyInfo(log, fmt.Sprintf("переключение пневмоблока %d", valve))

	err := gas.Switch(log, ctx, c.Type, comm.New(port, commCfg), c.Addr, valve)
	if err == nil {
		go gui.NotifyGas(int(valve))
		guiwork.NotifyInfo(log, fmt.Sprintf("пневомблок переключен %d", valve))
	}
	return err
}
