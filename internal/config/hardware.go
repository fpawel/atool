package config

import (
	"fmt"
	"github.com/ansel1/merry"
	"sort"
)

type Hardware map[string]Device

func (xs Hardware) GetDevice(deviceType string) (Device, error) {
	device, f := xs[deviceType]
	if !f {
		return Device{}, fmt.Errorf("не заданы параметры устройства %s", deviceType)
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
	return nil
}

func (xs Hardware) ListDevices() (r []string) {
	for name := range xs {
		r = append(r, name)
	}
	return
}

func (xs Hardware) ListProductTypes(deviceType string) []string {
	d, _ := xs[deviceType]
	return d.ProductTypes
}
