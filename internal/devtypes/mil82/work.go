package mil82

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/comm"
)

func mainWork() error {
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

	var w wrk

	for i := 1; i < 5; i++ {
		k := fmt.Sprintf("c%d", i)
		c, ok := pv[k]
		if !ok {
			return merry.Errorf("нет значения ПГС%d", i)
		}
		w.C[i] = c
	}

	var ok bool
	w.Type, ok = prodTypes[party.ProductType]
	if !ok {
		return merry.Errorf("%s не правильное исполнение %s", party.DeviceType, party.ProductType)
	}

	return nil
}

type wrk struct {
	C    [4]float64
	Type productType
	log  comm.Logger
	ctx  context.Context
}
