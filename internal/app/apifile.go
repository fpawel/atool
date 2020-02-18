package app

import (
	"context"
	"fmt"
	"github.com/fpawel/atool/internal/config/configlua"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"sort"
)

type fileSvc struct{}

var _ api.FileService = new(fileSvc)

func (h *fileSvc) GetProductsValues(_ context.Context, partyID int64) (*apitypes.PartyProductsValues, error) {

	result := new(apitypes.PartyProductsValues)

	party, err := data.GetParty(db, partyID)
	if err != nil {
		return nil, err
	}

	for _, p := range party.Products {
		result.Products = append(result.Products, convertDataProductToApiProduct(p))
	}

	d, _ := configlua.Devices[party.DeviceType]

	values, err := getPartyProductsValues(party.PartyID)
	if err != nil {
		return nil, err
	}

	for _, sect := range d.Sections {
		y := &apitypes.SectionProductParamsValues{
			Section: sect.Name,
		}

		for _, prm := range sect.Params {
			y.Keys = append(y.Keys, prm.Key)
		}

		for _, prm := range sect.Params {
			xs := []string{prm.Name}
			for _, p := range party.Products {
				var s string
				if m, f := values[prm.Key]; f {
					if v, f := m[p.ProductID]; f {
						s = fmt.Sprintf("%v", v)
					}
				}
				xs = append(xs, s)
			}
			y.Values = append(y.Values, xs)
		}

		result.Sections = append(result.Sections, y)
	}

	sectAll := getPartyProductsValuesAll(party, values)
	result.Sections = append(result.Sections, &sectAll)

	return result, nil
}

func getPartyProductsValuesAll(party data.Party, values mapStrIntFloat) apitypes.SectionProductParamsValues {

	result := apitypes.SectionProductParamsValues{
		Section: "Все сохранённые значения",
	}

	for k, m := range values {
		xs := []string{k}
		for _, p := range party.Products {
			s := ""
			if v, f := m[p.ProductID]; f {
				s = fmt.Sprintf("%v", v)
			}
			xs = append(xs, s)
		}
		if len(xs) > 1 {
			result.Values = append(result.Values, xs)
			result.Keys = append(result.Keys, k)
		}
	}

	if len(result.Values) > 1 {
		sort.Slice(result.Keys, func(i, j int) bool {
			return result.Keys[i] < result.Keys[j]
		})
		vs := result.Values[1:]
		sort.Slice(vs, func(i, j int) bool {
			return vs[i][0] < vs[j][0]
		})
	}

	return result
}

func getPartyProductsValues(partyID int64) (mapStrIntFloat, error) {
	const q2 = `
SELECT product_id, key, value
FROM product_value
WHERE product_id IN (SELECT product_id FROM product WHERE party_id = ?)`
	var values1 []struct {
		ProductID int64   `db:"product_id"`
		Key       string  `db:"key"`
		Value     float64 `db:"value"`
	}
	if err := db.Select(&values1, q2, partyID); err != nil {
		return nil, err
	}

	values := map[string]mapIntFloat{}

	for _, x := range values1 {
		if values[x.Key] == nil {
			values[x.Key] = mapIntFloat{}
		}
		values[x.Key][x.ProductID] = x.Value
	}
	return values, nil
}

type mapStrIntFloat = map[string]mapIntFloat
