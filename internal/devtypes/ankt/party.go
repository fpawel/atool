package ankt

import (
	"fmt"
	"github.com/fpawel/atool/internal/data"
)

func initParty() error {
	party, err := data.GetCurrentParty()
	if err != nil {
		return err
	}
	pv, err := data.GetPartyValues1(party.PartyID)
	if err != nil {
		return err
	}

	Type, ok := productTypes[party.ProductType]
	if !ok {
		Type = productTypesList[0]
		if _, err := data.DB.Exec(
			`UPDATE party SET product_type = ? WHERE party_id = (SELECT party_id FROM app_config)`,
			Type.String()); err != nil {
			return err
		}
	}

	xs := []float64{0, 25, 50, 100, 50, 100}
	for i, v := range xs {
		key := fmt.Sprintf("c%d", i+1)
		if _, f := pv[key]; !f {
			if err := data.SetCurrentPartyValue(key, v); err != nil {
				return err
			}
		}
	}

	if err := data.SetCurrentPartyValue(keyTempNorm.String(), 20); err != nil {
		return err
	}
	if err := data.SetCurrentPartyValue(keyTempLow.String(), -60); err != nil {
		return err
	}
	if err := data.SetCurrentPartyValue(keyTempHigh.String(), 80); err != nil {
		return err
	}

	return nil
}
