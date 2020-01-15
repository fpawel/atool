package config

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/comm"
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
	BlowGas              time.Duration    `yaml:"blow_gas"`
	HoldTemperature      time.Duration    `yaml:"hold_temperature"`
	FloatPrecision       int              `yaml:"float_precision"`
	ProductTypes         []string         `yaml:"product_types"`
	PartyParams          PartyParams      `yaml:"party_params"`
	ProductParams        ProductParams    `yaml:"product_params"`
	Hardware             Hardware         `yaml:"hardware"`
	Gas                  Gas              `yaml:"gas"`
	Temperature          Temperature      `yaml:"temperature"`
	Ktx500               ktx500.Config    `yaml:"ktx500"`
	Coefficients         []Coefficients   `yaml:"coefficients"`
	InactiveCoefficients map[int]struct{} `yaml:"inactive_coefficients"`
	ParamsNames          map[int]string   `yaml:"params_names"`
}

type ProductParams = map[string]map[string]string

type PartyParams = map[string]string

func SetYaml(strYaml []byte) error {
	var c Config
	if err := yaml.Unmarshal(strYaml, &c); err != nil {
		return err
	}
	c.validate()
	if err := c.Validate(); err != nil {
		return err
	}
	comm.SetEnableLog(c.LogComm)
	mu.Lock()
	defer mu.Unlock()
	must.PanicIf(writeFile(strYaml))
	cfg = c
	return nil
}

func Get() (r Config) {
	mu.Lock()
	defer mu.Unlock()
	must.UnmarshalJson(must.MarshalJson(cfg), &r)
	return
}

func Set(c Config) error {
	c.validate()
	if err := c.Validate(); err != nil {
		return err
	}
	b := must.MarshalYaml(c)
	mu.Lock()
	defer mu.Unlock()
	if err := writeFile(b); err != nil {
		return err
	}
	comm.SetEnableLog(c.LogComm)
	cfg = c
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

	if len(c.ProductTypes) == 0 {
		return merry.New("список исполнений партии не должен быть пустым")
	}

	if len(c.PartyParams) == 0 {
		return merry.New("список параметров партии не должен быть пустым")
	}

	if len(c.ProductParams) == 0 {
		return merry.New("список параметров приборов не должен быть пустым")
	}

	m := make(map[string]struct{})
	for _, x := range c.ProductTypes {
		if _, f := m[x]; f {
			return merry.Errorf("дублирование исполнения партии %q", x)
		}
		m[x] = struct{}{}
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

func writeFile(b []byte) error {
	return ioutil.WriteFile(filename(), b, 0666)
}

func filename() string {
	return filepath.Join(filepath.Dir(os.Args[0]), "config.yaml")
}

func readFile() (Config, error) {
	var c Config
	data, err := ioutil.ReadFile(filename())
	if err != nil {
		return c, err
	}
	err = yaml.Unmarshal(data, &c)
	return c, err
}

func (c *Config) validate() {
	if len(c.ProductTypes) == 0 {
		c.ProductTypes = []string{"00.01", "00.02"}
	}

	if len(c.PartyParams) == 0 {
		c.PartyParams = defaultPartyParams()
	}

	if len(c.ProductParams) == 0 {
		c.ProductParams = defaultProductParams()
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
	if c.ParamsNames == nil {
		c.ParamsNames = map[int]string{
			0: "C",
			2: "I",
		}
	}
}

func init() {
	var err error
	c, err := readFile()
	if err != nil {
		fmt.Println(err, "file:", filename())
		c = defaultConfig()
	}
	c.validate()
	if err := c.Validate(); err != nil {
		fmt.Println(err)
		c = defaultConfig()
	}
	must.PanicIf(Set(c))
}

var (
	mu  sync.Mutex
	cfg = defaultConfig()
)
