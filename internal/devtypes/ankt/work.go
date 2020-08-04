package ankt

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/devtypes/ankt/anktvar"
	"github.com/fpawel/atool/internal/devtypes/devdata"
	"github.com/fpawel/atool/internal/hardware"
	"github.com/fpawel/atool/internal/pkg/comports"
	"github.com/fpawel/atool/internal/pkg/numeth"
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
			if i == 2 && !w.t.Chan[1].gasName.isCO2() || i > 4 && !w.t.Chan2 {
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

	isCO2 := w.t.Chan[0].gasName.isCO2()
	isChan2 := w.t.Chan2
	isPress := w.t.Press

	readSaveLin := func(gas gas, Chan cChan) workgui.WorkFunc {
		return workgui.NewWorkFuncFromList(
			blowGas(gas),
			readAndSaveVar(Chan.keyLin(gas), Chan.Nfo().Cout),
		)
	}

	readSaveTempGas := func(Chan cChan, keyTemp keyTemp, gas gas) workgui.WorkFunc {
		t := Chan.Nfo()
		return workgui.NewWorkFuncFromList(
			blowGas(gas),
			readAndSaveVar(keyGasVar(keyTemp, gas, t.Tpp), t.Tpp),
			readAndSaveVar(keyGasVar(keyTemp, gas, t.Var2), t.Var2),
		)
	}

	readSaveTemp := func(keyTemp keyTemp) work {
		return newFuncLst(
			w.holdTemp(keyTemp),

			readSaveTempGas(chan1, keyTemp, gas1),

			readAndSaveVar(keyGasVar(keyTemp, gas1, chan2nfo.Tpp), chan2nfo.Tpp).ApplyIf(isChan2),
			readAndSaveVar(keyGasVar(keyTemp, gas1, chan2nfo.Var2), chan2nfo.Var2).ApplyIf(isChan2),
			readAndSaveVar(keyPT(keyTemp), anktvar.VdatP).ApplyIf(isPress),

			readSaveTempGas(chan1, keyTemp, gas4),
			readSaveTempGas(chan2, keyTemp, gas6).ApplyIf(isChan2),
			blowGas(gas1),
		).Work(fmt.Sprintf("снятие термокомпенсации: %s", keyTemp.What()))
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

		newWork("снятие линеаризации", newFuncLst(
			readSaveLin(gas1, chan1),
			readAndSaveVar(chan2.keyLin(gas1), chan2nfo.Cout).ApplyIf(isChan2),
			readSaveLin(gas2, chan1),
			readSaveLin(gas3, chan1).ApplyIf(isCO2),
			readSaveLin(gas4, chan1),
			readSaveLin(gas5, chan2).ApplyIf(isChan2),
			readSaveLin(gas6, chan2).ApplyIf(isChan2),
		)),
		w.calcLin1(),
		w.calcLin2().ApplyIf(isChan2),
		readSaveTemp(keyTempNorm),
		readSaveTemp(keyTempLow),
		readSaveTemp(keyTempHigh),
		w.calcT0Ch1(),
		w.calcTKCh1(),
		w.calcT0Ch2().ApplyIf(isChan2),
		w.calcTKCh2().ApplyIf(isChan2),
		w.calcPT().ApplyIf(isPress),
	)
}

type xy = [2]float64

type work = workgui.Work

func values3(x workparty.ProductValues) func(gas gas, Var modbus.Var) []float64 {
	return func(gas gas, Var modbus.Var) []float64 {
		return x.GetValuesNaN(mapTemps(func(t keyTemp) string {
			return keyGasVar(t, gas, Var)
		}))
	}
}

func (w wrk) calcT0Ch1() work {
	return workparty.InterpolateCfs{
		Name:        "термокомпенсация нуля канала 1",
		Coefficient: kefCh1T0v0,
		Count:       3,
		Format:      floatBitsFormat,
		InterpolateCfsFunc: func(pv workparty.ProductValues) ([]numeth.Coordinate, error) {
			f := values3(pv)
			t := f(gas1, chan1nfo.Tpp)
			v := f(gas1, chan1nfo.Var2)
			return []xy{
				{t[0], -v[0]},
				{t[1], -v[1]},
				{t[2], -v[2]},
			}, nil
		},
	}.Work()
}

func (w wrk) calcT0Ch2() work {
	return workparty.InterpolateCfs{
		Name:        "термокомпенсация нуля канала 2",
		Coefficient: kefCh2T0v0,
		Count:       3,
		Format:      floatBitsFormat,
		InterpolateCfsFunc: func(pv workparty.ProductValues) ([]numeth.Coordinate, error) {
			f := values3(pv)
			t := f(gas1, chan2nfo.Tpp)
			v := f(gas1, chan2nfo.Var2)
			return []xy{
				{t[0], -v[0]},
				{t[1], -v[1]},
				{t[2], -v[2]},
			}, nil
		},
	}.Work()
}

