package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/atool/internal/workparty"
	"github.com/fpawel/hardware/temp/ktx500"
)

type tempDeviceSvc struct{}

var _ api.TemperatureDeviceService = new(tempDeviceSvc)

func (tempDeviceSvc) Start(context.Context) error {
	workgui.RunTask(log, "термокамера: старт", func() error {
		d, err := workparty.GetTemperatureDevice()
		if err != nil {
			return err
		}
		return d.Start(log, context.Background())
	})
	return nil
}

func (tempDeviceSvc) Stop(context.Context) error {
	workgui.RunTask(log, "термокамера: стоп", func() error {
		d, err := workparty.GetTemperatureDevice()
		if err != nil {
			return err
		}
		return d.Stop(log, context.Background())
	})
	return nil
}

func (tempDeviceSvc) Setup(_ context.Context, temperature float64) error {
	workgui.RunTask(log, fmt.Sprintf("термокамера: уставка %v", temperature), func() error {
		d, err := workparty.GetTemperatureDevice()
		if err != nil {
			return err
		}
		return d.Setup(log, context.Background(), temperature)
	})
	return nil
}
func (tempDeviceSvc) CoolingOn(context.Context) error {
	workgui.RunTask(log, "термокамера: включить охлаждение", func() error {
		dd, err := workparty.GetTemperatureDevice()
		if err != nil {
			return err
		}
		d, f := dd.(ktx500.Client)
		if !f {
			return merry.Errorf("заданный тип термокамеры %q не поддерживает управление охлаждением",
				config.Get().Temperature.Type)
		}
		return d.CoolingOn(log, context.Background())
	})
	return nil
}
func (tempDeviceSvc) CoolingOff(context.Context) error {
	workgui.RunTask(log, "термокамера: выключить охлаждение", func() error {
		d, err := workparty.GetTemperatureDevice()
		if err != nil {
			return err
		}
		c, f := d.(ktx500.Client)
		if !f {
			return merry.Errorf("заданный тип термокамеры %q не поддерживает управление охлаждением",
				config.Get().Temperature.Type)
		}
		return c.CoolingOff(log, context.Background())
	})
	return nil
}
