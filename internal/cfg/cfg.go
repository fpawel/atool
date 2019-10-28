package cfg

import (
	"github.com/BurntSushi/toml"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/comm"
	"github.com/powerman/structlog"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type Config struct {
	LogComm  bool        `toml:"log_comm" comment:"выводить посылки приёмопередачи в консоль"`
	Comports []Comport   `toml:"comports" comment:"СОМ порты"`
}

type Comport struct {
	Name                  string `toml:"name" comment:"имя СОМ порта"`
	Baud                  int    `toml:"baud" comment:"скорость приёмопередачи СОМ порта"`
	ReadTimeoutMillis     int    `toml:"read_timeout" comment:"таймаут получения ответа, мс"`
	ReadByteTimeoutMillis int    `toml:"read_byte_timeout" comment:"таймаут окончания ответа, мс"`
	MaxAttemptsRead       int    `toml:"max_attempts_read" comment:"число попыток получения ответа"`
	PauseMillis           int    `toml:"pause" comment:"пауза перед опросом, мс"`
}

func Open(log *structlog.Logger) {
	defer func() {
		comm.SetEnableLog(config.LogComm)
	}()

	fileName := fileName()

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		save()
	}
	data, err := ioutil.ReadFile(fileName)
	must.PanicIf(err)
	if err = toml.Unmarshal(data, &config); err != nil {
		log.PrintErr(err, "file", fileName)
		save()
	}
}

func SetToml(strToml string) error {
	mu.Lock()
	defer mu.Unlock()
	if err := toml.Unmarshal([]byte(strToml), &config); err != nil {
		return err
	}
	comm.SetEnableLog(config.LogComm)
	write([]byte(strToml))
	return nil
}

func Set(v Config) {
	mu.Lock()
	defer mu.Unlock()
	must.UnmarshalJSON(must.MarshalJSON(&v), &config)
	comm.SetEnableLog(config.LogComm)
	save()
	return
}

func Get() (result Config) {
	mu.Lock()
	defer mu.Unlock()
	must.UnmarshalJSON(must.MarshalJSON(&config), &result)
	return
}

func fileName() string {
	return filepath.Join(filepath.Dir(os.Args[0]), "config.toml")
}
func save() {
	write(must.MarshalToml(&config))
}
func write(data []byte) {
	must.WriteFile(fileName(), data, 0666)
}

var (
	mu     sync.Mutex
	config = Config{
		LogComm: false,
		Comports: []Comport{
			{
				Baud:                  9600,
				Name:                  "COM1",
				ReadByteTimeoutMillis: 50,
				ReadTimeoutMillis:     700,
				MaxAttemptsRead:       3,
			},
		},
	}
)
