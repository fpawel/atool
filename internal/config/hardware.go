package config

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config/devicecfg"
	"sort"
)

type Hardware map[string]devicecfg.Device

func (xs Hardware) GetDevice(deviceType string) (devicecfg.Device, error) {
	device, f := xs[deviceType]
	if !f {
		return devicecfg.Device{}, fmt.Errorf("не заданы параметры устройства %s", deviceType)
	}
	return device, nil
}

func (xs Hardware) GetDeviceParamAddresses(deviceType string) (ps []int) {
	device, _ := xs[deviceType]
	for _, p := range device.Params {
		for n := 0; n < p.Count; n++ {
			ps = append(ps, p.ParamAddr+2*n)
		}

	}
	sort.Ints(ps)
	return
}

func (xs Hardware) Validate() error {
	if len(xs) == 0 {
		return merry.New("список устройств не должен быть пустым")
	}
	for name, d := range xs {
		if err := d.Validate(); err != nil {
			return merry.Prepend(err, name)
		}
	}
	return nil
}

func (xs Hardware) ListDevices() (r []string) {
	for name := range xs {
		r = append(r, name)
	}
	return
}
