package config

import (
	"github.com/ansel1/merry"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/hardware/gas"
	"time"
)

type Gas struct {
	Type               gas.DevType   `yaml:"type"`
	Addr               modbus.Addr   `yaml:"addr"`
	Comport            string        `yaml:"comport"`
	TimeoutGetResponse time.Duration `yaml:"timeout_get_response"` // таймаут получения ответа
	TimeoutEndResponse time.Duration `yaml:"timeout_end_response"` // таймаут окончания ответа
	MaxAttemptsRead    int           `yaml:"max_attempts_read"`    //число попыток получения ответа
	BlowDuration       time.Duration `yaml:"blow_duration"`
}

func (c Gas) Validate() error {
	if _, f := map[gas.DevType]struct{}{
		gas.Mil82:   {},
		gas.Lab73CO: {},
	}[c.Type]; !f {
		return merry.Errorf("не правильный тип газового блока: %d", c.Type)
	}
	return nil
}
