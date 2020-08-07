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

	for _, key := range []string{
		"d_chan1_sc_beg_t_norm",
		"d_chan1_sc_mid_t_norm",
		"d_chan1_sc_end_t_norm",
		"d_chan1_sc_beg_t_low",
		"d_chan1_sc_mid_t_low",
		"d_chan1_sc_end_t_low",
		"d_chan1_sc_beg_t_high",
		"d_chan1_sc_mid_t_high",
		"d_chan1_sc_end_t_high",
	} {
		if _, f := pv[key]; !f {
			if err := data.SetCurrentPartyValue(key, 1); err != nil {
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

	if _, f := pv[keySendCmdSetWorkModeIntervalSec]; !f {
		if err := data.SetCurrentPartyValue(keySendCmdSetWorkModeIntervalSec, 10); err != nil {
			return err
		}
	}

	return nil
}
