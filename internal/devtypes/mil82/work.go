package mil82

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/hardware"
	"github.com/fpawel/atool/internal/pkg/numeth"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/atool/internal/workparty"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/powerman/structlog"
	"time"
)

type Work = workgui.Work

func work(log comm.Logger, ctx context.Context) error {
	w := new(wrk)
	if err := w.init(); err != nil {
		return err
	}
	return w.mainWork(log, ctx)
}

type wrk struct {
	C            [4]float64
	Type         productType
	ks           KefValueMap
	temps        map[string]float64
	linearDegree linearDegree
	warn         hardware.WithWarn
}

const (
	floatBitsFormat = modbus.BCD
	linearDegree3   = 3
	linearDegree4   = 4
)

func (x *wrk) init() error {
	*x = wrk{
		ks:    make(KefValueMap),
		temps: make(map[string]float64),
	}
	party, err := data.GetCurrentParty()
	if err != nil {
		return err
	}
	if party.DeviceType != "МИЛ-82" {
		return merry.Errorf("нельзя выполнить настройку МИЛ-82 для %s", party.DeviceType)
	}
	pv, err := data.GetPartyValues1(party.PartyID)
	if err != nil {
		return err
	}

	lnDgr, ok := pv[keyLinearDegree]
	if !ok {
		return merry.New("не указана степень линеаризации")
	}
	x.linearDegree = linearDegree(lnDgr)
	if err := x.linearDegree.validate(); err != nil {
		return err
	}

	x.Type, ok = prodTypes[party.ProductType]
	if !ok {
		return merry.Errorf("%s не правильное исполнение %s", party.DeviceType, party.ProductType)
	}

	Tnorm, ok := pv[keyTempNorm]
	if !ok {
		return merry.Errorf("нет значения %q", keyTempNorm)
	}
	Tlow, ok := pv[keyTempLow]
	if !ok {
		return merry.Errorf("нет значения %q", keyTempLow)
	}
	Thigh, ok := pv[keyTempHigh]
	if !ok {
		return merry.Errorf("нет значения %q", keyTempHigh)
	}
	T80, ok := pv[keyTestTemp80]
	if !ok {
		return merry.Errorf("нет значения %q", keyTestTemp80)
	}
	x.temps[keyTempNorm] = Tnorm
	x.temps[keyTempLow] = Tlow
	x.temps[keyTempHigh] = Thigh
	x.temps[keyTestTemp80] = T80

	for i := 1; i < 5; i++ {
		k := fmt.Sprintf("c%d", i)
		c, ok := pv[k]
		if !ok {
			return merry.Errorf("нет значения ПГС%d", i)
		}
		x.C[i] = c
	}

	x.ks = KefValueMap{
		2:  float64(time.Now().Year()),
		8:  x.Type.Scale0,
		9:  x.Type.Scale,
		10: x.C[0],
		11: x.C[3],
		16: 0,
		17: 1,
		18: 0,
		19: 0,
		23: 0,
		24: 0,
		25: 0,
		26: 1,
		27: 0,
		28: 0,
		37: 1,
		38: 0,
		39: 0,

		47: encode2(float64(time.Now().Month()), float64(x.Type.Index)),
	}

	for k, v := range x.Type.Kef {
		x.ks[k] = v
	}

	x.ks[5], ok = map[string]float64{
		"CO2":   7,
		"CH4":   14,
		"C3H8":  14,
		"C6H14": 14,
	}[x.Type.Gas]
	if !ok {
		return merry.Errorf("%s %s: нет кода единиц измерения для газа %s", party.DeviceType, party.ProductType, x.Type.Gas)
	}

	x.ks[6], ok = map[string]float64{
		"CO2":   4,
		"CH4":   5,
		"C3H8":  7,
		"C6H14": 7,
	}[x.Type.Gas]
	if !ok {
		return merry.Errorf("%s %s: нет кода газа %s", party.DeviceType, party.ProductType, x.Type.Gas)
	}
	x.ks[7], ok = map[float64]float64{
		4:   57,
		10:  7,
		20:  9,
		50:  0,
		100: 21,
	}[x.Type.Scale]
	if !ok {
		return merry.Errorf("%s %s: нет кода шкалы %v", party.DeviceType, party.ProductType, x.Type.Scale)
	}

	return nil
}

