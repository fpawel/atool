package cfg

import (
	"fmt"
	"strings"
)

const (
	ParamFormatBCD               ParamFormat = "bcd"
	ParamFormatFloatBigEndian    ParamFormat = "float_big_endian"
	ParamFormatFloatLittleEndian ParamFormat = "float_little_endian"
	ParamFormatIntBigEndian      ParamFormat = "int_big_endian"
	ParamFormatIntLittleEndian   ParamFormat = "int_little_endian"
)

type Params struct {
	Format    ParamFormat `yaml:"format"`
	ParamAddr int         `yaml:"reg"`
	Count     int         `yaml:"count"`
}

type ParamFormat string

func (p Params) Validate() error {
	if _, f := ParamFormats[p.Format]; !f {
		return fmt.Errorf(`не правильное знaчение format=%q: должно быть из списка %s`, p.Format, formatParamFormats())
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

func formatParamFormats() string {
	var xs []string
	for s := range ParamFormats {
		xs = append(xs, string(s))
	}
	return strings.Join(xs, ",")
}

var ParamFormats = map[ParamFormat]struct{}{
	ParamFormatBCD:               {},
	ParamFormatFloatBigEndian:    {},
	ParamFormatFloatLittleEndian: {},
	ParamFormatIntBigEndian:      {},
	ParamFormatIntLittleEndian:   {},
}
