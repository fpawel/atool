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

func work(log comm.Logger, ctx context.Context) error {
	w, err := newWrk()
	if err != nil {
		return err
	}
	return w.mainWork(log, ctx)
}

type wrk struct {
	C            [4]float64
	Type         productType
	ks           KefValueMap
	temps        map[ptTemp]float64
	linearDegree linearDegree
	warn         hardware.WithWarn
}

type ptTemp string

const (
	tLow            = "t_low"
	tHigh           = "t_high"
	tNorm           = "t_norm"
	floatBitsFormat = modbus.BCD
	linearDegree3   = 3
	linearDegree4   = 4
)

func (x *wrk) init() error {
	*x = wrk{
		ks:    make(KefValueMap),
		temps: make(map[ptTemp]float64),
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
	x.temps[tNorm] = Tnorm
	x.temps[tLow] = Tlow
	x.temps[tHigh] = Thigh

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

func (x *wrk) holdTemperature(pt ptTemp) workgui.WorkFunc {
	t, ok := x.temps[pt]
	if !ok {
		panic(fmt.Errorf("не правильная точка температуры %q", pt))
	}
	return x.warn.HoldTemperature(t)
}

func (x *wrk) mainWork(log *structlog.Logger, ctx context.Context) error {
	type lst = workgui.WorkFuncList

	var lin workgui.WorkFuncList

	linGases := []byte{1, 3, 4}
	if x.linearDegree == linearDegree4 {
		linGases = []byte{1, 2, 3, 4}
	}

	for _, gas := range linGases {
		lin = append(lin,
			x.warn.BlowGas(gas),
			workparty.ReadAndSaveProductParam(0, floatBitsFormat, fmt.Sprintf("lin%d", gas)))
	}

	works := workgui.Works{
		{
			"запись коэффициентов",
			x.writeProductsCoefficients,
		},
		{
			"установка НКУ",
			x.holdTemperature(tNorm),
		},
		{
			"нормировка",
			lst{
				x.warn.BlowGas(1),
				x.write32(8, 10000),
			}.Do,
		},
		{
			"калибровка нуля",
			lst{
				x.warn.BlowGas(1),
				x.write32(1, x.C[1]),
			}.Do,
		},
		{
			"калибровка чувствительности",
			lst{
				x.warn.BlowGas(4),
				x.write32(2, x.C[3]),
				x.warn.BlowGas(1),
			}.Do,
		},
		{
			"снятие линеаризации",
			lin.Do,
		},
		{
			"расчёт и запись линеаризации",
			func(log comm.Logger, ctx context.Context) error {
				workparty.ProcessEachActiveProduct(nil, func(log comm.Logger, ctx context.Context, product workparty.Product) error {
					type xy = numeth.Coordinate
					var dt []xy
					for _, gas := range linGases {
						dt = append(dt, xy{x.C[gas-1]})
					}
					product.InterpolationCoefficients("lin", dt, 16, 4, floatBitsFormat)

				})
			},
		},
	}.ExecuteSelectWorksDialog(ctx.Done())
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return works.Work("Настройка МИЛ-82").Run(log, ctx)
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

func (x linearDegree) validate() error {
	switch x {
	case linearDegree3, linearDegree4:
		return nil
	default:
		return merry.Errorf("unexpcpected linear degree value: %s", x)
	}
}
