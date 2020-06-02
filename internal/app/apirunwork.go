package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/pkg/comports"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/atool/internal/workparty"
	"github.com/fpawel/comm/modbus"
)

type runWorkSvc struct{}

var _ api.RunWorkService = new(runWorkSvc)

func (h *runWorkSvc) SearchProducts(ctx context.Context, comportName string) error {
	return workparty.RunSearchProducts(log, ctx, comportName)
}

func (h *runWorkSvc) Connect(_ context.Context) error {
	return workparty.RunInterrogate(log, appCtx)
}

func (h *runWorkSvc) Interrupt(_ context.Context) error {
	if !workgui.IsConnected() {
		return nil
	}
	workgui.Interrupt()
	workgui.Wait()
	return nil
}

func (h *runWorkSvc) InterruptDelay(_ context.Context) error {
	workgui.InterruptDelay(log)
	return nil
}

func (h *runWorkSvc) Connected(_ context.Context) (bool, error) {
	return workgui.IsConnected(), nil
}

func (h *runWorkSvc) Command(_ context.Context, cmd int16, s string) error {
	b, err := parseHexBytes(s)
	if err != nil {
		return merry.New("ожидалась последовательность байтов HEX")
	}
	workparty.RunRawCommand(log, appCtx, modbus.ProtoCmd(cmd), b)
	return nil
}

func (h *runWorkSvc) SwitchGas(_ context.Context, valve int8) error {
	workgui.RunTask(log, fmt.Sprintf("переключение пневмоблока %d", valve), func() error {
		err := workparty.SwitchGas(log, context.Background(), byte(valve))
		comports.CloseComport(config.Get().Gas.Comport)
		return err
	})
	return nil
}
