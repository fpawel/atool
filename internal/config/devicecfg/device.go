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
	//Name               string        `yaml:"name"`
	Baud               int            `yaml:"baud"`
	TimeoutGetResponse time.Duration  `yaml:"timeout_get_response"` // таймаут получения ответа
	TimeoutEndResponse time.Duration  `yaml:"timeout_end_response"` // таймаут окончания ответа
	MaxAttemptsRead    int            `yaml:"max_attempts_read"`    //число попыток получения ответа
	Pause              time.Duration  `yaml:"pause"`                //пауза перед опросом
	NetAddr            NetAddr        `yaml:"net_addr"`
	Params             []Params       `yaml:"params"`
	PartyParams        PartyParams    `yaml:"party_params"`
	Coefficients       []Coefficients `yaml:"coefficients"`
	ParamsNames        map[int]string `yaml:"params_names"`
}

type PartyParams = map[string]string

type NetAddr struct {
	Cmd    modbus.DevCmd          `yaml:"cmd"`
	Format modbus.FloatBitsFormat `yaml:"format"`
}

func (d Device) BufferSize() (r int) {
	for _, p := range d.Params {
		x := p.ParamAddr*2 + p.Count*4
		if r < x {
			r = x
		}
	}
	return
}

func (d Device) ParamAddresses() (ps []int) {
	for _, p := range d.Params {
		for i := 0; i < p.Count; i++ {
			ps = append(ps, p.ParamAddr+i*2)
		}
	}
	sort.Ints(ps)
	return
}

func (d Device) ParamName(paramAddr int) string {
	for n, s := range d.ParamsNames {
		if n == paramAddr {
			return fmt.Sprintf("%d: %s", paramAddr, s)
		}
	}
	return fmt.Sprintf("%d", paramAddr)
}

func (d Device) Validate() error {

	if len(d.Params) == 0 {
		return merry.New("список групп параметров устройства не должен быть пустым")
	}

	if d.Pause < 0 {
		return merry.Errorf(`не правильное значение pause=%v: должно не меньше нуля`, d.Pause)
	}
	if d.MaxAttemptsRead < 0 {
		return merry.Errorf(`не правильное значение max_attempts_read=%v: должно не меньше нуля`, d.MaxAttemptsRead)
	}
	if d.TimeoutGetResponse < 0 {
		return merry.Errorf(`не правильное значение timeout_get_response=%v: должно не меньше нуля`, d.TimeoutGetResponse)
	}
	if d.TimeoutEndResponse < 0 {
		return merry.Errorf(`не правильное значение timeout_end_response=%v: должно не меньше нуля`, d.TimeoutEndResponse)
	}
	if d.MaxAttemptsRead < 0 {
		return merry.Errorf(`не правильное значение max_attempts_read=%v: должно не меньше нуля`, d.MaxAttemptsRead)
	}
	if d.Baud < 0 {
		return fmt.Errorf(`не правильное знaчение baud=%v: должно не меньше нуля`, d.Baud)
	}
	for i, p := range d.Params {
		if err := p.Validate(); err != nil {
			return merry.Errorf(`группа параметров номер %d: %+v: %w`, i, p, err)
		}
	}

	m := make(map[int]struct{})
	for _, x := range d.ParamAddresses() {
		if _, f := m[x]; f {
			return merry.Errorf(`дублирование адреса параметра %d`, x)
		}
		m[x] = struct{}{}
	}

	if err := d.NetAddr.Format.Validate(); err != nil {
		return merry.Append(err, "net_addr.format")
	}

	if len(d.PartyParams) == 0 {
		return merry.New("список параметров партии не должен быть пустым")
	}

	for i, c := range d.Coefficients {
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

func (d Device) GetCoefficientFormat(n int) (FloatBitsFormat, error) {
	for _, c := range d.Coefficients {
		if err := c.Validate(); err != nil {
			return "", fmt.Errorf("коэффициент %d: %+v: %w", n, c, err)
		}
		if n >= c.Range[0] && n <= c.Range[1] {
			return c.Format, nil
		}
	}
	return "", fmt.Errorf("коэффициент %d не найден в настройках", n)
}

func (d Device) ListCoefficients() (xs []int) {
	m := map[int]struct{}{}
	for _, p := range d.Coefficients {
		for i := p.Range[0]; i <= p.Range[1]; i++ {
			m[i] = struct{}{}
		}
	}
	for i := range m {
		xs = append(xs, i)
	}
	sort.Ints(xs)
	return
}