func (x *wrk) holdTemperature(pt string) workgui.WorkFunc {
	t, ok := x.temps[pt]
	if !ok {
		panic(fmt.Errorf("не правильная точка температуры %q", pt))
	}
	return x.warn.HoldTemperature(t)
}

func (x *wrk) mainWork(log *structlog.Logger, ctx context.Context) error {
	type lst = workgui.WorkFuncList
	works := workgui.Works{
		{
			"запись коэффициентов",
			x.writeProductsCoefficients,
		},
		{
			"установка НКУ",
			x.holdTemperature(keyTempNorm),
		},
		{
			"нормировка",
			lst{
				x.warn.BlowGas(1),
				x.write32(8, 10000),
			}.Do,
		},
		x.adjust(),
		{
			"снятие линеаризации",
			x.linRead(),
		},
		x.linCalcAndWrite(),
		x.readSaveTempComp(keyTempLow),
		x.readSaveTempComp(keyTempHigh),
		x.readSaveTempComp(keyTempNorm),
		workgui.NewWorks("расчёт и запись термокомпенсации",
			x.calcWriteT0(),
			x.calcWriteTK(),
			x.calcWriteTM()),
		workgui.New("снятие сигналов каналов", workparty.ReadCfs(workparty.CfsList{20, 21, 43, 44}, floatBitsFormat)),
		workgui.New("НКУ: снятие для проверки погрешности", workgui.WorkFuncList{
			x.warn.HoldTemperature(x.temps[keyTempNorm]),
			x.adjust().Func,
			x.readSaveSectionGases(keyTestTempNorm, []byte{1, 2, 3, 4}),
		}.Do),
		workgui.New(fmt.Sprintf("Т-: снятие для проверки погрешности: %g⁰C", x.temps[keyTempLow]),
			x.readSaveTemp(keyTestTempLow)),
		workgui.New(fmt.Sprintf("Т+: снятие для проверки погрешности: %g⁰C", x.temps[keyTempHigh]),
			x.readSaveTemp(keyTestTempHigh)),
		workgui.New(fmt.Sprintf("90⁰C: снятие для проверки погрешности: %g⁰C", x.temps[keyTestTemp80]),
			x.readSaveTemp(keyTestTemp80)),
		workgui.New("НКУ: повторное снятие для проверки погрешности",
			x.readSaveTemp(keyTest2)),
	}.ExecuteSelectWorksDialog(ctx.Done())
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return works.Work("Настройка МИЛ-82").Run(log, ctx)
}

func (x *wrk) adjust() workgui.Work {
	type lst = workgui.WorkFuncList
	return workgui.NewWorks("калибровка",
		Work{
			Name: "калибровка нуля",
			Func: lst{
				x.warn.BlowGas(1),
				x.write32(1, x.C[1]),
			}.Do,
		},
		Work{
			Name: "калибровка чувствительности",
			Func: lst{
				x.warn.BlowGas(4),
				x.write32(2, x.C[3]),
				x.warn.BlowGas(1),
			}.Do,
		})
}

func (x *wrk) calcWriteT0() workgui.Work {
	return workparty.InterpolateCfs{
		Name:        "T0 начало шкалы",
		Coefficient: 23,
		Count:       3,
		Format:      floatBitsFormat,
		InterpolateCfsFunc: func(pv1 workparty.ProductValues) ([]xy, error) {
			pv := productValues{pv1}
			t1 := pv.getTempValuesNaN(1, varTemp)
			var1 := pv.getTempValuesNaN(1, var16)
			return []xy{
				{t1[0], -var1[0]},
				{t1[1], -var1[1]},
				{t1[2], -var1[2]},
			}, nil
		},
	}.Work()
}

