package app

import (
	"context"
	"fmt"
	"github.com/fpawel/atool/internal/config/appcfg"
	"github.com/fpawel/atool/internal/config/devicecfg"
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
	return runWork(workparty.NewWorkWriteAllCfs(in))
}

func (*coefficientsSvc) ReadAll(context.Context) error {
	return runWork(workparty.NewWorkReadAllCfs())
}

func (h *coefficientsSvc) ListCoefficients(_ context.Context) (r []*apitypes.Coefficient, err error) {
	r = []*apitypes.Coefficient{}
	party, err := data.GetCurrentParty()
	if err != nil {
		return nil, err
	}
	device, err := appcfg.GetDeviceByName(party.DeviceType)
	if err != nil {
		return nil, err
	}
	for _, i := range device.Config.ListCoefficients() {
		_, inactive := appcfg.Sets.InactiveCoefficients[i]

		kef := &apitypes.Coefficient{
			N:      int32(i),
			Active: !inactive,
			Name:   fmt.Sprintf("%d", i),
		}
		if device.Config.CfsNames != nil {
			name, fName := device.Config.CfsNames[i]
			if fName {
				kef.Name = fmt.Sprintf("%d %s", i, name)
			}
		}

		r = append(r, kef)
	}
	return
}

func (h *coefficientsSvc) SetActive(_ context.Context, n int32, active bool) (err error) {
	if active {
		delete(appcfg.Sets.InactiveCoefficients, devicecfg.Kef(n))
	} else {
		appcfg.Sets.InactiveCoefficients[devicecfg.Kef(n)] = struct{}{}
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
