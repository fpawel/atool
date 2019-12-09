package cfg

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Hardware []Device

type Device struct {
	Name               string        `yaml:"name"`
	Baud               int           `yaml:"baud"`
	TimeoutGetResponse time.Duration `yaml:"timeout_get_response"` // таймаут получения ответа
	TimeoutEndResponse time.Duration `yaml:"timeout_end_response"` // таймаут окончания ответа
	MaxAttemptsRead    int           `yaml:"max_attempts_read"`    //число попыток получения ответа
	Pause              time.Duration `yaml:"pause"`                //пауза перед опросом
	Params             []Params      `yaml:"params"`
}

type Params struct {
	Format    ParamFormat `yaml:"format"`
	ParamAddr int         `yaml:"reg"`
	Count     int         `yaml:"count"`
}

type ParamFormat string

const (
	ParamFormatBCD               ParamFormat = "bcd"
	ParamFormatFloatBigEndian    ParamFormat = "float_big_endian"
	ParamFormatFloatLittleEndian ParamFormat = "float_little_endian"
	ParamFormatIntBigEndian      ParamFormat = "int_big_endian"
	ParamFormatIntLittleEndian   ParamFormat = "int_little_endian"
)

var ParamFormats = map[ParamFormat]struct{}{
	ParamFormatBCD:               {},
	ParamFormatFloatBigEndian:    {},
	ParamFormatFloatLittleEndian: {},
	ParamFormatIntBigEndian:      {},
	ParamFormatIntLittleEndian:   {},
}

func (xs Hardware) ParamAddresses() (ps []int) {
	m := map[int]struct{}{}
	for _, p := range xs {
		for _, p := range p.ParamAddresses() {
			m[p] = struct{}{}
		}
	}
	for p := range m {
		ps = append(ps, p)
	}
	sort.Ints(ps)
	return
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

func (xs Hardware) Validate() error {
	if len(xs) == 0 {
		return errors.New("список устройств не должен быть пустым")
	}
	m := map[string]struct{}{}
	for i, d := range xs {
		if err := d.Validate(); err != nil {
			return fmt.Errorf(`устройство номер %d с именем %q: %w`, i, d.Name, err)
		}
		if _, f := m[d.Name]; f {
			return fmt.Errorf(`дублирование имени устройства: номер %d`, i)
		}
		m[d.Name] = struct{}{}
	}
	return nil
}

func (xs Hardware) DeviceByName(name string) (Device, bool) {
	for _, d := range xs {
		if d.Name == name {
			return d, true
		}
	}
	return Device{}, false
}

func (xs Hardware) ListDevices() (r []string) {
	for _, d := range xs {
		r = append(r, d.Name)
	}
	return
}

func (d Device) Validate() error {

	if len(d.Name) < 0 {
		return errors.New("не задано имя устройства")
	}

	if re := regexp.MustCompile("\\s+"); re.MatchString(d.Name) {
		return errors.New(`имя устройства не должно содержать пробелов`)
	}

	if len(d.Params) == 0 {
		return errors.New("список групп параметров устройства не должен быть пустым")
	}

	if d.Pause < 0 {
		return fmt.Errorf(`не правильное знaчение pause=%v: должно не меньше нуля`, d.Pause)
	}
	if d.MaxAttemptsRead < 0 {
		return fmt.Errorf(`не правильное знaчение max_attempts_read=%v: должно не меньше нуля`, d.MaxAttemptsRead)
	}
	if d.TimeoutGetResponse < 0 {
		return fmt.Errorf(`не правильное знaчение timeout_get_response=%v: должно не меньше нуля`, d.TimeoutGetResponse)
	}
	if d.TimeoutEndResponse < 0 {
		return fmt.Errorf(`не правильное знaчение timeout_end_response=%v: должно не меньше нуля`, d.TimeoutEndResponse)
	}
	if d.MaxAttemptsRead < 0 {
		return fmt.Errorf(`не правильное знaчение max_attempts_read=%v: должно не меньше нуля`, d.MaxAttemptsRead)
	}
	if d.Baud < 0 {
		return fmt.Errorf(`не правильное знaчение baud=%v: должно не меньше нуля`, d.Baud)
	}
	for i, p := range d.Params {
		if err := p.Validate(); err != nil {
			return fmt.Errorf(`группа параметров номер %d: %+v: %w`, i, p, err)
		}
	}

	for i, x := range d.Params {
		for j, y := range d.Params {
			if i == j {
				continue
			}
			if x.ParamAddr >= y.ParamAddr && x.ParamAddr < y.Count*2 {
				return fmt.Errorf(`перекрываются адреса регистров групп параметров номер %d и номер %d`, i, j)
			}
		}
	}
	return nil
}

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
