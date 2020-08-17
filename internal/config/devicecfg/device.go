package devicecfg

import (
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"sort"
	"time"
)

type Device struct {
	FloatFormat        FloatBitsFormat                `yaml:"float_format"`
	Baud               int                            `yaml:"baud"`
	TimeoutGetResponse time.Duration                  `yaml:"timeout_get_response"` // таймаут получения ответа
	TimeoutEndResponse time.Duration                  `yaml:"timeout_end_response"` // таймаут окончания ответа
	MaxAttemptsRead    int                            `yaml:"max_attempts_read"`    //число попыток получения ответа
	Pause              time.Duration                  `yaml:"pause"`                //пауза перед опросом
	NetAddr            modbus.DevCmd                  `yaml:"net_addr"`
	Vars               []Vars                         `yaml:"vars,flow"`
	CfsList            []Cfs                          `yaml:"cfs_list,flow"`
	VarsFormat         map[modbus.Var]FloatBitsFormat `yaml:"vars_format"`
	VarsNames          map[modbus.Var]string          `yaml:"vars_names"`
	CfsNames           map[Kef]string                 `yaml:"cfs_names"`
	Commands           map[modbus.DevCmd]string       `yaml:"commands_names"`
}

type Kef uint16

type Cfs [2]Kef

func (c Cfs) Validate() error {
	if c[0] < 0 {
		return merry.New("значение Range[0] должно быть не меньше нуля")
	}
	if c[0] > c[1] {
		return merry.New("значение Range[1] должно быть меньше значения Range[0]")
	}
	return nil
}

type PartyParams = map[string]string

type FloatBitsFormat = modbus.FloatBitsFormat

type Vars [2]modbus.Var

func (p Vars) Var() modbus.Var {
	return p[0]
}

func (p Vars) Count() int {
	return int(p[1])
}

func (p Vars) Validate() error {

	if p.Count() < 1 {
		return merry.Errorf(`не правильное значение count=%d: должно быть боьше нуля`, p.Count)
	}
	if p.Var() < 0 {
		return merry.Errorf(`не правильное знaчение reg=%d: должно быть не меньше нуля`, p.Var)
	}
	if p.Var()+p[1] > 0xFFFF {
		return merry.Errorf(`не правильное значение сумы значений reg+count=%d: должно быть не больше 0xFFFF`,
			p.Var()+p[1])
	}
	return nil
}

func (d Device) paramAddresses() (ps []modbus.Var) {
	for _, p := range d.Vars {
		for i := 0; i < p.Count(); i++ {
			ps = append(ps, p.Var()+modbus.Var(i)*2)
		}
	}
	sort.Slice(ps, func(i, j int) bool {
		return ps[i] < ps[j]
	})
	return
}

func (d Device) VarFormat(Var modbus.Var) FloatBitsFormat {
	if d.VarsFormat != nil {
		if fmt, f := d.VarsFormat[Var]; f {
			return fmt
		}
	}
	return d.FloatFormat
}

func (d Device) Validate() error {

	if err := d.FloatFormat.Validate(); err != nil {
		return err
	}

	if d.VarsFormat != nil {
		for k, v := range d.VarsFormat {
			if err := v.Validate(); err != nil {
				return merry.Appendf(err, "var %d", k)
			}
		}
	}

	if len(d.Vars) == 0 {
		return merry.New("список групп параметров устройства не должен быть пустым")
	}

	if d.Pause < 0 {
		return merry.Errorf(`не правильное значение pause=%v: должно быть не меньше нуля`, d.Pause)
	}
	if d.MaxAttemptsRead < 0 {
		return merry.Errorf(`не правильное значение max_attempts_read=%v: должно быть не меньше нуля`, d.MaxAttemptsRead)
	}
	if d.TimeoutGetResponse < 0 {
		return merry.Errorf(`не правильное значение timeout_get_response=%v: должно быть не меньше нуля`, d.TimeoutGetResponse)
	}
	if d.TimeoutEndResponse < 0 {
		return merry.Errorf(`не правильное значение timeout_end_response=%v: должно быть не меньше нуля`, d.TimeoutEndResponse)
	}
	if d.MaxAttemptsRead < 0 {
		return merry.Errorf(`не правильное значение max_attempts_read=%v: должно быть не меньше нуля`, d.MaxAttemptsRead)
	}
	if d.Baud < 0 {
		return merry.Errorf(`не правильное значение baud=%v: должно быть не меньше нуля`, d.Baud)
	}

	for _, p := range d.Vars {
		if err := p.Validate(); err != nil {
			return merry.Appendf(err, `группа параметров %+v`, p)
		}
	}

	for i, c := range d.CfsList {
		if err := c.Validate(); err != nil {
			return merry.Appendf(err, "диапазон к-тов номер %d", i)
		}
	}

	return nil
}

func (d Device) CommConfig() comm.Config {
	return comm.Config{
		TimeoutGetResponse: d.TimeoutGetResponse,
		TimeoutEndResponse: d.TimeoutEndResponse,
		MaxAttemptsRead:    d.MaxAttemptsRead,
		Pause:              d.Pause,
	}
}

func (d Device) ListCoefficients() (xs []Kef) {
	m := map[Kef]struct{}{}
	for _, p := range d.CfsList {
		for i := p[0]; i <= p[1]; i++ {
			m[i] = struct{}{}
		}
	}
	for i := range m {
		xs = append(xs, i)
	}
	sort.Slice(xs, func(i, j int) bool {
		return xs[i] < xs[j]
	})
	return
}