func (x *wrk) calcWriteTK() workgui.Work {
	return workparty.InterpolateCfs{
		Name:        "TK конец шкалы",
		Coefficient: 26,
		Count:       3,
		Format:      floatBitsFormat,
		InterpolateCfsFunc: func(pv1 workparty.ProductValues) ([]xy, error) {
			pv := productValues{pv1}
			t4 := pv.getTempValuesNaN(4, varTemp)
			var4 := pv.getTempValuesNaN(4, var16)
			var1 := pv.getTempValuesNaN(1, var16)
			xy := make([]numeth.Coordinate, 3)
			for i := 0; i < 3; i++ {
				xy[i] = numeth.Coordinate{t4[i], (var4[2] - var1[2]) / (var4[i] - var1[i])}
			}
			return xy, nil
		},
	}.Work()
}

func (x *wrk) calcWriteTM() workgui.Work {
	return workparty.InterpolateCfs{
		Name:        "TM середина шкалы",
		Coefficient: 37,
		Count:       3,
		Format:      floatBitsFormat,
		InterpolateCfsFunc: func(pv1 workparty.ProductValues) ([]xy, error) {
			pv := productValues{pv1}
			C4 := x.C[3]

			K16 := pv.KefNaN(16)
			K17 := pv.KefNaN(17)
			K18 := pv.KefNaN(18)
			K19 := pv.KefNaN(19)

			v1Norm := pv.tempValNaN(keyTempNorm, 1, var16)
			v3Norm := pv.tempValNaN(keyTempNorm, 3, var16)
			v4Norm := pv.tempValNaN(keyTempNorm, 4, var16)

			v1Low := pv.tempValNaN(keyTempLow, 1, var16)
			v3Low := pv.tempValNaN(keyTempLow, 3, var16)
			v4Low := pv.tempValNaN(keyTempLow, 4, var16)

			v1High := pv.tempValNaN(keyTempHigh, 1, var16)
			v3High := pv.tempValNaN(keyTempHigh, 3, var16)
			v4High := pv.tempValNaN(keyTempHigh, 4, var16)

			x1 := C4 * (v1Norm - v3Norm) / (v1Norm - v4Norm)
			x2 := C4 * (v1Low - v3Low) / (v1Low - v4Low)
			d := K16 + K17*x2 + K18*x2*x2 + K19*x2*x2*x2 - x2

			yLow := (K16 + K17*x1 + K18*x1*x1 + K19*x1*x1*x1 - x2) / d

			x1 = C4 * (v1Norm - v3Norm) / (v1Norm - v4Norm)
			x2 = C4 * (v1High - v3High) / (v1High - v4High)

			d = K16 + K17*x2 + K18*x2*x2 + K19*x2*x2*x2 - x2

			yHi := (K16 + K17*x1 + K18*x1*x1 + K19*x1*x1*x1 - x2) / d

			t1 := pv.tempValNaN(keyTempLow, 3, varTemp)
			t2 := pv.tempValNaN(keyTempNorm, 3, varTemp)
			t3 := pv.tempValNaN(keyTempHigh, 3, varTemp)

			return []xy{
				{t1, yLow},
				{t2, 1},
				{t3, yHi},
			}, nil
		},
	}.Work()
}

func (x *wrk) readSaveTempComp(ptTemp string) workgui.Work {
	return workgui.New(fmt.Sprintf("cнятие термокомпенсации Т-: %g⁰C", x.temps[ptTemp]),
		x.readSaveTemp(ptTemp))
}

func (x *wrk) readSaveTemp(ptTemp string) workgui.WorkFunc {
	return workgui.WorkFuncList{
		x.warn.HoldTemperature(x.temps[ptTemp]),
		x.readSaveSectionGases(ptTemp, []byte{1, 3, 4}),
	}.Do
}

