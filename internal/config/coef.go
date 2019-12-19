package config

import (
	"github.com/ansel1/merry"
)

type Coefficients struct {
	Range  [2]int          `yaml:"range,flow"`
	Format FloatBitsFormat `yaml:"format"`
}

func (c Coefficients) Validate() error {
	if c.Range[0] < 0 {
		return merry.New("значение Range[0] должно быть не меньше нуля")
	}
	if c.Range[0] > c.Range[1] {
		return merry.New("значение Range[1] должно быть меньше значения Range[0]")
	}
	return c.Format.Validate()
}
