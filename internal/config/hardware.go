package config

import (
	"github.com/ansel1/merry"
	"sort"
)

type Hardware map[string]Device

func (xs Hardware) ParamAddresses(devices map[string]struct{}) (ps []int) {
	m := map[int]struct{}{}
	for name, p := range xs {
		if _, f := devices[name]; !f {
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
		return merry.New("список устройств не должен быть пустым")
	}
	return nil
}

func (xs Hardware) ListDevices() (r []string) {
	for name := range xs {
		r = append(r, name)
	}
	return
}
