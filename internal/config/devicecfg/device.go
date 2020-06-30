package devicecfg

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"sort"
	"time"
)

type Device struct {
	Baud               int                   `yaml:"baud"`
	TimeoutGetResponse time.Duration         `yaml:"timeout_get_response"` // таймаут получения ответа
	TimeoutEndResponse time.Duration         `yaml:"timeout_end_response"` // таймаут окончания ответа
	MaxAttemptsRead    int                   `yaml:"max_attempts_read"`    //число попыток получения ответа
	Pause              time.Duration         `yaml:"pause"`                //пауза перед опросом
	NetAddr            NetAddr               `yaml:"net_addr"`
	ParamsList         []Params              `yaml:"params_list"`
	CfsList            []Cfs                 `yaml:"cfs_list"`
	ParamsNames        map[modbus.Var]string `yaml:"params_names"`
	CfsNames           map[Kef]string        `yaml:"cfs_names"`
	ProductTypesVars   []ProductTypeVars     `yaml:"product_type_vars"`
}

type ProductTypeVars struct {
	Names         []string `yaml:"names"`
	ParamsRngList []Params `yaml:"list"`
}

type Kef uint16

type Cfs struct {
	Range  [2]Kef          `yaml:"range,flow"`
	Format FloatBitsFormat `yaml:"format"`
}

func (c Cfs) Validate() error {
	if c.Range[0] < 0 {
		return merry.New("значение Range[0] должно быть не меньше нуля")
	}
	if c.Range[0] > c.Range[1] {
		return merry.New("значение Range[1] должно быть меньше значения Range[0]")
	}
	return c.Format.Validate()
}

type PartyParams = map[string]string

type NetAddr struct {
	Cmd    modbus.DevCmd          `yaml:"cmd"`
	Format modbus.FloatBitsFormat `yaml:"format"`
}

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

func (d Device) ParamsRng(prodType string) []Params {
	xs := d.ParamsList
	for _, y := range d.ProductTypesVars {
		for _, s := range y.Names {
			if s == prodType {
				xs = append(xs, y.ParamsRngList...)
			}
		}
	}
	return xs
}

func (d Device) BufferSize(prodType string) (r int) {
	for _, p := range d.ParamsRng(prodType) {
		x := p.ParamAddr*2 + p.Count*4
		if r < int(x) {
			r = int(x)
		}
	}
	return
}

func (d Device) ParamAddresses(prodType string) (ps []modbus.Var) {
	for _, p := range d.ParamsRng(prodType) {
		for i := 0; i < int(p.Count); i++ {
			ps = append(ps, p.ParamAddr+modbus.Var(i)*2)
		}
	}
	sort.Slice(ps, func(i, j int) bool {
		return ps[i] < ps[j]
	})
	return
}

func (d Device) ParamName(paramAddr modbus.Var) string {
	for n, s := range d.ParamsNames {
		if n == paramAddr {
			return fmt.Sprintf("%d: %s", paramAddr, s)
		}
	}
	return fmt.Sprintf("%d", paramAddr)
}

func (d Device) Validate() error {

	if len(d.ParamsList) == 0 {
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

	for _, p := range d.ParamsRng("") {
		if err := p.Validate(); err != nil {
			return merry.Appendf(err, `группа параметров %+v`, p)
		}
	}
	m := make(map[modbus.Var]struct{})
	for _, x := range d.ParamAddresses("") {
		if _, f := m[x]; f {
			return merry.Errorf(`дублирование адреса параметра %d`, x)
		}
		m[x] = struct{}{}
	}

	for _, y := range d.ProductTypesVars {
		for _, s := range y.Names {
			for _, p := range d.ParamsRng(s) {
				if err := p.Validate(); err != nil {
					return merry.Appendf(err, `группа параметров %s: %+v`, s, p)
				}
			}

			m := make(map[modbus.Var]struct{})
			for _, x := range d.ParamAddresses(s) {
				if _, f := m[x]; f {
					return merry.Errorf(`дублирование адреса параметра %s: %d`, s, x)
				}
				m[x] = struct{}{}
			}
		}
	}

	if err := d.NetAddr.Format.Validate(); err != nil {
		return merry.Append(err, "net_addr.format")
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

func (d Device) GetCoefficientFormat(n Kef) (FloatBitsFormat, error) {
	for _, c := range d.CfsList {
		if err := c.Validate(); err != nil {
			return "", merry.Prependf(err, "коэффициент %d: %+v", n, c)
		}
		if n >= c.Range[0] && n <= c.Range[1] {
			return c.Format, nil
		}
	}
	return "", merry.Errorf("коэффициент %d не найден в настройках", n)
}

func (d Device) ListCoefficients() (xs []Kef) {
	m := map[Kef]struct{}{}
	for _, p := range d.CfsList {
		for i := p.Range[0]; i <= p.Range[1]; i++ {
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
