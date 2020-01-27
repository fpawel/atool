package config

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
	Baud               int           `yaml:"baud"`
	TimeoutGetResponse time.Duration `yaml:"timeout_get_response"` // таймаут получения ответа
	TimeoutEndResponse time.Duration `yaml:"timeout_end_response"` // таймаут окончания ответа
	MaxAttemptsRead    int           `yaml:"max_attempts_read"`    //число попыток получения ответа
	Pause              time.Duration `yaml:"pause"`                //пауза перед опросом
	NetAddr            NetAddr       `yaml:"net_addr"`
	Params             []Params      `yaml:"params"`
}

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

func (d Device) Validate() error {

	//if len(d.Name) < 0 {
	//	return merry.New("не задано имя устройства")
	//}
	//
	//if re := regexp.MustCompile("\\s+"); re.MatchString(d.Name) {
	//	return merry.New(`имя устройства не должно содержать пробелов`)
	//}

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

	for i, x := range d.Params {
		for j, y := range d.Params {
			if i == j {
				continue
			}
			if x.ParamAddr >= y.ParamAddr && x.ParamAddr < y.Count*2 {
				return merry.Errorf(`перекрываются адреса регистров групп параметров номер %d и номер %d`, i, j)
			}
		}
	}
	if err := d.NetAddr.Format.Validate(); err != nil {
		return merry.Append(err, "net_addr.format")
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
