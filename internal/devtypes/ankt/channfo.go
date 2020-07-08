package ankt

import "fmt"

type chanNfo struct {
	gasName gasName
	scale   float64
}

func (x chanNfo) String() string {
	if x.gasName.isCH() {
		return string(x.gasName)
	}
	return fmt.Sprintf("%s%v", x.gasName, x.scale)
}

func (x chanNfo) code() float64 {
	switch x.gasName {
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

type gasName string

func (x gasName) units() unitsName {
	if x.isCH() {
		return unitsLel
	}
	return unitsLel
}

func (x gasName) isCO2() bool {
	return !x.isCH()
}

func (x gasName) isCH() bool {
	switch x {
	case CH4, C3H8, SumCH:
		return true
	}
	return false
}

type unitsName string

func (x unitsName) code() float64 {
	switch x {
	case unitsVolume:
		return 3
	case unitsLel:
		return 4
	}
	panic(x)
}

const (
	unitsVolume unitsName = "%об"
	unitsLel    unitsName = "%НКПР"
	CO2         gasName   = "CO₂"
	C3H8        gasName   = "C₃H₈"
	SumCH       gasName   = "∑CH"
	CH4         gasName   = "CH₄"
)
