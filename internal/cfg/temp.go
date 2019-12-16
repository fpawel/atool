package cfg

import (
	"github.com/ansel1/merry"
	"github.com/fpawel/hardware/temp/ktx500"
	"time"
)

type Temperature struct {
	Type               TempDevType   `yaml:"type"`
	Comport            string        `yaml:"comport"`
	TimeoutGetResponse time.Duration `yaml:"timeout_get_response"`
	TimeoutEndResponse time.Duration `yaml:"timeout_end_response"`
	MaxAttemptsRead    int           `yaml:"max_attempts_read"`
	Ktx500             ktx500.Config `yaml:"ktx500"`
}

func (c Temperature) Validate() error {
	if err := c.Type.Validate(); err != nil {
		return err
	}
	return nil
}

type TempDevType byte

func (c TempDevType) Validate() error {
	switch c {
	case T800:
		return nil
	case T2500:
		return nil
	case Ktx500:
		return nil
	default:
		return merry.Errorf("не правильный тип термокамеры: %d", c)
	}
}

const (
	T800 TempDevType = iota
	T2500
	Ktx500
)
