package gui

import (
	"github.com/fpawel/comm/modbus"
)

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

type ProgressInfo struct {
	Cmd      ProgressCmd
	Progress int
	Max      int
	What     string
}

type ProgressCmd int

const (
	ProgressHide ProgressCmd = iota
	ProgressShow
	ProgressProgress
)

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

func (x Status) WithError(err error) Status {
	if err == nil {
		x.Ok = true
		return x
	}
	if len(x.Text) > 0 {
		x.Text += ": " + err.Error()
	} else {
		x.Text = err.Error()
	}
	x.Ok = false
	return x
}

type CoefficientValue struct {
	ProductID   int64
	Read        bool
	Coefficient int
	Result      string
	Ok          bool
}

type LuaConfigParam struct {
	Name    string
	TheType string
	List    []string
	Value   string
}
