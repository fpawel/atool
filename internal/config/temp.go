package config

import (
	"fmt"
	"github.com/ansel1/merry"
	"time"
)

type Temperature struct {
	Type               TempDevType   `yaml:"type"`
	Comport            string        `yaml:"comport"`
	TimeoutGetResponse time.Duration `yaml:"timeout_get_response"`
	TimeoutEndResponse time.Duration `yaml:"timeout_end_response"`
	MaxAttemptsRead    int           `yaml:"max_attempts_read"`
}

func (c Temperature) Validate() error {
	if err := c.Type.Validate(); err != nil {
		return err
	}
	return nil
}

type TempDevType byte

func (c TempDevType) String() string {
	switch c {
	case T800:
		return "T800"
	case T2500:
		return "T2500"
	case Ktx500:
		return "KTX500"
	default:
		return fmt.Sprintf("%d", c)
	}
}

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
