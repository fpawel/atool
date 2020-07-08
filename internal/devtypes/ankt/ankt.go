package ankt

import (
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/devtypes/ankt/anktvar"
	"github.com/fpawel/atool/internal/devtypes/devdata"
	"github.com/fpawel/comm/modbus"
	"sort"
	"time"
)

var (
	Device = devdata.Device{
		Name:   deviceName,
		Config: deviceConfig,
		ProductTypes: func() (xs []string) {
			for _, t := range productTypesList {
				xs = append(xs, t.String())
			}
			sort.Strings(xs)
			return
		}(),
		PartyParams: []devdata.PartyParam{
			{
				Key:  "c1",
				Name: "Анкат: концентрация ПГС1: начало шк.",
			},
			{
				Key:  "c2",
				Name: "Анкат: концентрация ПГС2: середина шк.к.1",
			},
			{
				Key:  "c3",
				Name: "Анкат: концентрация ПГС3: доп.середина шк.к.1 CO₂",
			},
			{
				Key:  "c4",
				Name: "Анкат: концентрация ПГС4: конец шк.к.1",
			},
			{
				Key:  "c5",
				Name: "Анкат: концентрация ПГС5: середина шк.к.2",
			},
			{
				Key:  "c6",
				Name: "Анкат: концентрация ПГС6: конец шк.к.2",
			},
			{
				Key:  keyTempNorm.String(),
				Name: "уставка температуры НКУ,⁰C",
			},
			{
				Key:  keyTempLow.String(),
				Name: "уставка низкой температуры,⁰C",
			},
			{
				Key:  keyTempHigh.String(),
				Name: "уставка высокой температуры,⁰C",
			},
		},
		InitParty: initParty,
		Work:      main,

		DataSections: dataSections,
		Calc:         nil,
	}

	deviceConfig = devicecfg.Device{
		Baud:               9600,
		TimeoutGetResponse: time.Second,
		TimeoutEndResponse: 50 * time.Millisecond,
		MaxAttemptsRead:    5,
		Pause:              50 * time.Millisecond,
		NetAddr: devicecfg.NetAddr{
			Cmd:    7,
			Format: modbus.BCD,
		},
		CfsList: []devicecfg.Cfs{
			{
				Range:  [2]devicecfg.Kef{0, 50},
				Format: modbus.BCD,
			},
		},
		ParamsNames: anktvar.Names,
		CfsNames:    KfsNames,
		ParamsList:  varsParamRng(anktvar.Vars),
		ProductTypesVars: func() []devicecfg.ProductTypeVars {
			xsC2 := devicecfg.ProductTypeVars{
				ParamsRngList: varsParamRng(anktvar.VarsChan2),
			}
			for _, t := range productTypesList {
				if t.Chan2 {
					xsC2.Names = append(xsC2.Names, t.String())
				}
			}

			xsP := devicecfg.ProductTypeVars{
				ParamsRngList: varsParamRng(anktvar.VarsP),
			}
			for _, t := range productTypesList {
				if t.Press {
					xsP.Names = append(xsP.Names, t.String())
				}
			}
			return []devicecfg.ProductTypeVars{xsC2, xsP}
		}(),
	}
)

const (
	deviceName = "Анкат-7664МИКРО"
)

func varsParamRng(vars []modbus.Var) (xs []devicecfg.Params) {
	for _, v := range vars {
		xs = append(xs, devicecfg.Params{
			Format:    modbus.BCD,
			ParamAddr: v,
			Count:     1,
		})
	}
	return
}
