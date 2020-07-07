package appcfg

import (
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/config/appsets"
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/devtypes/devdata"
)

var (
	DeviceTypes map[string]devdata.Device
	Cfg         config.Config
	Sets        = new(appsets.Settings)
)

func Init(devices ...devdata.Device) {
	DeviceTypes = make(map[string]devdata.Device)
	for _, d := range devices {
		DeviceTypes[d.Name] = d
	}
	var err error
	Cfg, err = config.LoadOrDefault(hardware())
	if err != nil {
		panic(err)
	}
	if err := Sets.Open(); err != nil {
		panic(err)
	}
}

func Reload() error {
	c, err := config.LoadOrDefault(hardware())
	if err != nil {
		return err
	}
	Cfg = c
	return nil
}

func hardware() devicecfg.Hardware {
	hardware := devicecfg.Hardware{}
	for name, d := range DeviceTypes {
		hardware[name] = d.Config
	}
	return hardware
}
