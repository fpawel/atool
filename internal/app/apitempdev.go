package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/guiwork"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/hardware/temp/ktx500"
)

type tempDeviceSvc struct{}

var _ api.TemperatureDeviceService = new(tempDeviceSvc)

func (tempDeviceSvc) Start(context.Context) error {
	guiwork.RunTask(log, "термокамера: старт", func() error {
		d, err := getTemperatureDevice()
		if err != nil {
			return err
		}
		return d.Start(log, context.Background())
	})
	return nil
}

func (tempDeviceSvc) Stop(context.Context) error {
	guiwork.RunTask(log, "термокамера: стоп", func() error {
		d, err := getTemperatureDevice()
		if err != nil {
			return err
		}
		return d.Stop(log, context.Background())
	})
	return nil
}

func (tempDeviceSvc) Setup(_ context.Context, temperature float64) error {
	guiwork.RunTask(log, fmt.Sprintf("термокамера: уставка %v", temperature), func() error {
		d, err := getTemperatureDevice()
		if err != nil {
			return err
		}
		return d.Setup(log, context.Background(), temperature)
	})
	return nil
}
func (tempDeviceSvc) CoolingOn(context.Context) error {
	guiwork.RunTask(log, "термокамера: включить охлаждение", func() error {
		dd, err := getTemperatureDevice()
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
	guiwork.RunTask(log, "термокамера: выключить охлаждение", func() error {
		d, err := getTemperatureDevice()
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
