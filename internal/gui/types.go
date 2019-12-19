package gui

import "github.com/fpawel/comm/modbus"

type CommTransaction struct {
	Addr     modbus.Addr
	Comport  string
	Request  string
	Response string
	Ok       bool
}

type ProductParamValue struct {
	Addr      modbus.Addr
	Comport   string
	ParamAddr int
	Value     string
}

type PopupMessage struct {
	Text    string
	Ok      bool
	Warning bool
}

type CoefficientValue struct {
	ProductID   int64
	What        string
	Coefficient int
	Result      string
	Ok          bool
}
