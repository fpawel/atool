package app

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/cfg"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/hardware/gas"
)

type hardwareConnSvc struct{}

var _ api.HardwareConnectionService = new(hardwareConnSvc)

func (h *hardwareConnSvc) Connect(_ context.Context) error {
	if connected() {
		return nil
	}
	runInterrogate()
	return nil
}

func (h *hardwareConnSvc) Disconnect(_ context.Context) error {
	if !connected() {
		return nil
	}
	disconnect()
	wgConnect.Wait()
	return nil
}

func (h *hardwareConnSvc) Connected(_ context.Context) (bool, error) {
	return connected(), nil
}

func (h *hardwareConnSvc) Command(_ context.Context, cmd int16, s string) error {
	b, err := parseHexBytes(s)
	if err != nil {
		return merry.New("ожидалась последовательность байтов HEX")
	}
	if connected() {
		disconnect()
		wgConnect.Wait()
	}
	runRawCommand(modbus.ProtoCmd(cmd), b)
	return nil
}

func (h *hardwareConnSvc) SwitchGas(ctx context.Context, deviceType int8, comport string, valve int8) error {
	c := cfg.Get()
	c.Gas.Type = gas.DevType(deviceType)
	c.Gas.Comport = comport
	if err := c.Validate(); err != nil {
		return err
	}
	err := switchGas(ctx, c.Gas, byte(valve))
	if err == nil {
		cfg.Set(c)
	}
	return err
}
