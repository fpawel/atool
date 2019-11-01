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
)

type Config struct {
	LogComm  bool     `yaml:"log_comm"` // выводить посылки приёмопередачи в консоль
	Comports []string `yaml:"comports"` // СОМ порты
}

func SetYaml(strYaml string) error {
	if err := yaml.Unmarshal([]byte(strYaml), &config); err != nil {
		return err
	}
	mu.Lock()
	defer mu.Unlock()
	comm.SetEnableLog(config.LogComm)
	mustWrite([]byte(strYaml))
	return nil
}

func GetYaml() string {
	mu.Lock()
	defer mu.Unlock()
	return string(must.MarshalYaml(&config))
}

func Set(v Config) {
	mu.Lock()
	defer mu.Unlock()
	data := must.MarshalYaml(&v)
	must.UnmarshalYaml(data, &config)
	comm.SetEnableLog(config.LogComm)
	mustWrite(data)
	return
}

func Get() (result Config) {
	mu.Lock()
	defer mu.Unlock()
	must.UnmarshalYaml(must.MarshalYaml(&config), &result)
	return
}

func mustWrite(b []byte) {
	must.WriteFile(filename(), b, 0666)
}

func filename() string {
	return filepath.Join(filepath.Dir(os.Args[0]), "config.toml")
}

var (
	mu     sync.Mutex
	config = func() Config {
		x := Config{
			LogComm:  false,
			Comports: []string{"COM1"},
		}
		filename := filename()

		mustWrite := func() {
			mustWrite(must.MarshalYaml(&x))
		}

		if _, err := os.Stat(filename); os.IsNotExist(err) {
			mustWrite()
		}

		data, err := ioutil.ReadFile(filename)
		must.PanicIf(err)

		if err = yaml.Unmarshal(data, &x); err != nil {
			fmt.Println(err, "file:", filename)
			mustWrite()
		}

		comm.SetEnableLog(x.LogComm)
		return x
	}()
)
