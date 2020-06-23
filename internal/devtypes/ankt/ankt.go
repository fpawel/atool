package ankt

import (
	"fmt"
)

type productType struct {
	Index int
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
		{
			Index: 11,
			Chan: [2]chanT{
				{
					gas:   "CO₂",
					scale: 2,
				},
			},
		},
		{
			Index: 12,
			Chan: [2]chanT{
				{
					gas:   "CO₂",
					scale: 5,
				},
			},
		},
		{
			Index: 13,
			Chan: [2]chanT{
				{
					gas:   "CO₂",
					scale: 10,
				},
			},
		},
		{
			Chan2: true,
			Index: 11,
			Chan: [2]chanT{
				{
					gas:   "CO₂",
					scale: 2,
				},
				{
					gas:   "CO₂",
					scale: 100,
				},
			},
		},
		{
			Chan2: true,
			Index: 11,
			Chan: [2]chanT{
				{
					gas:   "CO₂",
					scale: 5,
				},
				{
					gas:   "CO₂",
					scale: 100,
				},
			},
		},
		{
			Chan2: true,
			Index: 11,
			Chan: [2]chanT{
				{
					gas:   "CO₂",
					scale: 10,
				},
				{
					gas:   "CO₂",
					scale: 100,
				},
			},
		},
		{
			Index: 15,
			Chan: [2]chanT{
				{
					gas:   "∑CH",
					scale: 100,
				},
			},
		},
		{
			Index: 14,
			Chan: [2]chanT{
				{
					gas:   "C₃H₈",
					scale: 100,
				},
			},
		},
		{
			Index: 16,
			Chan: [2]chanT{
				{
					gas:   "CH₄",
					scale: 100,
				},
			},
		},
		{
			Index: 11,
			Chan: [2]chanT{
				{
					gas:   "CO₂",
					scale: 2,
				},
			},
		},
		{
			Index: 12,
			Chan: [2]chanT{
				{
					gas:   "CO₂",
					scale: 5,
				},
			},
		},
		{
			Index: 13,
			Chan: [2]chanT{
				{
					gas:   "CO₂",
					scale: 10,
				},
			},
		},
		{
			Index: 16,
			Chan: [2]chanT{
				{
					gas:   "CH₄",
					scale: 100,
				},
			},
		},
		{
			Index: 14,
			Chan: [2]chanT{
				{
					gas:   "C₃H₈",
					scale: 100,
				},
			},
		},
		{
			Index: 15,
			Chan: [2]chanT{
				{
					gas:   "∑CH",
					scale: 100,
				},
			},
		},
		{
			Index: 11,
			Chan: [2]chanT{
				{
					gas:   "CO₂",
					scale: 2,
				},
			},
		},
		{
			Index: 12,
			Chan: [2]chanT{
				{
					gas:   "CO₂",
					scale: 5,
				},
			},
		},
		{
			Index: 13,
			Chan: [2]chanT{
				{
					gas:   "CO₂",
					scale: 10,
				},
			},
		},
		{
			Index: 16,
			Chan: [2]chanT{
				{
					gas:   "CH₄",
					scale: 100,
				},
			},
		},
		{
			Index: 14,
			Chan: [2]chanT{
				{
					gas:   "C₃H₈",
					scale: 100,
				},
			},
		},
		{
			Index: 15,
			Chan: [2]chanT{
				{
					gas:   "∑CH",
					scale: 100,
				},
			},
		},
		{
			Index: 16,
			Chan: [2]chanT{
				{
					gas:   "CH₄",
					scale: 100,
				},
			},
		},
		{
			Index: 14,
			Chan: [2]chanT{
				{
					gas:   "C₃H₈",
					scale: 100,
				},
			},
		},
		{
			Index: 15,
			Chan: [2]chanT{
				{
					gas:   "∑CH",
					scale: 100,
				},
			},
		},
		{
			Index: 16,
			Chan: [2]chanT{
				{
					gas:   "CH₄",
					scale: 100,
				},
			},
		},
		{
			Index: 14,
			Chan: [2]chanT{
				{
					gas:   "C₃H₈",
					scale: 100,
				},
			},
		},
		{
			Index: 15,
			Chan: [2]chanT{
				{
					gas:   "∑CH",
					scale: 100,
				},
			},
		},
		{
			Index: 16,
			Chan: [2]chanT{
				{
					gas:   "CH₄",
					scale: 100,
				},
			},
		},
		{
			Index: 14,
			Chan: [2]chanT{
				{
					gas:   "C₃H₈",
					scale: 100,
				},
			},
		},
		{
			Index: 15,
			Chan: [2]chanT{
				{
					gas:   "∑CH",
					scale: 100,
				},
			},
		},
		{
			Index: 16,
			Chan: [2]chanT{
				{
					gas:   "CH₄",
					scale: 100,
				},
			},
		},
		{
			Index: 14,
			Chan: [2]chanT{
				{
					gas:   "C₃H₈",
					scale: 100,
				},
			},
		},
		{
			Index: 15,
			Chan: [2]chanT{
				{
					gas:   "∑CH",
					scale: 100,
				},
			},
		},
		{
			Index: 16,
			Chan: [2]chanT{
				{
					gas:   "CH₄",
					scale: 100,
				},
			},
		},
		{
			Index: 14,
			Chan: [2]chanT{
				{
					gas:   "C₃H₈",
					scale: 100,
				},
			},
		},
		{
			Index: 15,
			Chan: [2]chanT{
				{
					gas:   "∑CH",
					scale: 100,
				},
			},
		},
		{
			Index: 16,
			Chan: [2]chanT{
				{
					gas:   "CH₄",
					scale: 100,
				},
			},
		},
		{
			Index: 14,
			Chan: [2]chanT{
				{
					gas:   "C₃H₈",
					scale: 100,
				},
			},
		},
		{
			Index: 15,
			Chan: [2]chanT{
				{
					gas:   "∑CH",
					scale: 100,
				},
			},
		},
		{
			Index: 16,
			Chan: [2]chanT{
				{
					gas:   "CH₄",
					scale: 100,
				},
			},
		},
		{
			Index: 14,
			Chan: [2]chanT{
				{
					gas:   "C₃H₈",
					scale: 100,
				},
			},
		},
		{
			Index: 15,
			Chan: [2]chanT{
				{
					gas:   "∑CH",
					scale: 100,
				},
			},
		},
		{
			Chan2: true,
			Index: 33,
			Chan: [2]chanT{
				{
					gas:   "CO₂",
					scale: 2,
				},
				{
					gas:   "CO₂",
					scale: 100,
				},
			},
		},
		{
			Chan2: true,
			Index: 33,
			Chan: [2]chanT{
				{
					gas:   "CO₂",
					scale: 5,
				},
				{
					gas:   "CO₂",
					scale: 100,
				},
			},
		},
		{
			Chan2: true,
			Index: 33,
			Chan: [2]chanT{
				{
					gas:   "CO₂",
					scale: 10,
				},
				{
					gas:   "CO₂",
					scale: 100,
				},
			},
		},
		{
			Chan2: true,
			Index: 33,
			Chan: [2]chanT{
				{
					gas:   "CH₄",
					scale: 100,
				},
				{
					gas:   "CH₄",
					scale: 100,
				},
			},
		},
		{
			Chan2: true,
			Index: 33,
			Chan: [2]chanT{
				{
					gas:   "C₃H₈",
					scale: 100,
				},
				{
					gas:   "C₃H₈",
					scale: 100,
				},
			},
		},
		{
			Chan2: true,
			Index: 33,
			Chan: [2]chanT{
				{
					gas:   "∑CH",
					scale: 100,
				},
				{
					gas:   "∑CH",
					scale: 100,
				},
			},
		},
		{
			Chan2: true,
			Index: 34,
			Chan: [2]chanT{
				{
					gas:   "CO₂",
					scale: 2,
				},
				{
					gas:   "CO₂",
					scale: 100,
				},
			},
		},
		{
			Chan2: true,
			Index: 34,
			Chan: [2]chanT{
				{
					gas:   "CO₂",
					scale: 5,
				},
				{
					gas:   "CO₂",
					scale: 100,
				},
			},
		},
		{
			Chan2: true,
			Index: 34,
			Chan: [2]chanT{
				{
					gas:   "CO₂",
					scale: 10,
				},
				{
					gas:   "CO₂",
					scale: 100,
				},
			},
		},
		{
			Chan2: true,
			Index: 34,
			Chan: [2]chanT{
				{
					gas:   "CH₄",
					scale: 100,
				},
				{
					gas:   "CH₄",
					scale: 100,
				},
			},
		},
		{
			Chan2: true,
			Index: 34,
			Chan: [2]chanT{
				{
					gas:   "C₃H₈",
					scale: 100,
				},
				{
					gas:   "C₃H₈",
					scale: 100,
				},
			},
		},
		{
			Chan2: true,
			Index: 34,
			Chan: [2]chanT{
				{
					gas:   "∑CH",
					scale: 100,
				},
				{
					gas:   "∑CH",
					scale: 100,
				},
			},
		},
		{
			Chan2: true,
			Index: 35,
			Chan: [2]chanT{
				{
					gas:   "CO₂",
					scale: 2,
				},
				{
					gas:   "CO₂",
					scale: 100,
				},
			},
		},
		{
			Chan2: true,
			Index: 35,
			Chan: [2]chanT{
				{
					gas:   "CO₂",
					scale: 5,
				},
				{
					gas:   "CO₂",
					scale: 100,
				},
			},
		},
		{
			Chan2: true,
			Index: 35,
			Chan: [2]chanT{
				{
					gas:   "CO₂",
					scale: 10,
				},
				{
					gas:   "CO₂",
					scale: 100,
				},
			},
		},
		{
			Chan2: true,
			Index: 35,
			Chan: [2]chanT{
				{
					gas:   "CH₄",
					scale: 100,
				},
				{
					gas:   "CH₄",
					scale: 100,
				},
			},
		},
		{
			Chan2: true,
			Index: 35,
			Chan: [2]chanT{
				{
					gas:   "C₃H₈",
					scale: 100,
				},
				{
					gas:   "C₃H₈",
					scale: 100,
				},
			},
		},
		{
			Chan2: true,
			Index: 35,
			Chan: [2]chanT{
				{
					gas:   "∑CH",
					scale: 100,
				},
				{
					gas:   "∑CH",
					scale: 100,
				},
			},
		},
		{
			Index: 16,
			Chan: [2]chanT{
				{
					gas:   "CH₄",
					scale: 100,
				},
			},
		},
		{
			Index: 14,
			Chan: [2]chanT{
				{
					gas:   "C₃H₈",
					scale: 100,
				},
			},
		},
		{
			Index: 15,
			Chan: [2]chanT{
				{
					gas:   "∑CH",
					scale: 100,
				},
			},
		},
		{
			Index: 16,
			Chan: [2]chanT{
				{
					gas:   "CH₄",
					scale: 100,
				},
			},
		},
		{
			Index: 14,
			Chan: [2]chanT{
				{
					gas:   "C₃H₈",
					scale: 100,
				},
			},
		},
		{
			Index: 15,
			Chan: [2]chanT{
				{
					gas:   "∑CH",
					scale: 100,
				},
			},
		},
		{
			Index: 16,
			Chan: [2]chanT{
				{
					gas:   "CH₄",
					scale: 100,
				},
			},
		},
		{
			Index: 14,
			Chan: [2]chanT{
				{
					gas:   "C₃H₈",
					scale: 100,
				},
			},
		},
		{
			Index: 15,
			Chan: [2]chanT{
				{
					gas:   "∑CH",
					scale: 100,
				},
			},
		},
	}
)
