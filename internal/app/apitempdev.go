package app

import (
	"context"
	"fmt"
	"github.com/fpawel/atool/internal/hardware"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/comm"
)

type tempDeviceSvc struct{}

var _ api.TemperatureDeviceService = new(tempDeviceSvc)

func (tempDeviceSvc) Start(context.Context) error {
	workgui.New("термокамера: старт", func(log comm.Logger, ctx context.Context) error {
		return hardware.TemperatureStart(log, ctx)
	})
	return nil
}

func (tempDeviceSvc) Stop(context.Context) error {
	workgui.New("термокамера: стоп", func(log comm.Logger, ctx context.Context) error {
		return hardware.TemperatureStop(log, ctx)
	})
	return nil
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
