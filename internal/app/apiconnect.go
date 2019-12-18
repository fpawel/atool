package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/comm/modbus"
)

type hardwareConnSvc struct{}

var _ api.HardwareConnectionService = new(hardwareConnSvc)

func (h *hardwareConnSvc) Connect(_ context.Context) error {
	if wrk.connected() {
		return nil
	}
	runInterrogate()
	return nil
}

func (h *hardwareConnSvc) Disconnect(_ context.Context) error {
	if !wrk.connected() {
		return nil
	}
	wrk.disconnect()
	wrk.wg.Wait()
	return nil
}

func (h *hardwareConnSvc) Connected(_ context.Context) (bool, error) {
	return wrk.connected(), nil
}

func (h *hardwareConnSvc) Command(_ context.Context, cmd int16, s string) error {
	b, err := parseHexBytes(s)
	if err != nil {
		return merry.New("ожидалась последовательность байтов HEX")
	}
	if wrk.connected() {
		wrk.disconnect()
		wrk.wg.Wait()
	}
	runRawCommand(modbus.ProtoCmd(cmd), b)
	return nil
}

func (h *hardwareConnSvc) SwitchGas(_ context.Context, valve int8) error {
	runTask(fmt.Sprintf("переключение клапана газового блока: %d", valve), func() (string, error) {
		err := switchGas(context.Background(), byte(valve))
		wrk.closeGasComport()
		return "", err
	})
	return nil
}
