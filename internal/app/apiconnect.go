package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/gui/comports"
	"github.com/fpawel/atool/internal/gui/guiwork"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/comm/modbus"
)

type hardwareConnSvc struct{}

var _ api.HardwareConnectionService = new(hardwareConnSvc)

func (h *hardwareConnSvc) Connect(_ context.Context) error {
	return runInterrogate()
}

func (h *hardwareConnSvc) Interrupt(_ context.Context) error {
	if !guiwork.IsConnected() {
		return nil
	}
	guiwork.Interrupt()
	guiwork.Wait()
	return nil
}

func (h *hardwareConnSvc) Connected(_ context.Context) (bool, error) {
	return guiwork.IsConnected(), nil
}

func (h *hardwareConnSvc) Command(_ context.Context, cmd int16, s string) error {
	b, err := parseHexBytes(s)
	if err != nil {
		return merry.New("ожидалась последовательность байтов HEX")
	}
	runRawCommand(modbus.ProtoCmd(cmd), b)
	return nil
}

func (h *hardwareConnSvc) SwitchGas(_ context.Context, valve int8) error {
	guiwork.RunTask(fmt.Sprintf("переключение клапана газового блока: %d", valve), func() (string, error) {
		err := switchGas(context.Background(), byte(valve))
		comports.CloseComport(config.Get().Gas.Comport)
		return "", err
	})
	return nil
}