func (x *wrk) readSaveSectionGases(dbKeySection string, gases []byte) workgui.WorkFunc {
	works := make(workgui.Works, 0)
	for _, gas := range gases {
		worksGas := workgui.WorkFuncList{
			x.warn.BlowGas(gas),
		}
		for _, Var := range vars {
			worksGas = append(worksGas, workparty.ReadAndSaveProductParam(Var, floatBitsFormat, dbKeySectionGasVar(dbKeySection, gas, Var)))
		}

		works = append(works, workgui.New(fmt.Sprintf("снятие %s газ %d", dbKeySection, gas), worksGas.Do))
	}
	return works.Work(fmt.Sprintf("снятие %s газы %v", dbKeySection, gases)).Perform
}

func (x *wrk) writeProductsCoefficients(log *structlog.Logger, ctx context.Context) error {
	xs, err := x.getProductsCoefficients()
	if err != nil {
		return err
	}
	return workparty.WriteProdsCfs(xs, nil)(log, ctx)
}

func (x *wrk) getProductsCoefficients() ([]workparty.ProductCoefficientValue, error) {
	var xs []workparty.ProductCoefficientValue
	products, err := data.GetActiveProducts()
	if err != nil {
		return nil, err
	}
	for _, p := range products {
		ks := copyKefValueMap(x.ks)
		ks[40] = encode2(float64(time.Now().Year()-2000), float64(p.Serial))
		for k, v := range ks {
			xs = append(xs, workparty.ProductCoefficientValue{
				ProductID:   p.ProductID,
				Coefficient: k,
				Value:       v,
			})
		}
	}
	return xs, nil
}

func (x *wrk) write32(cmd modbus.DevCmd, value float64) workgui.WorkFunc {
	return workparty.Write32(cmd, floatBitsFormat, value)
}

func (x *wrk) linRead() workgui.WorkFunc {
	var lin workgui.WorkFuncList
	for _, gas := range x.linGases() {
		lin = append(lin,
			x.warn.BlowGas(gas),
			workparty.ReadAndSaveProductParam(0, floatBitsFormat, dbKeyLin(gas)))
	}
	return lin.Do
}

func (x *wrk) linCalcAndWrite() workgui.Work {
	return workparty.InterpolateCfs{
		Name:        "расчёт и запись линеаризации",
		Coefficient: 16,
		Count:       4,
		Format:      floatBitsFormat,
		InterpolateCfsFunc: func(pv workparty.ProductValues) ([]numeth.Coordinate, error) {
			var dt []xy
			for _, gas := range x.linGases() {
				key := dbKeyLin(gas)
				y, ok := pv.Get(key)
				if !ok {
					return nil, merry.Errorf("нет значения %s", key)
				}
				dt = append(dt, xy{x.C[gas-1], y})
			}
			return dt, nil
		},
	}.Work()
}

func (x *wrk) linGases() []byte {
	linGases := []byte{1, 3, 4}
	if x.linearDegree == linearDegree4 {
		linGases = []byte{1, 2, 3, 4}
	}
	return linGases
}

func copyKefValueMap(x KefValueMap) KefValueMap {
	y := make(KefValueMap)
	for k, v := range x {
		y[k] = v
	}
	return y
}

func encode2(a, b float64) float64 {
	return a*10000 + b
}

type linearDegree float64

type xy = [2]float64

func (x linearDegree) validate() error {
	switch x {
	case linearDegree3, linearDegree4:
		return nil
	default:
		return merry.Errorf("unexpcpected linear degree value: %s", x)
	}
}

func dbKeySectionGasVar(dbKeySection string, gas byte, Var modbus.Var) string {
	return fmt.Sprintf("%s_gas%d_var%d", dbKeySection, gas, Var)
}

func dbKeyLin(gas byte) string {
	return fmt.Sprintf("lin%d", gas)
}

type productValues struct {
	workparty.ProductValues
}

func (x productValues) tempValNaN(ptT string, gas byte, Var modbus.Var) float64 {
	return x.GetNaN(dbKeySectionGasVar(ptT, gas, Var))
}

func (x productValues) getTempValuesNaN(gas byte, Var modbus.Var) []float64 {
	var keys []string
	for _, ptT := range ptsTemp {
		keys = append(keys, dbKeySectionGasVar(ptT, gas, Var))
	}
	return x.GetValuesNaN(keys)
}
