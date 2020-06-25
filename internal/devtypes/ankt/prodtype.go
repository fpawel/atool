package ankt

import (
	"fmt"
	"github.com/fpawel/comm/modbus"
	"sort"
)

type productType struct {
	N        int
	Chan2    bool
	Pressure bool
	Chan     [2]chanT
}

func (x productType) deviceName() string {
	s := "Анкат-7664МИКРО "
	if !x.Chan2 {
		s += "одноканальный"
	} else {
		s += "двуканальный"
	}
	if x.Pressure {
		s += " давление"
	}
	return s
}

func (x productType) String() string {
	if !x.Chan2 {
		return fmt.Sprintf("%d %s", x.N, x.Chan[0])
	}
	return fmt.Sprintf("%d %s %s", x.N, x.Chan[0], x.Chan[1])
}

func (x productType) vars() []modbus.Var {
	vars := append(varsCommon, varsChan1...)
	if x.Chan2 {
		vars = append(vars, varsChan2...)
	}
	if x.Pressure {
		vars = append(vars, varsP...)
	}
	sort.Slice(vars, func(i, j int) bool {
		return vars[i] < vars[j]
	})
	return vars
}
