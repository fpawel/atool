package devicecfg

import (
	"github.com/ansel1/merry"
	"sort"
)

type Hardware map[string]Device

func (xs Hardware) GetDevice(deviceType string) (Device, error) {
	device, f := xs[deviceType]
	if !f {
		return Device{}, merry.Errorf("не заданы параметры устройства %s", deviceType)
	}
	return device, nil
}

func (xs Hardware) Validate() error {
	for name, d := range xs {
		if err := d.Validate(); err != nil {
			return merry.Prepend(err, name)
		}
	}
	return nil
}

func (xs Hardware) DeviceNames() (r []string) {
	for name := range xs {
		r = append(r, name)
	}
	sort.Strings(r)
	return
}
