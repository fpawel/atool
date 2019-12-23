package config

import (
	"fmt"
	"github.com/fpawel/comm/modbus"
)

type FloatBitsFormat = modbus.FloatBitsFormat
type Params struct {
	Format    FloatBitsFormat `yaml:"format"`
	ParamAddr int             `yaml:"reg"`
	Count     int             `yaml:"count"`
}

func (p Params) Validate() error {
	if err := p.Format.Validate(); err != nil {
		return fmt.Errorf(`не правильное знaчение format=%q: %w`, p.Format, err)
	}
	if p.Count < 1 {
		return fmt.Errorf(`не правильное знaчение count=%d: должно быть боьше нуля`, p.Count)
	}
	if p.ParamAddr < 0 {
		return fmt.Errorf(`не правильное знaчение reg=%d: должно быть не меньше нуля`, p.ParamAddr)
	}
	if p.ParamAddr+p.Count > 0xFFFF {
		return fmt.Errorf(`не правильное знaчение сумы значений reg+count=%d: должно быть не больше 0xFFFF`,
			p.ParamAddr+p.Count)
	}
	return nil
}
