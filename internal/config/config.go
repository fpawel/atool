package config

import (
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/cfgfile"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/hardware/temp/ktx500"
	"gopkg.in/yaml.v3"
	"os"
)

type Kef = devicecfg.Kef

type Config struct {
	Hardware       devicecfg.Hardware `yaml:"hardware"`
	LogComm        bool               `yaml:"log_comm"`
	FloatPrecision int                `yaml:"float_precision"`
	Gas            Gas                `yaml:"gas"`
	Temperature    Temperature        `yaml:"temperature"`
	WarmSheets     Mil82WarmSheets    `yaml:"warm_sheets"`
	Ktx500         ktx500.Config      `yaml:"ktx500"`
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

func (c Config) addHardware(hardware devicecfg.Hardware) {
	for name, d := range hardware {
		if _, f := c.Hardware[name]; !f {
			c.Hardware[name] = d
		}
	}
}

func LoadOrDefault(hardware devicecfg.Hardware) (Config, error) {
	if _, err := os.Stat(Filename()); os.IsNotExist(err) {
		c := defaultConfig()
		c.addHardware(hardware)
		return c, nil
	}
	return Load(hardware)
}

func Load(hardware devicecfg.Hardware) (Config, error) {
	var x Config
	if err := file.Get(&x); err != nil {
		return x, err
	}
	if err := x.Validate(); err != nil {
		return x, err
	}
	if len(x.Hardware) == 0 {
		x.Hardware = devicecfg.Hardware{}
	}
	x.addHardware(hardware)
	comm.SetEnableLog(x.LogComm)
	return x, nil
}

func Filename() string {
	return file.Filename()
}

var file = cfgfile.New("config.yaml", yaml.Marshal, yaml.Unmarshal)
