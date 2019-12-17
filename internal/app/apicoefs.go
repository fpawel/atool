package app

import (
	"context"
	"github.com/fpawel/atool/internal/cfg"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
)

type coefficientsSvc struct {
}

var _ api.CoefficientsService = new(coefficientsSvc)

func (*coefficientsSvc) WriteAll(context.Context) error {
	return nil
}

func (*coefficientsSvc) ReadAll(context.Context) error {
	runReadAllCoefficients()
	return nil
}

func (h *coefficientsSvc) ListCoefficients(_ context.Context) (r []*apitypes.Coefficient, err error) {
	c := cfg.Get()
	for _, i := range c.ListCoefficients() {
		_, inactive := c.InactiveCoefficients[i]
		r = append(r, &apitypes.Coefficient{
			N:      int32(i),
			Active: !inactive,
		})
	}
	return
}

func (h *coefficientsSvc) SetActive(_ context.Context, n int32, active bool) (err error) {
	c := cfg.Get()
	if active {
		delete(c.InactiveCoefficients, int(n))
	} else {
		c.InactiveCoefficients[int(n)] = struct{}{}
	}
	return cfg.Set(c)
}
