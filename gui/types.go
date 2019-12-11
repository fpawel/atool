package gui

import "github.com/fpawel/comm/modbus"

type CommTransaction struct {
	Addr    modbus.Addr
	Comport string
	Request string
	Result  string
	Ok      bool
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
