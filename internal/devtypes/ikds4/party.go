package ikds4

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

	if _, f := pv[keyLinearDegree]; !f {
		if err := data.SetCurrentPartyValue(keyLinearDegree, 4); err != nil {
			return err
		}
	}

	for i, v := range []float64{0, 25, 50, 100} {
		key := fmt.Sprintf("c%d", i+1)
		if _, f := pv[key]; !f {
			if err := data.SetCurrentPartyValue(key, v); err != nil {
				return err
			}
		}
	}
	Type, ok := prodTypes[party.ProductType]
	if !ok {
		Type = prodTypesList[0]
		if _, err := data.DB.Exec(
			`UPDATE party SET product_type = ? WHERE party_id = (SELECT party_id FROM app_config)`,
			Type.Name); err != nil {
			return err
		}
	}

	if err := data.SetCurrentPartyValue(keyTempNorm, 20); err != nil {
		return err
	}
	if err := data.SetCurrentPartyValue(keyTempLow, -40); err != nil {
		return err
	}
	if err := data.SetCurrentPartyValue(keyTempHigh, 50); err != nil {
		return err
	}
	return nil
}
