package app

import (
	"context"
	"fmt"
	"github.com/fpawel/atool/internal/config"
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
	err := gas.Switch(log, ctx, c.Type, comm.New(port, commCfg), c.Addr, valve)
	if err == nil {
		guiwork.JournalInfo(log, fmt.Sprintf("газовый блок: %d", valve))
	} else {
		guiwork.JournalErr(log, fmt.Errorf("газовый блок: %d: %s", valve, err))
	}
	return err
}
