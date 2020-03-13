package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/guiwork"
	"github.com/fpawel/atool/internal/pkg/comports"
	"github.com/fpawel/comm"
	"github.com/fpawel/gofins/fins"
	"github.com/fpawel/hardware/temp"
	"github.com/fpawel/hardware/temp/ktx500"
	"github.com/fpawel/hardware/temp/tempcomport"
	"math"
)

var (
	ktx500Client *fins.Client
)

func getTemperatureDevice() (temp.TemperatureDevice, error) {
	comports.CloseComport(config.Get().Temperature.Comport)
	conf := config.Get()
	confTemp := conf.Temperature
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
		ktx500Client, err = conf.Ktx500.NewFinsClient()
		if err != nil {
			return nil, err
		}
		return ktx500.NewClient(ktx500Client, confTemp.MaxAttemptsRead), nil
	}
}

func getTemperatureComportReader() comm.T {
	c := config.Get().Temperature
	return comm.New(
		comports.GetComport(c.Comport, 9600),
		comm.Config{
			TimeoutGetResponse: c.TimeoutGetResponse,
			TimeoutEndResponse: c.TimeoutEndResponse,
			MaxAttemptsRead:    c.MaxAttemptsRead,
			Pause:              0,
		})
}

func setupTemperature(log logger, ctx context.Context, destinationTemperature float64) error {

	wrapErr := func(err error) error {
		if err == nil {
			return nil
		}
		return merry.Appendf(err, "термокамера: перевод на %v⁰C", destinationTemperature)
	}

	dev, err := getTemperatureDevice()
	if err != nil {
		return wrapErr(err)
	}
	if err := dev.Setup(log, ctx, destinationTemperature); err != nil {
		return wrapErr(err)
	}

	go gui.NotifyTemperatureSetPoint(destinationTemperature)

	// измерения, полученные в процесе опроса приборов во время данной задержки
	ms := new(measurements)

	defer ms.Save()

	errorsOccurred := errorsOccurred{}

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		currentTemperature, err := dev.Read(log, ctx)

		if err != nil {
			err = wrapErr(merry.Append(err, "считывание температуры"))
			guiwork.JournalErr(log, err)
			return err
		}

		log.Info(fmt.Sprintf("температура %v⁰C", currentTemperature))

		go gui.NotifyTemperature(currentTemperature)

		if math.Abs(currentTemperature-destinationTemperature) < 2 {
			guiwork.JournalInfo(log, fmt.Sprintf("термокамера вышла на температуру %v⁰C: %v⁰C", destinationTemperature, currentTemperature))
			return nil
		}

		if err := readProductsParams(ctx, ms, errorsOccurred); err != nil {
			guiwork.JournalErr(log, wrapErr(err))
		}
	}
}
