package ankt

import (
	"fmt"
	"github.com/fpawel/atool/internal/devtypes/devdata"
)

var dataSections = func() (result devdata.DataSections) {
	type ds = devdata.DataSection

	addDs := func(ds ds) {
		result = append(result, ds)
	}

	addDs(ds{
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

	x := ds{Name: "линеаризация"}
	for gas := gas1; gas <= gas4; gas++ {
		x.Params = append(x.Params, chan1.dataParamLin(gas))
	}
	x.Params = append(x.Params, chan2.dataParamLin(gas1))
	for gas := gas5; gas <= gas6; gas++ {
		x.Params = append(x.Params, chan2.dataParamLin(gas))
	}
	addDs(x)

	return
}()
