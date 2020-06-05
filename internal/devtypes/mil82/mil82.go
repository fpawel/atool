package mil82

import (
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/devtypes/devdata"
	"github.com/fpawel/comm/modbus"
	"time"
)

var Device = devdata.Device{

	GetCalcSectionsFunc: calcSections,

	ProductTypes: prodTypeNames,

	DataSections: DataSections(),

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
		PartyParams: devicecfg.PartyParams{
			"c1":            "ПГС1",
			"c2":            "ПГС2",
			"c3":            "ПГС3",
			"c4":            "ПГС4",
			"linear_degree": "Степень линеаризации",
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
}
