package gui

import "github.com/fpawel/comm/modbus"

type CommTransaction struct {
	Addr     modbus.Addr
	Comport  string
	Request  string
	Response string
	Duration string
	Ok       bool
}

type ProductParamValue struct {
	Addr      modbus.Addr
	Comport   string
	ParamAddr int
	Value     string
}

type PopupMessage struct {
	Text string
	Ok   bool
}
