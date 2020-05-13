package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/guiwork"
	"github.com/fpawel/atool/internal/pkg/comports"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/comm/modbus"
	"github.com/powerman/structlog"
)

type runWorkSvc struct{}

var _ api.RunWorkService = new(runWorkSvc)

func (h *runWorkSvc) SearchProducts(ctx context.Context, comportName string) error {
	return guiwork.RunWork(log, appCtx, "поиск приборов сети", func(log *structlog.Logger, ctx context.Context) error {
		return searchProducts(log, ctx, comportName)
	})
}

func (h *runWorkSvc) Connect(_ context.Context) error {
	return runInterrogate()
}

func (h *runWorkSvc) Interrupt(_ context.Context) error {
	if !guiwork.IsConnected() {
		return nil
	}
	guiwork.Interrupt()
	guiwork.Wait()
	return nil
}

func (h *runWorkSvc) InterruptDelay(_ context.Context) error {
	guiwork.InterruptDelay(log)
	return nil
}

func (h *runWorkSvc) Connected(_ context.Context) (bool, error) {
	return guiwork.IsConnected(), nil
}

func (h *runWorkSvc) Command(_ context.Context, cmd int16, s string) error {
	b, err := parseHexBytes(s)
	if err != nil {
		return merry.New("ожидалась последовательность байтов HEX")
	}
	runRawCommand(modbus.ProtoCmd(cmd), b)
	return nil
}

func (h *runWorkSvc) SwitchGas(_ context.Context, valve int8) error {
	guiwork.RunTask(log, fmt.Sprintf("переключение клапана газового блока: %d", valve), func() error {
		err := switchGas(context.Background(), byte(valve))
		comports.CloseComport(config.Get().Gas.Comport)
		return err
	})
	return nil
}
