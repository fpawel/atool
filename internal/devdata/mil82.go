package devdata

import (
	"fmt"
	"github.com/fpawel/atool/internal/devdata/calcmil82"
)

func init() {
	mil82 := Device{
		Calc: calcmil82.Calc,
	}

	type ds = DataSection

	addDs := func(ds ds) {
		mil82.DataSections = append(mil82.DataSections, ds)
	}

	x := ds{Name: "Линеаризация"}
	for i := 1; i <= 4; i++ {
		x.Params = append(x.Params, DataParam{
			Key:  fmt.Sprintf("lin%d", i),
			Name: fmt.Sprintf("%d", i),
		})
	}
	addDs(x)

	tXs1 := func(gas int) (xs []DataParam) {
		for _, Var := range []int{2, 16} {
			for i, k := range []string{"t_low", "t_norm", "t_high"} {
				xs = append(xs, DataParam{
					Key:  fmt.Sprintf("%s_gas%d_var%d", k, gas, Var),
					Name: fmt.Sprintf("%d: Var%d", i, Var),
				})
			}
		}
		return
	}

	addDs(ds{
		Name:   "Термоконмпенсация начала шкалы",
		Params: tXs1(1),
	})

	addDs(ds{
		Name:   "Термоконмпенсация конца шкалы",
		Params: tXs1(4),
	})

	addDs(ds{
		Name:   "Термоконмпенсация середины шкалы",
		Params: tXs1(3),
	})

	fmtVar := func(n int) string {
		if s, f := (map[int]string{
			0:  "C",
			2:  "T",
			4:  "I",
			12: "Work",
			14: "Ref",
		})[n]; f {
			return s
		}
		return fmt.Sprintf("%d", n)
	}

	fmtGas := func(n int) string {
		return fmt.Sprintf("ПГС%d", n)
	}

	pts := map[string]string{
		"t_low":       "Низкая температура",
		"t_norm":      "Нормальная температура",
		"t_high":      "Высокая температура",
		"test_t_norm": "Проверка погрешности: НКУ",
		"test_t_low":  "Проверка погрешности: низкая температура",
		"test_t_high": "Проверка погрешности: высокая температура",
		"test2":       "Проверка погрешности: возврат НКУ",
		"tex1":        "Перед техпрогоном",
		"tex2":        "После техпрогона",
	}
	vars := []int{0, 2, 4, 8, 10, 12, 14, 16}

	for pkKey, ptName := range pts {
		x = ds{Name: ptName}
		for _, Var := range vars {
			for _, gas := range []int{1, 3, 4} {
				x.Params = append(x.Params, DataParam{
					Key:  fmt.Sprintf("%s_gas%d_var%d", pkKey, gas, Var),
					Name: fmt.Sprintf("%s: %s", fmtVar(Var), fmtGas(gas)),
				})
			}
		}
		addDs(x)
	}
	Devices["МИЛ-82"] = mil82
}
