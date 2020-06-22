package devicecfg

import (
	"github.com/ansel1/merry"
	"github.com/fpawel/comm/modbus"
)

type FloatBitsFormat = modbus.FloatBitsFormat
type Params struct {
	Format    FloatBitsFormat `yaml:"format"`
	ParamAddr modbus.Var      `yaml:"reg"`
	Count     modbus.Var      `yaml:"count"`
}

func (p Params) Validate() error {
	if err := p.Format.Validate(); err != nil {
		return merry.Prependf(err, `не правильное знaчение format=%q`, p.Format)
	}
	if p.Count < 1 {
		return merry.Errorf(`не правильное знaчение count=%d: должно быть боьше нуля`, p.Count)
	}
	if p.ParamAddr < 0 {
		return merry.Errorf(`не правильное знaчение reg=%d: должно быть не меньше нуля`, p.ParamAddr)
	}
	if p.ParamAddr+p.Count > 0xFFFF {
		return merry.Errorf(`не правильное знaчение сумы значений reg+count=%d: должно быть не больше 0xFFFF`,
			p.ParamAddr+p.Count)
	}
	return nil
}
