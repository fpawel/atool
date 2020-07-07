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

	x := ds{Name: "линеаризация: канал 1"}
	for i := 1; i <= 4; i++ {
		x.Params = append(x.Params, devdata.DataParam{
			Key:  fmt.Sprintf("lin%d_chan1", i),
			Name: fmt.Sprintf("канал 1 газ %d", i),
		})
	}
	x.Params = append(x.Params, devdata.DataParam{
		Key:  fmt.Sprintf("lin%d_chan2", 1),
		Name: fmt.Sprintf("канал 2 газ %d", 1),
	})
	for i := 5; i <= 6; i++ {
		x.Params = append(x.Params, devdata.DataParam{
			Key:  fmt.Sprintf("lin%d_chan2", i),
			Name: fmt.Sprintf("канал 2 газ %d", i),
		})
	}
	addDs(x)

	return
}()
