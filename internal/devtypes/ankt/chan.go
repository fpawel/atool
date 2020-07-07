package ankt

import "fmt"

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
	return 0
}

type gasT string

func (x gasT) units() unitsT {
	if x.isCH() {
		return unitsLel
	}
	return unitsLel
}

func (x gasT) isCO2() bool {
	return !x.isCH()
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

const (
	unitsVolume unitsT = "%об"
	unitsLel    unitsT = "%НКПР"
	CO2         gasT   = "CO₂"
	C3H8        gasT   = "C₃H₈"
	SumCH       gasT   = "∑CH"
	CH4         gasT   = "CH₄"
)
