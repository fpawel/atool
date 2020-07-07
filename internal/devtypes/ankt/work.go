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

func main(log comm.Logger, ctx context.Context) error {
	w := &wrk{
		temps: make(map[keyTemp]float64),
	}
	return w.main(log, ctx)
}

var warn = hardware.WithWarn{}

type wrk struct {
	t     productType
	C     [6]float64
	temps map[keyTemp]float64
}

func (w wrk) main(log comm.Logger, ctx context.Context) error {

	party, err := data.GetCurrentParty()
	if err != nil {
		return err
	}
	if party.DeviceType != deviceName {
		return merry.Errorf("нельзя выполнить настройку Анкат-7664МИКРО для %s", party.DeviceType)
	}
	pv, err := data.GetPartyValues1(party.PartyID)
	if err != nil {
		return err
	}

	Tnorm, ok := pv[keyTempNorm.String()]
	if !ok {
		return merry.Errorf("нет значения %q", keyTempNorm)
	}
	Tlow, ok := pv[keyTempLow.String()]
	if !ok {
		return merry.Errorf("нет значения %q", keyTempLow)
	}
	Thigh, ok := pv[keyTempHigh.String()]
	if !ok {
		return merry.Errorf("нет значения %q", keyTempHigh)
	}

	w.temps[keyTempNorm] = Tnorm
	w.temps[keyTempLow] = Tlow
	w.temps[keyTempHigh] = Thigh

	w.t, ok = productTypes[party.ProductType]
	if !ok {
		return merry.Errorf("%s не правильное исполнение %s", party.DeviceType, party.ProductType)
	}

	for i := 0; i < 6; i++ {
		var ok bool
		w.C[i], ok = pv[fmt.Sprintf("c%d", i+1)]
		if !ok {
			if i == 2 && !w.t.Chan[1].gas.isCO2() || i > 4 && !w.t.Chan2 {
				continue
			}
			return merry.Errorf("нет значения ПГС%d", i)
		}
	}

	works := w.mainWorks().ExecuteSelectWorksDialog(ctx.Done())
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return works.Do(log, ctx)
}

func (w wrk) mainWorks() workgui.Works {

	newWork := workgui.New

	isCO2 := w.t.Chan[0].gas.isCO2()
	isChan2 := w.t.Chan2

	readSaveLin := func(gas gas, Chan cChan) workgui.WorkFunc {
		return workgui.NewWorkFuncFromList(
			blowGas(gas),
			readAndSaveParam(Chan.Cout(), Chan.dbKeyLin(gas)),
		)
	}

	readSaveTempGas := func(Chan cChan, keyTemp keyTemp, gas gas) workgui.WorkFunc {
		return workgui.NewWorkFuncFromList(
			blowGas(gas),
			w.readAndSaveParams(dbKeyTemp(Chan, keyTemp, gas)),
		)
	}

	readSaveTemp := func(keyTemp keyTemp) workgui.WorkFunc {
		return newFuncLst(
			w.holdTemp(keyTemp),
			readSaveTempGas(chan1, keyTemp, gas1),
			w.readAndSaveParams(dbKeyTemp(2, keyTempNorm, gas1)).ApplyIf(isChan2),
			readSaveTempGas(chan1, keyTempNorm, gas4),
			readSaveTempGas(chan2, keyTempNorm, gas6).ApplyIf(isChan2),
			blowGas(gas1),
		)
	}

	return workgui.NewWorks(
		newWork("корректировка температуры mcu", correctTmcu),
		newWork("установка режима работы 2", setWorkMode(2)),
		newWork("установка коэфффициентов", w.writeInitCfs),
		newWork("нормировка", newFuncLst(
			blowGas(gas1),
			write32(8, 1000),
			write32(9, 1000).ApplyIf(isChan2),
		)),
		newWork("калибровка", newFuncLst(
			blowGas(gas1),
			write32(1, w.C[0]),
			write32(4, w.C[0]).ApplyIf(isChan2),

			blowGas(gas4),
			write32(2, w.C[3]),

			newFuncLst(
				blowGas(gas5),
				write32(5, w.C[5]),
			).ApplyIf(isChan2),
		)),

		newWork("линеаризация", newFuncLst(
			readSaveLin(gas1, chan1),
			readAndSaveParam(chan2.Cout(), chan2.dbKeyLin(gas1)).ApplyIf(isChan2),
			readSaveLin(gas2, chan1),
			readSaveLin(gas3, chan1).ApplyIf(isCO2),
			readSaveLin(gas4, chan1),
			readSaveLin(gas5, chan2).ApplyIf(isChan2),
			readSaveLin(gas6, chan2).ApplyIf(isChan2),
		)),

		newWork("термокомпенсация", newFuncLst(
			readSaveTemp(keyTempNorm),
			readSaveTemp(keyTempLow),
			readSaveTemp(keyTempHigh),
		)),
	)
}

func write32(cmd modbus.DevCmd, value float64) workgui.WorkFunc {
	return workparty.Write32(cmd, floatBitsFormat, value)
}

func readAndSaveParam(param modbus.Var, dbKey string) workgui.WorkFunc {
	return workparty.ReadAndSaveProductParam(param, floatBitsFormat, dbKey)
}

func (w wrk) readAndSaveParams(dbKey string) workgui.WorkFunc {
	var xs workgui.WorkFuncList
	for _, Var := range deviceConfig.ParamAddresses(w.t.String()) {
		xs = append(xs, workparty.ReadAndSaveProductParam(Var, floatBitsFormat, dbKey))
	}
	return xs.Do
}

func (w wrk) writeInitCfs(log comm.Logger, ctx context.Context) error {
	var xs []workparty.ProductCoefficientValue
	products, err := data.GetActiveProducts()
	if err != nil {
		return err
	}
	for _, p := range products {
		xs = append(xs, w.initProductCfsValues(p)...)
	}
	return workparty.WriteProdsCfs(xs, nil)(log, ctx)
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
		warn.HoldTemperature(20),
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

func (w wrk) initProductCfsValues(p data.Product) (ks []workparty.ProductCoefficientValue) {
	xs := workparty.CfsValues{
		2: float64(time.Now().Year()),
		3: float64(p.Serial),

		10: w.C[0],
		11: w.C[3],
		19: w.C[0],
		20: w.C[5],

		43: 740, // PC
		44: 0,
		45: 0, // PT
		46: 1,
		47: 0,
		23: 0, // LIN1
		24: 1,
		25: 0,
		26: 0,
		33: 0, // LIN2
		34: 1,
		35: 0,
		36: 0,
		27: 0, // T0 1
		28: 0,
		29: 0,
		30: 0, // TK 1
		31: 0,
		32: 0,
		37: 0, // T0 2
		38: 0,
		39: 0,
		40: 0, // TK 2
		41: 0,
		42: 0,
	}
	for k, v := range xs {
		ks = append(ks, workparty.ProductCoefficientValue{
			ProductID:   p.ProductID,
			Coefficient: k,
			Value:       v,
		})
	}
	return
}

func (w wrk) holdTemp(temp keyTemp) workgui.WorkFunc {
	temp.mustCheck()
	return hardwareWarn.HoldTemperature(w.temps[temp])
}

func blowGas(gas gas) workgui.WorkFunc {
	gas.mustCheck()
	return hardwareWarn.BlowGas(byte(gas))
}

var (
	newFuncLst = workgui.NewWorkFuncFromList

	hardwareWarn = hardware.WithWarn{}
)
