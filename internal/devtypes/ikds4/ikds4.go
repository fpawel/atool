package ikds4

import (
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/devtypes/devdata"
	"github.com/fpawel/atool/internal/devtypes/mil82"
	"github.com/fpawel/comm/modbus"
	"time"
)

var (
	Device = devdata.Device{
		Work:         main,
		Name:         "ИКД-С4",
		Calc:         calcSections,
		ProductTypes: prodTypeNames,
		DataSections: dataSections(),
		Config:       deviceConfig,
		PartyParams:  partyParams,
		InitParty:    initParty,
	}

	partyParams = []devdata.PartyParam{
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
	}

	deviceConfig = devicecfg.Device{
		Baud:               9600,
		TimeoutGetResponse: time.Second,
		TimeoutEndResponse: 50 * time.Millisecond,
		MaxAttemptsRead:    5,
		Pause:              50 * time.Millisecond,
		NetAddr: devicecfg.NetAddr{
			Cmd:    12,
			Format: modbus.FloatBigEndian,
		},
		ParamsRng: []devicecfg.ParamsRng{
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
		CfsRngList: []devicecfg.CfsRng{
			{
				Range:  [2]devicecfg.Kef{0, 50},
				Format: modbus.FloatBigEndian,
			},
		},
		ParamsNames: map[modbus.Var]string{
			0:   "C",
			2:   "I",
			4:   "Is",
			12:  "Work",
			14:  "Ref",
			16:  "Var16",
			200: "Var200",
		},
		CfsNames: mil82.KfsNames,
	}
)

const (
	keyLinearDegree = "linear_degree"
	keyTempNorm     = "t_norm"
	keyTempLow      = "t_low"
	keyTempHigh     = "t_high"

	keyTestTempNorm = "test_t_norm"
	keyTestTempLow  = "test_t_low"
	keyTestTempHigh = "test_t_high"

	keyTest2 = "test2"
	keyTex1  = "tex1"
	keyTex2  = "tex2"
)

const (
	varConcentration modbus.Var = 0
	varTemp          modbus.Var = 2
	var16                       = 16
)

var (
	vars    = []modbus.Var{varConcentration, varTemp, 4, 8, 10, 12, 14, var16}
	ptsTemp = []string{keyTempLow, keyTempNorm, keyTempHigh}
)

type KefValueMap = map[devicecfg.Kef]float64
