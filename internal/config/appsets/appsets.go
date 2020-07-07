package appsets

import (
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/pkg/cfgfile"
	"gopkg.in/yaml.v3"
	"os"
)

type Settings struct {
	InactiveCoefficients map[config.Kef]struct{} `yaml:"inactive_coefficients"`
}

func (x *Settings) Save() error {
	return file.Set(x)
}

func (x *Settings) Open() error {
	if _, err := os.Stat(file.Filename()); os.IsNotExist(err) {
		x.InactiveCoefficients = make(map[config.Kef]struct{})
		return nil
	}
	return file.Get(x)
}

var (
	file = cfgfile.New("settings.yaml", yaml.Marshal, yaml.Unmarshal)
)
