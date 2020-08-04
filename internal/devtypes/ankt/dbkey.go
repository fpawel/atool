package ankt

import (
	"fmt"
	"github.com/fpawel/atool/internal/devtypes/ankt/anktvar"
	"github.com/fpawel/atool/internal/devtypes/devdata"
	"github.com/fpawel/comm/modbus"
)

type cChan int

func (x cChan) keyLin(gas gas) string {
	x.mustCheck()
	gas.mustCheck()
	return fmt.Sprintf("lin%d_chan%d", gas, x)
}

type cChanNfo struct {
	Cout, Var2, Tpp modbus.Var
	KefT0, KefTK    kef
}

func (x cChan) dataParamLin(gas gas) devdata.DataParam {
	return devdata.DataParam{
		Key:  x.keyLin(gas),
		Name: fmt.Sprintf("канал %d газ %d", x, gas),
	}
}

func (x cChan) Nfo() cChanNfo {
	x.mustCheck()
	switch x {
	case chan1:
		return cChanNfo{
			Cout:  anktvar.CoutCh0,
			Var2:  anktvar.Var2Ch0,
			Tpp:   anktvar.TppCh0,
			KefT0: kefCh1T0v0,
			KefTK: kefCh1TKv0,
		}
	case chan2:
		return cChanNfo{
			Cout:  anktvar.CoutCh1,
			Var2:  anktvar.Var2Ch1,
			Tpp:   anktvar.TppCh1,
			KefT0: kefCh2T0v0,
			KefTK: kefCh2TKv0,
		}
	default:
		panic("unexpected")
	}
}

func (x cChan) mustCheck() {
	switch x {
	case chan1, chan2:
		return
	default:
		panic(fmt.Sprintf("invalid measure chan: %d", x))
	}
}

type keyTemp string

func (x keyTemp) String() string {
	return string(x)
}

func (x keyTemp) What() string {
	switch x {
	case keyTempNorm:
		return "нормальная температура"
	case keyTempLow:
		return "низкая температура"
	case keyTempHigh:
		return "высокая температура"
	default:
		return string(x)
	}
}

func keyGasVar(x keyTemp, gas gas, Var modbus.Var) string {
	x.mustCheck()
	gas.mustCheck()
	return fmt.Sprintf("%s_gas%d_var%d", x, gas, Var)
}

func keyPT(x keyTemp) string {
	return fmt.Sprintf("pt_%s", x)
}

func (x keyTemp) mustCheck() {
	switch x {
	case keyTempNorm, keyTempLow, keyTempHigh:
		return
	default:
		panic("invalid temp key: " + x)
	}
}

type gas byte

func (x gas) mustCheck() {
	switch x {
	case gas1, gas2, gas3, gas4, gas5, gas6:
		return
	default:
		panic(fmt.Sprintf("invalid gasName: %d", x))
	}
}

const (
	chan1 cChan = 1
	chan2 cChan = 2

	gas1 gas = 1
	gas2 gas = 2
	gas3 gas = 3
	gas4 gas = 4
	gas5 gas = 5
	gas6 gas = 6

	keyTempNorm keyTemp = "t_norm"
	keyTempLow  keyTemp = "t_low"
	keyTempHigh keyTemp = "t_high"
)

func mapTemps(f func(temp keyTemp) string) (xs []string) {
	for _, k := range keysTemp {
		xs = append(xs, f(k))
	}
	return
}

var (
	keysTemp = []keyTemp{keyTempLow, keyTempNorm, keyTempHigh}
	chan2nfo = chan2.Nfo()
	chan1nfo = chan1.Nfo()
)
