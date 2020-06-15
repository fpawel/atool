package mil82

import (
	"fmt"
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/devtypes/devdata"
	"github.com/fpawel/comm/modbus"
	"time"
)

var Device = devdata.Device{

	Name: "МИЛ-82",

	Work: work,

	Calc: calcSections,

	ProductTypes: prodTypeNames,

	DataSections: DataSections(),

	InitParty: initParty,

	Config: devicecfg.Device{
		Baud:               9600,
		TimeoutGetResponse: time.Second,
		TimeoutEndResponse: 50 * time.Millisecond,
		MaxAttemptsRead:    5,
		Pause:              50 * time.Millisecond,
		NetAddr: devicecfg.NetAddr{
			Cmd:    12,
			Format: modbus.BCD,
		},
		Params: []devicecfg.Params{
			{
				Format:    modbus.BCD,
				ParamAddr: 0,
				Count:     2,
			},
			{
				Format:    modbus.BCD,
				ParamAddr: 4,
				Count:     1,
			},
			{
				Format:    modbus.BCD,
				ParamAddr: 12,
				Count:     2,
			},
			{
				Format:    modbus.BCD,
				ParamAddr: 16,
				Count:     1,
			},
		},
		Coefficients: []devicecfg.Coefficients{
			{
				Range:  [2]int{0, 50},
				Format: modbus.BCD,
			},
		},
		ParamsNames: map[int]string{
			0:  "C",
			2:  "I",
			4:  "Is",
			12: "Work",
			14: "Ref",
			16: "Var16",
		},
	},

	PartyParams: []devdata.PartyParam{
		{
			Key:  "c1",
			Name: "ПГС1",
		},
		{
			Key:  "c2",
			Name: "ПГС2",
		},
		{
			Key:  "c3",
			Name: "ПГС3",
		},
		{
			Key:  "c4",
			Name: "ПГС5",
		},
		{
			Key:        keyLinearDegree,
			Name:       "степень линеаризации",
			ValuesList: []string{"3", "4"},
		},
		{
			Key:  keyTempNorm,
			Name: "уставка температуры НКУ,⁰C",
		},
		{
			Key:  keyTempLow,
			Name: "уставка низкой температуры,⁰C",
		},
		{
			Key:  keyTempHigh,
			Name: "уставка высокой температуры,⁰C",
		},
	},
}

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
		if _, err := data.DB.Exec(`UPDATE party SET product_type = ? WHERE party_id = (SELECT party_id FROM app_config)`, Type.Name); err != nil {
			return err
		}
	}

	if err := data.SetCurrentPartyValue(keyTempNorm, 20); err != nil {
		return err
	}
	if err := data.SetCurrentPartyValue(keyTempLow, Type.TempMin); err != nil {
		return err
	}
	if err := data.SetCurrentPartyValue(keyTempHigh, Type.TempMax); err != nil {
		return err
	}
	return nil
}

const (
	keyLinearDegree = "linear_degree"
	keyTempNorm     = "temp_norm"
	keyTempLow      = "temp_low"
	keyTempHigh     = "temp_high"
)