func (w wrk) calcTKCh1() work {
	return workparty.InterpolateCfs{
		Name:        "термокомпенсация конца шкалы канала 1",
		Coefficient: kefCh1TKv0,
		Count:       3,
		Format:      floatBitsFormat,
		InterpolateCfsFunc: func(pv workparty.ProductValues) ([]numeth.Coordinate, error) {
			f := values3(pv)
			t := f(gas4, chan1nfo.Tpp)
			var4 := f(gas4, chan1nfo.Var2)
			var1 := f(gas1, chan1nfo.Var2)
			xy := make([]numeth.Coordinate, 3)
			for i := 0; i < 3; i++ {
				xy[i] = numeth.Coordinate{t[i], (var4[2] - var1[2]) / (var4[i] - var1[i])}
			}
			return xy, nil
		},
	}.Work()
}

func (w wrk) calcTKCh2() work {
	return workparty.InterpolateCfs{
		Name:        "термокомпенсация конца шкалы 2",
		Coefficient: kefCh2TKv0,
		Count:       3,
		Format:      floatBitsFormat,
		InterpolateCfsFunc: func(pv workparty.ProductValues) ([]numeth.Coordinate, error) {
			f := values3(pv)
			t := f(gas6, chan2nfo.Tpp)
			var4 := f(gas6, chan2nfo.Var2)
			var1 := f(gas1, chan2nfo.Var2)
			xy := make([]numeth.Coordinate, 3)
			for i := 0; i < 3; i++ {
				xy[i] = numeth.Coordinate{t[i], (var4[2] - var1[2]) / (var4[i] - var1[i])}
			}
			return xy, nil
		},
	}.Work()
}

func (w wrk) calcPT() work {
	return workparty.InterpolateCfs{
		Name:        "компенсация влияния темпемпературы на давление",
		Coefficient: kefPT0,
		Count:       3,
		Format:      floatBitsFormat,
		InterpolateCfsFunc: func(pv workparty.ProductValues) ([]numeth.Coordinate, error) {
			f := values3(pv)
			t := f(gas1, chan1nfo.Tpp)
			v := pv.GetValuesNaN(mapTemps(keyPT))
			return []xy{
				{t[0], v[0]},
				{t[1], v[1]},
				{t[2], v[2]},
			}, nil
		},
	}.Work()
}

func (w wrk) calcLin1() work {
	return workparty.InterpolateCfs{
		Name:        "линеаризация канала 1",
		Coefficient: kefCh1Lin0,
		Count:       4,
		Format:      floatBitsFormat,
		InterpolateCfsFunc: func(pv workparty.ProductValues) ([]numeth.Coordinate, error) {
			isCO := w.t.Chan[0].gasName.isCO2()
			gases := []gas{gas1, gas3, gas4}
			if isCO {
				gases = []gas{gas1, gas2, gas3, gas4}
			}
			var dt []xy
			for _, gas := range gases {
				key := chan1.keyLin(gas)
				y, ok := pv.Get(key)
				if !ok {
					return nil, merry.Errorf("нет значения %s", key)
				}
				dt = append(dt, xy{w.C[gas-1], y})
			}
			return dt, nil
		},
	}.Work()
}

func (w wrk) calcLin2() work {
	return workparty.InterpolateCfs{
		Name:        "линеаризация канала 2",
		Coefficient: kefCh2Lin0,
		Count:       4,
		Format:      floatBitsFormat,
		InterpolateCfsFunc: func(pv workparty.ProductValues) ([]numeth.Coordinate, error) {
			var dt []xy
			for _, gas := range []gas{gas1, gas5, gas6} {
				key := chan1.keyLin(gas)
				y, ok := pv.Get(key)
				if !ok {
					return nil, merry.Errorf("нет значения %s", key)
				}
				dt = append(dt, xy{w.C[gas-1], y})
			}
			return dt, nil
		},
	}.Work()
}

func write32(cmd modbus.DevCmd, value float64) workgui.WorkFunc {
	return workparty.Write32(cmd, floatBitsFormat, value)
}

func readAndSaveVar(dbKey string, Var modbus.Var) workgui.WorkFunc {
	return workparty.ReadAndSaveProductVar(Var, floatBitsFormat, dbKey)
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

func setProductWorkMode(log comm.Logger, ctx context.Context, product devdata.Product, mode float64) error {
	_, err := modbus.Request{
		Addr:     product.Addr,
		ProtoCmd: 0x16,
		Data:     append([]byte{0xA0, 0, 0, 2, 4}, modbus.BCD6(mode)...),
	}.GetResponse(log, ctx, comports.Comm(product.Comport, deviceConfig))
	return err
}

var (
	newFuncLst = workgui.NewWorkFuncFromList

	hardwareWarn = hardware.WithWarn{}
)
