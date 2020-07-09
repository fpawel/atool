package app

import (
	"context"
	"fmt"
	"github.com/fpawel/atool/internal/hardware"
	"github.com/fpawel/atool/internal/thriftgen/api"
)

type tempDeviceSvc struct{}

var _ api.TemperatureDeviceService = new(tempDeviceSvc)

func (tempDeviceSvc) Start(context.Context) error {
	runSingleTask(hardware.TemperatureStart)
	return nil
}

func (tempDeviceSvc) Stop(context.Context) error {
	runSingleTask(hardware.TemperatureStop)
	return nil
}

func (tempDeviceSvc) SetDestination(ctx context.Context, temperature float64) (err error) {
	runSingleTask(hardware.TemperatureSetDestination(temperature))
	return nil
}
func (tempDeviceSvc) GetTemperature(ctx context.Context) (r float64, err error) {
	return hardware.GetCurrentTemperature(log, appCtx)
}
func (tempDeviceSvc) Setup(_ context.Context, temperature float64) error {
	return runWork(
		hardware.TemperatureSetup(temperature).
			Work(fmt.Sprintf("термокамера: уставка %v", temperature)))
}
func (tempDeviceSvc) CoolingOn(context.Context) error {
	return runWorkFunc("термокамера: включить охлаждение", hardware.SetTemperatureCool(true))
}

func (tempDeviceSvc) CoolingOff(context.Context) error {
	return runWorkFunc("термокамера: выключить охлаждение", hardware.SetTemperatureCool(false))
}
