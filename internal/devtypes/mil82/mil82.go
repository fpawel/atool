package mil82

import (
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/devtypes/devdata"
	"github.com/fpawel/comm/modbus"
	"time"
)

var Device = devdata.Device{
	Name:         "МИЛ-82",
	Work:         main,
	Calc:         calcSections,
	ProductTypes: prodTypeNames,
	DataSections: dataSections(),
	InitParty:    initParty,
	Config:       deviceConfig,
	PartyParams:  partyParams,
}

const (
	keyLinearDegree = "linear_degree"
	keyTempNorm     = "t_norm"
	keyTempLow      = "t_low"
	keyTempHigh     = "t_high"

	keyTestTempNorm = "test_t_norm"
	keyTestTempLow  = "test_t_low"
	keyTestTempHigh = "test_t_high"
	keyTestTemp80   = "test_t80"

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

var (
	deviceConfig = devicecfg.Device{
		Baud:               9600,
		TimeoutGetResponse: time.Second,
		TimeoutEndResponse: 50 * time.Millisecond,
		MaxAttemptsRead:    5,
		Pause:              50 * time.Millisecond,
		NetAddr:            12,
		FloatFormat:        modbus.BCD,
		Vars: []devicecfg.Vars{
			{0, 2},
			{4, 1},
			{12, 2},
			{16, 1},
		},
		CfsList: []devicecfg.Cfs{
			{0, 50},
		},
		VarsNames: paramsNames,
		CfsNames:  KfsNames,
		Commands: map[modbus.DevCmd]string{
			1: "Корректировка нуля",
			2: "Корректировка чувствительности",
			8: "Нормировка",
		},
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
		{
			Key:  keyTestTemp80,
			Name: "уставка высокой температуры 80⁰C",
		},
	}

	paramsNames = map[modbus.Var]string{
		varConcentration: "C",
		varTemp:          "T",
		4:                "Is",
		12:               "Work",
		14:               "Ref",
		var16:            "Var16",
	}
)

type KefValueMap = map[devicecfg.Kef]float64
