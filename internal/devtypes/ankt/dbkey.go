package ankt

import (
	"fmt"
	"github.com/fpawel/atool/internal/devtypes/ankt/anktvar"
	"github.com/fpawel/comm/modbus"
)

func dbKeyTemp(Chan cChan, keyTemp keyTemp, gas gas) string {
	Chan.mustCheck()
	keyTemp.mustCheck()
	gas.mustCheck()
	return fmt.Sprintf("%s_gas%d_chan%d", keyTemp, gas, Chan)
}

type cChan int

func (x cChan) dbKeyLin(gas gas) string {
	x.mustCheck()
	gas.mustCheck()
	return fmt.Sprintf("lin%d_chan%d", gas, x)
}

func (x cChan) Cout() modbus.Var {
	x.mustCheck()
	switch x {
	case chan1:
		return anktvar.CoutCh0
	case chan2:
		return anktvar.CoutCh1
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
		panic(fmt.Sprintf("invalid gas: %d", x))
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
