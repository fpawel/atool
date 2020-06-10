package appcfg

import (
	"github.com/fpawel/atool/internal/config"
	"os"
)

var (
	Cfg config.Config
)

func init() {
	if _, err := os.Stat(config.Filename()); os.IsNotExist(err) {
		Cfg = config.DefaultConfig()
		return
	}
	if err := Cfg.Load(); err != nil {
		panic(err)
	}
}
