package config

import (
	"fmt"
	"time"
)

type Temperature struct {
	Type               TempDevType   `yaml:"type"`
	Comport            string        `yaml:"comport"`
	TimeoutGetResponse time.Duration `yaml:"timeout_get_response"`
	TimeoutEndResponse time.Duration `yaml:"timeout_end_response"`
	MaxAttemptsRead    int           `yaml:"max_attempts_read"`
	HoldDuration       time.Duration `yaml:"hold_duration"`
	TempNorm           float64       `yaml:"temp_norm"`
	TempLow            float64       `yaml:"temp_low"`
	TempHigh           float64       `yaml:"temp_high"`
}

func (c Temperature) Validate() error {
	if err := c.Type.Validate(); err != nil {
		return err
	}
	return nil
}

type TempDevType string

func (c TempDevType) Validate() error {
	switch c {
	case T800:
		return nil
	case T2500:
		return nil
	case Ktx500:
		return nil
	default:
		return fmt.Errorf("не правильный тип термокамеры: %s", c)
	}
}

const (
	T800   TempDevType = "T800"
	T2500  TempDevType = "T2500"
	Ktx500 TempDevType = "КТХ-500"
)
