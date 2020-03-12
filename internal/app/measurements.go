package app

import (
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/data"
)

type measurements struct {
	xs []data.Measurement
}

func (x *measurements) add(ProductID int64, ParamAddr int, Value float64) {

	x.xs = append(x.xs, data.NewMeasurement(ProductID, ParamAddr, Value))
	if len(x.xs) >= 1000 {
		x.Save()
		x.xs = nil
	}
}

func (x *measurements) Save() {
	if err := data.SaveMeasurements(x.xs); err != nil {
		log.PrintErr(merry.Append(err, "fail to insert measurements"))
	}
}
