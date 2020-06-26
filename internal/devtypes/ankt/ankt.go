package ankt

import (
	"fmt"
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/comm/modbus"
	"time"
)

var (
	Device1 = productTypesList.filter(func(p productType) bool {
		return !p.Chan2 && !p.Pressure
	}).device()

	Device1P = productTypesList.filter(func(p productType) bool {
		return !p.Chan2 && p.Pressure
	}).device()

	Device2 = productTypesList.filter(func(p productType) bool {
		return p.Chan2 && !p.Pressure
	}).device()

	Device2P = productTypesList.filter(func(p productType) bool {
		return p.Chan2 && p.Pressure
	}).device()
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
	deviceConfig0 = devicecfg.Device{
		Baud:               9600,
		TimeoutGetResponse: time.Second,
		TimeoutEndResponse: 50 * time.Millisecond,
		MaxAttemptsRead:    5,
		Pause:              50 * time.Millisecond,
		NetAddr: devicecfg.NetAddr{
			Cmd:    7,
			Format: modbus.BCD,
		},
		CfsRngList: []devicecfg.CfsRng{
			{
				Range:  [2]devicecfg.Kef{0, 50},
				Format: modbus.BCD,
			},
		},
		ParamsNames: paramsNames,
		CfsNames:    KfsNames,
	}
)

func deviceConfig(vars []modbus.Var) devicecfg.Device {
	x := deviceConfig0
	for _, v := range vars {
		x.ParamsRng = append(x.ParamsRng, devicecfg.ParamsRng{
			Format:    modbus.BCD,
			ParamAddr: v,
			Count:     1,
		})
	}
	return x
}
