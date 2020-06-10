package config

import (
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/devtypes"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/cfgfile"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/hardware/temp/ktx500"
	"gopkg.in/yaml.v3"
)

type Config struct {
	LogComm              bool               `yaml:"log_comm"`
	FloatPrecision       int                `yaml:"float_precision"`
	Hardware             devicecfg.Hardware `yaml:"hardware"`
	Gas                  Gas                `yaml:"gas"`
	Temperature          Temperature        `yaml:"temperature"`
	WarmSheets           Mil82WarmSheets    `yaml:"warm_sheets"`
	Ktx500               ktx500.Config      `yaml:"ktx500"`
	InactiveCoefficients map[int]struct{}   `yaml:"inactive_coefficients"`
}

type Mil82WarmSheets struct {
	Enable  bool        `yaml:"enable"`
	Addr    modbus.Addr `yaml:"addr"`
	TempOn  float64     `yaml:"temp_on"`
	TempOff float64     `yaml:"temp_off"`
}

func (c Config) FormatFloat(v float64) string {
	return pkg.FormatFloat(v, c.FloatPrecision)
}

func (c Config) Validate() error {
	if err := c.Hardware.Validate(); err != nil {
		return err
	}
	if err := c.Gas.Validate(); err != nil {
		return err
	}
	if err := c.Temperature.Validate(); err != nil {
		return err
	}

	return nil
}

func (c *Config) Save() error {
	return file.Set(c)
}

func (c *Config) Load() error {
	if err := file.Get(c); err != nil {
		return err
	}
	if err := c.Validate(); err != nil {
		return err
	}
	c.addDefinedDevices()
	return nil
}

func (c *Config) addDefinedDevices() {
	if len(c.Hardware) == 0 {
		c.Hardware = devicecfg.Hardware{}
	}
	for name, d := range devtypes.DeviceTypes {
		if _, f := c.Hardware[name]; !f {
			c.Hardware[name] = d.Config
		}
	}
}

func Filename() string {
	return file.Filename()
}

var file = cfgfile.New("config.yaml", yaml.Marshal, yaml.Unmarshal)
