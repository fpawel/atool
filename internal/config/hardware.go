package config

import (
	"errors"
	"fmt"
	"sort"
)

type Hardware []Device

func (xs Hardware) ParamAddresses(devices map[string]struct{}) (ps []int) {
	m := map[int]struct{}{}
	for _, p := range xs {
		if _, f := devices[p.Name]; !f {
			continue
		}
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
