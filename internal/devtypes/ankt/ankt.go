package ankt

import (
	"fmt"
)

type productType struct {
	N     int
	Chan  [2]chanT
	Chan2 bool
}

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

var (
	unitsVolume unitsT = "%об"
	unitsLel    unitsT = "%НКПР"

	CO2   gasT = "CO₂"
	C3H8  gasT = "C₃H₈"
	SumCH gasT = "∑CH"
	CH4   gasT = "CH₄"

	productTypes = []productType{
		prodT1(11, "CO₂", 2),
		prodT1(12, "CO₂", 5),
		prodT1(13, "CO₂", 10),
		prodT2(11, "CO₂", 2, "CH₄", 100),
		prodT2(12, "CO₂", 5, "CH₄", 100),
		prodT2(13, "CO₂", 10, "CH₄", 100),
		prodT1(15, "∑CH", 100),
		prodT1(14, "C₃H₈", 100),
		prodT1(16, "CH₄", 100),
		prodT1(11, "CO₂", 2),
		prodT1(12, "CO₂", 5),
		prodT1(13, "CO₂", 10),
		prodT1(16, "CH₄", 100),
		prodT1(14, "C₃H₈", 100),
		prodT1(15, "∑CH", 100),
		prodT1(11, "CO₂", 2),
		prodT1(12, "CO₂", 5),
		prodT1(13, "CO₂", 10),
		prodT1(16, "CH₄", 100),
		prodT1(14, "C₃H₈", 100),
		prodT1(15, "∑CH", 100),
		prodT1(16, "CH₄", 100),
		prodT1(14, "C₃H₈", 100),
		prodT1(15, "∑CH", 100),
		prodT1(16, "CH₄", 100),
		prodT1(14, "C₃H₈", 100),
		prodT1(15, "∑CH", 100),
		prodT1(16, "CH₄", 100),
		prodT1(14, "C₃H₈", 100),
		prodT1(15, "∑CH", 100),
		prodT1(16, "CH₄", 100),
		prodT1(14, "C₃H₈", 100),
		prodT1(15, "∑CH", 100),
		prodT1(16, "CH₄", 100),
		prodT1(14, "C₃H₈", 100),
		prodT1(15, "∑CH", 100),
		prodT1(16, "CH₄", 100),
		prodT1(14, "C₃H₈", 100),
		prodT1(15, "∑CH", 100),
		prodT1(16, "CH₄", 100),
		prodT1(14, "C₃H₈", 100),
		prodT1(15, "∑CH", 100),
		prodT2(11, "CO₂", 2, "CH₄", 100),
		prodT2(12, "CO₂", 5, "CH₄", 100),
		prodT2(13, "CO₂", 10, "CH₄", 100),
		prodT2(16, "CH₄", 100, "CH₄", 100),
		prodT2(14, "C₃H₈", 100, "CH₄", 100),
		prodT2(15, "∑CH", 100, "CH₄", 100),
		prodT2(11, "CO₂", 2, "CH₄", 100),
		prodT2(12, "CO₂", 5, "CH₄", 100),
		prodT2(13, "CO₂", 10, "CH₄", 100),
		prodT2(16, "CH₄", 100, "CH₄", 100),
		prodT2(14, "C₃H₈", 100, "CH₄", 100),
		prodT2(15, "∑CH", 100, "CH₄", 100),
		prodT2(11, "CO₂", 2, "CH₄", 100),
		prodT2(12, "CO₂", 5, "CH₄", 100),
		prodT2(13, "CO₂", 10, "CH₄", 100),
		prodT2(16, "CH₄", 100, "CH₄", 100),
		prodT2(14, "C₃H₈", 100, "CH₄", 100),
		prodT2(15, "∑CH", 100, "CH₄", 100),
		prodT1(16, "CH₄", 100),
		prodT1(14, "C₃H₈", 100),
		prodT1(15, "∑CH", 100),
		prodT1(16, "CH₄", 100),
		prodT1(14, "C₃H₈", 100),
		prodT1(15, "∑CH", 100),
		prodT1(16, "CH₄", 100),
		prodT1(14, "C₃H₈", 100),
		prodT1(15, "∑CH", 100),
	}
)

func prodT1(n int, gas gasT, scale float64) productType {
	return productType{
		N: n,
		Chan: [2]chanT{
			{
				gas:   gas,
				scale: scale,
			},
		},
	}
}

func prodT2(n int, gas gasT, scale float64, gas2 gasT, scale2 float64) productType {
	return productType{
		Chan2: true,
		N:     n,
		Chan: [2]chanT{
			{
				gas:   gas,
				scale: scale,
			},
			{
				gas:   gas2,
				scale: scale2,
			},
		},
	}
}
