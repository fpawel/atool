package config

import (
	"fmt"
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/devtypes"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/hardware/temp/ktx500"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type Config struct {
	LogComm              bool             `yaml:"log_comm"`
	FloatPrecision       int              `yaml:"float_precision"`
	Hardware             Hardware         `yaml:"hardware"`
	Gas                  Gas              `yaml:"gas"`
	Temperature          Temperature      `yaml:"temperature"`
	WarmSheets           Mil82WarmSheets  `yaml:"warm_sheets"`
	Ktx500               ktx500.Config    `yaml:"ktx500"`
	InactiveCoefficients map[int]struct{} `yaml:"inactive_coefficients"`
}

type Mil82WarmSheets struct {
	Enable  bool        `yaml:"enable"`
	Addr    modbus.Addr `yaml:"addr"`
	TempOn  float64     `yaml:"temp_on"`
	TempOff float64     `yaml:"temp_off"`
}

func SetYaml(strYaml []byte) error {
	var c Config
	if err := yaml.Unmarshal(strYaml, &c); err != nil {
		return err
	}
	c.validate()
	if err := c.Validate(); err != nil {
		return err
	}
	comm.SetEnableLog(c.LogComm)
	mu.Lock()
	defer mu.Unlock()
	must.PanicIf(writeFile(strYaml))
	cfg = c
	return nil
}

func Get() (r Config) {
	mu.Lock()
	defer mu.Unlock()
	must.UnmarshalJson(must.MarshalJson(cfg), &r)
	return
}

func Set(c Config) error {
	c.validate()
	if err := c.Validate(); err != nil {
		return err
	}
	b := must.MarshalYaml(c)
	mu.Lock()
	defer mu.Unlock()
	if err := writeFile(b); err != nil {
		return err
	}
	comm.SetEnableLog(c.LogComm)
	cfg = c
	return nil
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

func writeFile(b []byte) error {
	return ioutil.WriteFile(filename(), b, 0666)
}

func filename() string {
	return filepath.Join(filepath.Dir(os.Args[0]), "config.yaml")
}

func readFile() (Config, error) {
	var c Config
	data, err := ioutil.ReadFile(filename())
	if err != nil {
		return c, err
	}
	err = yaml.Unmarshal(data, &c)
	return c, err
}

func (c *Config) validate() {

	for d := range c.Hardware {
		dv := c.Hardware[d]
		if len(dv.PartyParams) == 0 {
			dv.PartyParams = defaultPartyParams()
		}

		if len(dv.Coefficients) == 0 {
			dv.Coefficients = []devicecfg.Coefficients{
				{
					Range:  [2]int{0, 50},
					Format: "float_big_endian",
				},
			}
		}
		if c.InactiveCoefficients == nil {
			c.InactiveCoefficients = make(map[int]struct{})
		}
		if dv.ParamsNames == nil {
			dv.ParamsNames = map[int]string{
				0: "C",
				2: "I",
			}
		}
		c.Hardware[d] = dv
	}

	for name, d := range devtypes.DeviceTypes {
		if _, f := c.Hardware[name]; !f {
			c.Hardware[name] = d.Config
		}
	}

}

func init() {
	var err error
	c, err := readFile()
	if err != nil {
		fmt.Println(err, "file:", filename())
		c = defaultConfig()
	}
	c.validate()
	if err := c.Validate(); err != nil {
		fmt.Println(err)
		c = defaultConfig()
	}
	must.PanicIf(Set(c))
}

var (
	mu  sync.Mutex
	cfg = defaultConfig()
)
