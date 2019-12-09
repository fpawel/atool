package cfg

import (
	"fmt"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/comm"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Config struct {
	LogComm  bool     `yaml:"log_comm"`
	Hardware Hardware `yaml:"hardware"`
}

func SetYaml(strYaml []byte) error {
	var c Config
	if err := yaml.Unmarshal(strYaml, &c); err != nil {
		return err
	}
	if err := c.Hardware.Validate(); err != nil {
		return err
	}
	comm.SetEnableLog(c.LogComm)
	mu.Lock()
	defer mu.Unlock()
	return write(strYaml)
}

func Get() Config {
	mu.Lock()
	defer mu.Unlock()
	r, err := read()
	must.PanicIf(err)
	return r
}

func write(b []byte) error {
	return ioutil.WriteFile(filename(), b, 0666)
}

func filename() string {
	return filepath.Join(filepath.Dir(os.Args[0]), "config.yaml")
}

func read() (Config, error) {
	var c Config
	data, err := ioutil.ReadFile(filename())
	if err != nil {
		return c, err
	}
	err = yaml.Unmarshal(data, &c)
	return c, err
}

var mu sync.Mutex

func init() {
	c, err := read()
	if err != nil {
		fmt.Println(err, "file:", filename())
		c = Config{
			LogComm: false,
			Hardware: Hardware{
				Device{
					Name:               "default",
					Baud:               9600,
					TimeoutGetResponse: time.Second,
					TimeoutEndResponse: time.Millisecond * 50,
					MaxAttemptsRead:    0,
					Pause:              0,
					Params: []Params{
						{
							Format:    "bcd",
							ParamAddr: 0,
							Count:     1,
						},
					},
				},
			},
		}
		must.PanicIf(write(must.MarshalYaml(c)))
	}
	comm.SetEnableLog(c.LogComm)
}
