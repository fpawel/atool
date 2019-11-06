package app

import (
	"time"
)

type appConfig struct {
	LogComport bool           `yaml:"log_comport" db:"log_comport"`
	Hardware   []deviceConfig `yaml:"hardware"`
}

type deviceConfig struct {
	Name               bool          `yaml:"name"`
	Baud               int           `yaml:"baud"`
	TimeoutGetResponse time.Duration `yaml:"timeout_get_response"`
	TimeoutEndResponse time.Duration `yaml:"timeout_end_response"`
	MaxAttemptsRead    int           `yaml:"max_attempts_read"`
	Pause              time.Duration `yaml:"pause"`
	Params             []paramConfig `yaml:"params"`
}

type paramConfig struct {
	Var    int    `yaml:"var"`
	Count  int    `yaml:"count"`
	Format string `yaml:"format"`
}
