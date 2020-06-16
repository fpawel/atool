package app

import (
	"context"
	"github.com/fpawel/atool/internal/config/appcfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/atool/internal/workparty"
	"regexp"
	"strconv"
)

type coefficientsSvc struct {
}

var _ api.CoefficientsService = new(coefficientsSvc)

func (*coefficientsSvc) WriteAll(_ context.Context, in []*apitypes.ProductCoefficientValue) error {
	return workparty.NewWorkWriteAllCfs(log, appCtx, in)
}

func (*coefficientsSvc) ReadAll(context.Context) error {
	return workparty.NewWorkReadCfs(log, appCtx)
}

func (h *coefficientsSvc) ListCoefficients(_ context.Context) (r []*apitypes.Coefficient, err error) {
	r = []*apitypes.Coefficient{}
	d, _ := getCurrentPartyDeviceConfig()
	for _, i := range d.ListCoefficients() {
		_, inactive := appcfg.Cfg.InactiveCoefficients[i]
		r = append(r, &apitypes.Coefficient{
			N:      int32(i),
			Active: !inactive,
		})
	}
	return
}

func (h *coefficientsSvc) SetActive(_ context.Context, n int32, active bool) (err error) {
	if active {
		delete(appcfg.Cfg.InactiveCoefficients, int(n))
	} else {
		appcfg.Cfg.InactiveCoefficients[int(n)] = struct{}{}
	}
	return nil
}

func (h *coefficientsSvc) GetCurrentPartyCoefficients(_ context.Context) ([]*apitypes.ProductCoefficientValue, error) {

	var xs []struct {
		ProductID int64   `db:"product_id"`
		Key       string  `db:"key"`
		Value     float64 `db:"value"`
	}
	err := data.DB.Select(&xs, `
WITH q1 AS (SELECT party_id FROM app_config LIMIT 1)
SELECT p.product_id, key, value
FROM product_value
         INNER JOIN product p on product_value.product_id = p.product_id
WHERE party_id = (SELECT party_id FROM q1)
  AND key LIKE 'K%'
`)
	if err != nil {
		return nil, err
	}

	result := make([]*apitypes.ProductCoefficientValue, 0)
	for _, x := range xs {
		xs := regexp.MustCompile(`^K(\d+)$`).FindStringSubmatch(x.Key)
		if len(xs) < 2 {
			continue
		}
		k, err := strconv.Atoi(xs[1])
		if err != nil {
			continue
		}
		result = append(result, &apitypes.ProductCoefficientValue{
			ProductID:   x.ProductID,
			Coefficient: int32(k),
			Value:       x.Value,
		})
	}
	return result, nil
}
