package gui

import "github.com/fpawel/comm/modbus"

type CommTransaction struct {
	Port     string
	Request  string
	Response string
	Ok       bool
}

type ProductConnection struct {
	ProductID int64
	Ok        bool
}

type ProductParamValue struct {
	Addr      modbus.Addr
	Comport   string
	ParamAddr int
	Value     string
}

type PopupLevel int

const (

	// только popup, не вносить в журнал
	LPopup PopupLevel = iota

	// popup и в журнал
	LJournal

	// в отдельный popup, который нужно специально закрывать, и в журнал
	LWarn
)

type Status struct {
	Text       string
	Ok         bool
	PopupLevel PopupLevel
}

type CoefficientValue struct {
	ProductID   int64
	What        string
	Coefficient int
	Result      string
	Ok          bool
}
