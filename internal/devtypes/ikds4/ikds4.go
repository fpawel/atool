package ikds4

import (
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/devtypes/devdata"
	"github.com/fpawel/atool/internal/devtypes/mil82"
	"github.com/fpawel/comm/modbus"
	"time"
)

var Device = devdata.Device{

	Name: "ИКД-С4",

	Calc: calcSections,

	ProductTypes: prodTypeNames,

	DataSections: mil82.DataSections(),

	Config: devicecfg.Device{
		Baud:               9600,
		TimeoutGetResponse: time.Second,
		TimeoutEndResponse: 50 * time.Millisecond,
		MaxAttemptsRead:    5,
		Pause:              50 * time.Millisecond,
		NetAddr: devicecfg.NetAddr{
			Cmd:    12,
			Format: modbus.FloatBigEndian,
		},
		Params: []devicecfg.Params{
			{
				Format:    modbus.FloatBigEndian,
				ParamAddr: 0,
				Count:     2,
			},
			{
				Format:    modbus.FloatBigEndian,
				ParamAddr: 4,
				Count:     1,
			},
			{
				Format:    modbus.FloatBigEndian,
				ParamAddr: 12,
				Count:     2,
			},
			{
				Format:    modbus.FloatBigEndian,
				ParamAddr: 16,
				Count:     1,
			},
			{
				Format:    modbus.FloatBigEndian,
				ParamAddr: 200,
				Count:     1,
			},
		},
		Coefficients: []devicecfg.Coefficients{
			{
				Range:  [2]int{0, 50},
				Format: modbus.FloatBigEndian,
			},
		},
		ParamsNames: map[int]string{
			0:   "C",
			2:   "I",
			4:   "Is",
			12:  "Work",
			14:  "Ref",
			16:  "Var16",
			200: "Var200",
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
	},
}
