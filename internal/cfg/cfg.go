package cfg

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/comm"
	"github.com/fpawel/hardware/gas"
	"github.com/fpawel/hardware/temp/ktx500"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type Config struct {
	LogComm              bool             `yaml:"log_comm"`
	FloatPrecision       int              `yaml:"float_precision"`
	Hardware             Hardware         `yaml:"hardware"`
	Gas                  Gas              `yaml:"gas"`
	Temperature          Temperature      `yaml:"temperature"`
	Coefficients         []Coefficients   `yaml:"coefficients"`
	InactiveCoefficients map[int]struct{} `yaml:"inactive_coefficients,flow"`
}

func SetYaml(strYaml []byte) error {
	var c Config
	if err := yaml.Unmarshal(strYaml, &c); err != nil {
		return err
	}
	if err := c.Validate(); err != nil {
		return err
	}
	if c.InactiveCoefficients == nil {
		c.InactiveCoefficients = make(map[int]struct{})
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

func Set(c Config) error {
	if c.InactiveCoefficients == nil {
		c.InactiveCoefficients = make(map[int]struct{})
	}
	b := must.MarshalYaml(c)
	if err := c.Validate(); err != nil {
		return err
	}
	mu.Lock()
	defer mu.Unlock()
	if err := write(b); err != nil {
		return err
	}
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

	for i, c := range c.Coefficients {
		if err := c.Validate(); err != nil {
			return merry.Appendf(err, "диапазон к-тов номер %d", i)
		}
	}
	return nil
}

func (c Config) GetCoefficientFormat(n int) (FloatBitsFormat, error) {
	for _, c := range c.Coefficients {
		if err := c.Validate(); err != nil {
			return "", fmt.Errorf("коэффициент %d: %+v: %w", n, c, err)
		}
		if n >= c.Range[0] && n <= c.Range[1] {
			return c.Format, nil
		}
	}
	return "", fmt.Errorf("коэффициент %d не найден в настройках", n)
}

func (c Config) ListCoefficients() (xs []int) {
	m := map[int]struct{}{}
	for _, p := range c.Coefficients {
		for i := p.Range[0]; i <= p.Range[1]; i++ {
			m[i] = struct{}{}
		}
	}
	for i := range m {
		xs = append(xs, i)
	}
	sort.Ints(xs)
	return
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
			LogComm:        false,
			FloatPrecision: 6,
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
			Gas: Gas{
				Type:               gas.Mil82,
				Addr:               100,
				Comport:            "COM1",
				TimeoutGetResponse: time.Second,
				TimeoutEndResponse: time.Millisecond * 50,
				MaxAttemptsRead:    0,
			},
			Temperature: Temperature{
				Type:               T800,
				Comport:            "COM1",
				TimeoutGetResponse: time.Second,
				TimeoutEndResponse: time.Millisecond * 50,
				MaxAttemptsRead:    1,
				Ktx500:             ktx500.NewDefaultConfig(),
			},
		}
	}
	if len(c.Coefficients) == 0 {
		c.Coefficients = []Coefficients{
			{
				Range:  [2]int{0, 50},
				Format: "float_big_endian",
			},
		}
	}
	if c.InactiveCoefficients == nil {
		c.InactiveCoefficients = make(map[int]struct{})
	}
	must.PanicIf(write(must.MarshalYaml(c)))
	comm.SetEnableLog(c.LogComm)
}
