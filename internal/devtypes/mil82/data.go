package mil82

import (
	"fmt"
	"github.com/fpawel/atool/internal/devtypes/devdata"
	"github.com/fpawel/comm/modbus"
)

func dataSections() (result devdata.DataSections) {

	type dataSection = devdata.DataSection

	addDs := func(ds dataSection) {
		result = append(result, ds)
	}

	addDs(dataSection{
		Name: "Коэффициенты",
		Params: func() (xs []devdata.DataParam) {
			for i := 0; i <= 100; i++ {
				xs = append(xs, devdata.DataParam{
					Key:  fmt.Sprintf("K%02d", i),
					Name: fmt.Sprintf("%d", i),
				})
			}
			return
		}(),
	})

	x := dataSection{Name: "Снятие: линеаризация"}
	for i := 1; i <= 4; i++ {
		x.Params = append(x.Params, devdata.DataParam{
			Key:  fmt.Sprintf("lin%d", i),
			Name: fmt.Sprintf("%d", i),
		})
	}
	addDs(x)

	tXs1 := func(gas int) (xs []devdata.DataParam) {
		for _, Var := range []int{2, 16} {
			for i, k := range []string{"t_low", "t_norm", "t_high"} {
				xs = append(xs, devdata.DataParam{
					Key:  fmt.Sprintf("%s_gas%d_var%d", k, gas, Var),
					Name: fmt.Sprintf("%d: Var%d", i, Var),
				})
			}
		}
		return
	}

	fmtVar := func(n modbus.Var) string {
		if s, f := paramsNames[n]; f {
			return s
		}
		return fmt.Sprintf("%d", n)
	}

	fmtGas := func(n int) string {
		return fmt.Sprintf("ПГС%d", n)
	}

	pts := [][2]string{
		{keyTempLow, "компенсация Т-"},
		{keyTempNorm, "компенсация НКУ"},
		{keyTempHigh, "компенсация Т+"},
		{keyTestTempNorm, "проверка погрешности: НКУ"},
		{keyTestTempLow, "проверка погрешности: Т-"},
		{keyTestTempHigh, "проверка погрешности: Т+"},
		{keyTestTemp80, "проверка погрешности: 80⁰C"},
		{keyTest2, "проверка погрешности: возврат НКУ"},
		{keyTex1, "перед техпрогоном"},
		{keyTex2, "после техпрогона"},
	}

	for _, ptT := range pts {
		pkKey := ptT[0]
		ptName := ptT[1]
		gases := []int{1, 3, 4}
		if pkKey == keyTestTempNorm {
			gases = []int{1, 2, 3, 4}
		}
		x = dataSection{Name: "Снятие: " + ptName}
		for _, Var := range vars {
			for _, gas := range gases {
				x.Params = append(x.Params, devdata.DataParam{
					Key:  fmt.Sprintf("%s_gas%d_var%d", pkKey, gas, Var),
					Name: fmt.Sprintf("%s: %s", fmtVar(Var), fmtGas(gas)),
				})
			}
		}
		addDs(x)
	}
	addDs(dataSection{
		Name:   "Расчёт термоконмпенсации начала шкалы",
		Params: tXs1(1),
	})

	addDs(dataSection{
		Name:   "Расчёт термоконмпенсации конца шкалы",
		Params: tXs1(4),
	})

	addDs(dataSection{
		Name:   "Расчёт термоконмпенсации середины шкалы",
		Params: tXs1(3),
	})

	return
}
