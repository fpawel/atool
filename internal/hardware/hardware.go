package hardware

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/config/appcfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg/comports"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/atool/internal/workparty"
	"github.com/fpawel/comm"
	"github.com/fpawel/gofins/fins"
	"github.com/fpawel/hardware/gas"
	"github.com/fpawel/hardware/temp"
	"github.com/fpawel/hardware/temp/ktx500"
	"github.com/fpawel/hardware/temp/tempcomport"
	"github.com/powerman/structlog"
	"math"
)

func TemperatureSetup(log comm.Logger, ctx context.Context, destinationTemperature float64) error {

	what := fmt.Sprintf("🌡 перевод термокамеры на %v⁰C", destinationTemperature)

	return workgui.Perform(log, ctx, what, func(log *structlog.Logger, ctx context.Context) error {
		// отключить газ
		_ = SwitchGas(log, ctx, 0)

		if !state.temp {
			if err := workgui.Perform(log, ctx, "запуск термокамеры", func(log *structlog.Logger, ctx context.Context) error {
				if err := TemperatureStop(log, ctx); err != nil {
					return err
				}
				if err := TemperatureStart(log, ctx); err != nil {
					return err
				}
				return nil
			}); err != nil {
				return err
			}
		}

		// запись уставки
		if err := TemperatureSetDestination(log, ctx, destinationTemperature); err != nil {
			return err
		}

		// измерения, полученные в процесе опроса приборов во время данной задержки
		ms := new(data.MeasurementCache)
		defer ms.Save()
		errorsOccurred := workparty.ErrorsOccurred{}

		for {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			currentTemperature, err := GetCurrentTemperature(log, ctx)
			if err != nil {
				err = merry.Prepend(err, what+", считывание температуры")
				workgui.NotifyErr(log, err)
				return err
			}
			log.Info(fmt.Sprintf("🌡 температура %v⁰C", currentTemperature))

			if math.Abs(currentTemperature-destinationTemperature) < 2 {
				workgui.NotifyInfo(log, fmt.Sprintf("🌡 термокамера вышла на температуру %v⁰C: %v⁰C", destinationTemperature, currentTemperature))
				return nil
			}

			// фоновый опрос приборов
			_ = workparty.ReadProductsParams(log, ctx, ms, errorsOccurred)
		}
	})
}

func TemperatureSetDestination(log comm.Logger, ctx context.Context, destinationTemperature float64) error {
	what := fmt.Sprintf("🌡 запись уставки термокамеры %v⁰C", destinationTemperature)
	return workgui.WithNotifyResult(log, what, func() error {
		dev, err := GetTemperatureDevice()
		if err != nil {
			return err
		}
		if err := dev.Setup(log, ctx, destinationTemperature); err != nil {
			return err
		}
		go gui.NotifyTemperatureSetPoint(destinationTemperature)
		return nil
	})
}

func TemperatureStart(log comm.Logger, ctx context.Context) error {
	return workgui.WithNotifyResult(log, "🌡 термокамера - старт", func() error {
		tempDevice, err := GetTemperatureDevice()
		if err != nil {
			return err
		}
		if err := tempDevice.Start(log, ctx); err != nil {
			return err
		}
		state.temp = true
		return nil
	})
}

func TemperatureStop(log comm.Logger, ctx context.Context) error {
	return workgui.WithNotifyResult(log, "🌡 термокамера - остановка", func() error {
		tempDevice, err := GetTemperatureDevice()
		if err != nil {
			return err
		}
		if err := tempDevice.Stop(log, ctx); err != nil {
			return err
		}
		state.temp = false
		return nil
	})
}

func GetCurrentTemperature(log comm.Logger, ctx context.Context) (float64, error) {
	tempDevice, err := GetTemperatureDevice()
	if err != nil {
		return math.NaN(), err
	}
	currentTemperature, err := tempDevice.Read(log, ctx)
	if err != nil {
		return math.NaN(), err
	}
	go gui.NotifyTemperature(currentTemperature)
	return currentTemperature, nil
}

func SwitchGas(log comm.Logger, ctx context.Context, valve byte) error {
	return workgui.WithNotifyResult(log, fmt.Sprintf("⛏ переключение газового блока %d", valve), func() error {
		c := appcfg.Cfg.Gas
		port := comports.GetComport(c.Comport, 9600)
		commCfg := comm.Config{
			TimeoutGetResponse: c.TimeoutGetResponse,
			TimeoutEndResponse: c.TimeoutEndResponse,
			MaxAttemptsRead:    c.MaxAttemptsRead,
		}
		err := gas.Switch(log, ctx, c.Type, comm.New(port, commCfg), c.Addr, valve)
		if err != nil {
			return err
		}
		go gui.NotifyGas(int(valve))
		state.gas = valve != 0
		return nil
	})
}

func CloseHardware(log comm.Logger, ctx context.Context) {

	if state.gas {
		_ = workgui.WithNotifyResult(log, "⛏ отключить газ по окончании настройки", func() error {
			return SwitchGas(log, ctx, 0)
		})
	}
	if state.temp {
		_ = workgui.WithNotifyResult(log, "⛏ остановить термокамеру по окончании настройки", func() error {
			return SwitchGas(log, ctx, 0)
		})

		if err := TemperatureStop(log, ctx); err != nil {
			workgui.NotifyErr(log, merry.Prepend(err, "остановить термокамеру по окончании настройки"))
		} else {
			workgui.NotifyInfo(log, "термокамера остановлена по окончании настройки")
		}
	}
	state.gas = false
	state.temp = false
	state.blowGas = 0
}

func GetTemperatureDevice() (temp.TemperatureDevice, error) {
	comports.CloseComport(appcfg.Cfg.Temperature.Comport)

	confTemp := appcfg.Cfg.Temperature
	if err := confTemp.Validate(); err != nil {
		return nil, err
	}

	switch confTemp.Type {
	case config.T800:
		return tempcomport.T800(getTemperatureComportReader()), nil
	case config.T2500:
		return tempcomport.T2500(getTemperatureComportReader()), nil
	default:
		if ktx500Client != nil {
			ktx500Client.Close()
		}
		var err error
		ktx500Client, err = appcfg.Cfg.Ktx500.NewFinsClient()
		if err != nil {
			return nil, err
		}
		return ktx500.NewClient(ktx500Client, confTemp.MaxAttemptsRead), nil
	}
}

func getTemperatureComportReader() comm.T {
	c := appcfg.Cfg.Temperature
	return comm.New(
		comports.GetComport(c.Comport, 9600),
		comm.Config{
			TimeoutGetResponse: c.TimeoutGetResponse,
			TimeoutEndResponse: c.TimeoutEndResponse,
			MaxAttemptsRead:    c.MaxAttemptsRead,
			Pause:              0,
		})
}

var (
	state struct {
		gas, temp bool
		blowGas   byte
	}

	ktx500Client *fins.Client
)
