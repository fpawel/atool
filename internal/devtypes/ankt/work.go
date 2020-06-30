package ankt

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/devtypes/ankt/anktvar"
	"github.com/fpawel/atool/internal/hardware"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/atool/internal/workparty"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"time"
)

const floatBitsFormat = modbus.BCD

var hardwareWarn = hardware.WithWarn{}

type wrk struct {
	t productType
}

func (w wrk) mainWorks() workgui.Works {
	return workgui.NewWorks(
		workgui.New("корректировка температуры mcu", correctTmcu),
		workgui.New("установка режима работы 2", setWorkMode(2)),
	)
}

func (w wrk) writeInitKfs(log comm.Logger, ctx context.Context) error {
	var xs []workparty.ProductCoefficientValue
	products, err := data.GetActiveProducts()
	if err != nil {
		return err
	}
	for _, p := range products {
		p := p
		kValues := workparty.CfsValues{
			3:  float64(p.Serial),
			43: 740,
			44: 0,
			2:  float64(time.Now().Year()),
			45: 0,
			46: 1,
			47: 0,
		}
		for k, v := range kValues {
			xs = append(xs, workparty.ProductCoefficientValue{
				ProductID:   p.ProductID,
				Coefficient: k,
				Value:       v,
			})
		}
	}
	return nil
}

func setWorkMode(value float64) workgui.WorkFunc {
	return workparty.ProcessEachActiveProduct(nil, func(log comm.Logger, ctx context.Context, p workparty.Product) error {
		_, err := modbus.Request{
			Addr:     p.Addr,
			ProtoCmd: 0x16,
			Data:     append([]byte{0xA0, 0, 0, 2, 4}, modbus.BCD6(value)...),
		}.GetResponse(log, ctx, p.Comm())
		if err != nil {
			return merry.Prependf(err, "установка режима работы %v", value)

		}
		workgui.NotifyInfo(log, fmt.Sprintf("%s: установлен режим работы %v", p, value))
		return nil
	})
}

func correctTmcu(log comm.Logger, ctx context.Context) error {
	const kefKdFt devicecfg.Kef = 48
	return workgui.NewWorkFuncList(
		workparty.WriteCfsValues(workparty.CfsValues{kefKdFt: 273}, floatBitsFormat),
		hardwareWarn.HoldTemperature(20),
		workparty.ProcessEachActiveProduct(nil, func(log comm.Logger, ctx context.Context, p workparty.Product) error {
			k48, err := p.ReadKef(log, ctx, kefKdFt, floatBitsFormat)
			if err != nil {
				return err
			}
			temp, err := hardware.GetCurrentTemperature(log, ctx)
			if err != nil {
				return err
			}
			tMcu, err := modbus.Read3Value(log, ctx, p.Comm(), p.Addr, anktvar.Tmcu, floatBitsFormat)
			if err != nil {
				return err
			}
			k49 := k48 + temp - tMcu
			workgui.NotifyInfo(log, fmt.Sprintf("%s: K49 = K48 + temp - tMcu = %v + %v - %v = %v", p, k48, temp, tMcu, k49))
			return p.WriteKef(49, floatBitsFormat, k48+temp-tMcu)(log, ctx)
		}),

		workparty.ReadCfs([]devicecfg.Kef{kefKdFt}, floatBitsFormat),
	).Do(log, ctx)
}
