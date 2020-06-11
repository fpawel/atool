package app

import (
	"context"
	"fmt"
	"github.com/fpawel/atool/internal/config/appcfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/devtypes/devdata"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"sort"
)

type fileSvc struct{}

var _ api.FileService = new(fileSvc)

func (h *fileSvc) GetProductsValues(_ context.Context, partyID int64, filterSerial int64) (*apitypes.PartyProductsValues, error) {

	result := new(apitypes.PartyProductsValues)

	var party data.PartyValues

	if err := data.GetPartyValues(partyID, &party, filterSerial); err != nil {
		return nil, err
	}

	sectAll := getPartyProductsValuesAll(party)
	result.Sections = append(result.Sections, &sectAll)

	for _, p := range party.Products {
		result.Products = append(result.Products, convertDataProductValuesToApiProduct(party, p))
	}

	device, _ := appcfg.DeviceTypes[party.DeviceType]

	//sort.Slice(device.DataSections, func(i, j int) bool {
	//	return device.DataSections[i].Name < device.DataSections[j].Name
	//})

	for _, sect := range device.DataSections {
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
				if v, f := p.Values[prm.Key]; f {
					s = fmt.Sprintf("%v", v)
				}
				xs = append(xs, s)
			}
			y.Values = append(y.Values, xs)
		}

		result.Sections = append(result.Sections, y)
	}

	var concentrationErrorSections devdata.CalcSections
	if device.Calc != nil {
		if err := device.Calc(party, &concentrationErrorSections); err != nil {
			result.CalcError = err.Error()
		} else {
			result.Calc = concentrationErrorSections
		}
	}

	return result, nil
}

func getPartyProductsValuesAll(party data.PartyValues) apitypes.SectionProductParamsValues {

	result := apitypes.SectionProductParamsValues{
		Section: "Все сохранённые значения",
	}

	xs := make(map[string]map[int64]float64)
	for _, p := range party.Products {

		for k, v := range p.Values {
			m, f := xs[k]
			if !f {
				m = make(map[int64]float64)
				xs[k] = m
			}
			m[p.ProductID] = v
		}
	}

	for k, m := range xs {
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

		sort.Slice(result.Values, func(i, j int) bool {
			return result.Values[i][0] < result.Values[j][0]
		})
	}

	return result
}

func convertDataProductValuesToApiProduct(party data.PartyValues, p data.ProductValues) *apitypes.Product {
	return &apitypes.Product{
		ProductID:      p.ProductID,
		PartyID:        party.PartyID,
		PartyCreatedAt: timeUnixMillis(party.CreatedAt),
		Addr:           int8(p.Addr),
		Serial:         int64(p.Serial),
	}
}
