package ankt

import (
	"fmt"
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/devtypes/ankt/anktvar"
	"github.com/fpawel/atool/internal/devtypes/devdata"
	"github.com/fpawel/comm/modbus"
	"sort"
	"time"
)

var (
	Device = devdata.Device{
		Name:   "Анкат-7664МИКРО",
		Config: deviceConfig,
		ProductTypes: func() (xs []string) {
			for _, t := range productTypesList {
				xs = append(xs, t.String())
			}
			sort.Strings(xs)
			return
		}(),

		DataSections: nil,
		PartyParams:  nil,
		InitParty:    nil,
		Calc:         nil,
		Work:         nil,
	}
)

type chanT struct {
	gas   gasT
	scale float64
}

func (x chanT) String() string {
	if x.gas.isCH() {
		return string(x.gas)
	}
	return fmt.Sprintf("%s%v", x.gas, x.scale)
}

func (x chanT) code() float64 {
	switch x.gas {
	case CH4:
		return 16
	case SumCH:
		return 15
	case C3H8:
		return 14
	case CO2:
		switch x.scale {
		case 2:
			return 11
		case 5:
			return 12
		case 10:
			return 13
		}
	}
	panic(fmt.Sprintf("%+v", x))
}

type gasT string

func (x gasT) units() unitsT {
	if x.isCH() {
		return unitsLel
	}
	return unitsLel
}

func (x gasT) isCH() bool {
	switch x {
	case CH4, C3H8, SumCH:
		return true
	}
	return false
}

type unitsT string

func (x unitsT) code() float64 {
	switch x {
	case unitsVolume:
		return 3
	case unitsLel:
		return 4
	}
	panic(x)
}

func scaleCode(x float64) float64 {
	switch x {
	case 2:
		return 2
	case 5:
		return 6
	case 10:
		return 7
	case 100:
		return 21
	}
	panic(x)
}

const (
	unitsVolume unitsT = "%об"
	unitsLel    unitsT = "%НКПР"
	CO2         gasT   = "CO₂"
	C3H8        gasT   = "C₃H₈"
	SumCH       gasT   = "∑CH"
	CH4         gasT   = "CH₄"
)

var (
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
				if t.Pressure {
					xsP.Names = append(xsP.Names, t.String())
				}
			}
			return []devicecfg.ProductTypeVars{xsC2, xsP}
		}(),
	}
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
