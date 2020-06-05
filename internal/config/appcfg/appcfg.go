package appcfg

import (
	"database/sql"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/devtypes"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/hardware/gas"
	"sort"
	"strconv"
	"strings"
	"time"
)

func GetParamValue(key string) (string, error) {
	if v, f := Params[key]; f {
		c := config.Get()
		return v.get(&c), nil
	}
	const q1 = `SELECT value FROM party_value WHERE party_id = (SELECT party_id FROM app_config) AND key = ?`
	var r string
	err := data.DB.Get(&r, q1, key)
	if err == sql.ErrNoRows {
		err = merry.Errorf("значение ключа партии %q не задано", key)
	}
	return r, err
}

func GetParamsValues() ([]*apitypes.ConfigParamValue, error) {

	p, err := data.GetCurrentParty()
	if err != nil {
		return nil, err
	}

	cfg := config.Get()

	device, _ := devtypes.DeviceTypes[p.DeviceType]

	xs := []*apitypes.ConfigParamValue{
		{
			Key:   "name",
			Name:  "Приборы: имя файла",
			Value: p.Name,
		},
		{
			Key:        "device_type",
			Name:       "Приборы: тип приборов",
			Value:      p.DeviceType,
			ValuesList: cfg.Hardware.ListDevices(),
		},
		{
			Key:        "product_type",
			Name:       "Приборы: исполнение",
			Value:      p.ProductType,
			ValuesList: device.ProductTypes,
		},
	}

	checkKey := func(k string) error {
		for _, x := range xs {
			if x.Key == k {
				return merry.Errorf("дублирование значений ключа %q", k)
			}
		}
		return nil
	}

	for k, x := range Params {
		if err := checkKey(k); err != nil {
			return nil, err
		}
		xs = append(xs, &apitypes.ConfigParamValue{
			Key:        k,
			Name:       x.Name,
			Type:       x.Type,
			ValuesList: x.List(),
			Value:      x.get(&cfg),
		})
	}

	m, err := getCurrentPartyValues()
	if err != nil {
		return nil, err
	}

	partyParamsOfDevice := func() devicecfg.PartyParams {
		dv, _ := cfg.Hardware[p.DeviceType]
		for k, v := range dv.PartyParams {
			dv.PartyParams[k] = "Приборы: " + v
		}
		return dv.PartyParams
	}()

	for key, name := range partyParamsOfDevice {
		if err := checkKey(key); err != nil {
			return nil, err
		}
		y := &apitypes.ConfigParamValue{
			Key:  key,
			Name: name,
			Type: "float",
		}
		if v, f := m[key]; f {
			y.Value = cfg.FormatFloat(v)
		}
		xs = append(xs, y)
	}
	sort.Slice(xs, func(i, j int) bool {
		return xs[i].Name < xs[j].Name
	})
	return xs, nil
}

type Param struct {
	Name string
	Type string
	list func() []string
	set  func(*config.Config, string) error
	get  func(*config.Config) string
}

func (x Param) Set(cfg *config.Config, value string) error {
	return x.set(cfg, value)
}

func (x Param) Get(cfg *config.Config) string {
	return x.get(cfg)
}

func (x Param) List() []string {
	if x.Type == "comport" {
		comports, _ := comport.Ports()
		return comports
	}
	if x.list == nil {
		return make([]string, 0)
	}
	return x.list()
}

var Params = map[string]Param{

	"temperature_type": {
		Name: "Термокамера: тип",
		list: func() []string {
			return []string{string(config.T800), string(config.T2500), string(config.Ktx500)}
		},
		get: func(c *config.Config) string {
			return string(c.Temperature.Type)
		},
		set: func(c *config.Config, s string) error {
			c.Temperature.Type = config.TempDevType(s)
			return nil
		},
	},

	"temperature_comport": {
		Name: "Термокамера: СОМ порт",
		Type: "comport",
		set: func(c *config.Config, s string) error {
			c.Temperature.Comport = s
			return nil
		},
		get: func(c *config.Config) string {
			return c.Temperature.Comport
		},
	},

	"temperature_hold_duration": newDurationParam("Термокамера: длительность выдержки",
		func(c *config.Config) *time.Duration {
			return &c.Temperature.HoldDuration
		}),

	"gas_blow_duration": newDurationParam("Газовый блок: длительность продувки",
		func(c *config.Config) *time.Duration {
			return &c.Gas.BlowDuration
		}),

	"gas_address": {
		Name: "Газовый блок: адрес",
		Type: "int",
		set: func(c *config.Config, s string) error {
			n, err := strconv.ParseInt(s, 10, 8)
			if err != nil {
				return err
			}
			c.Gas.Addr = modbus.Addr(n)
			return nil
		},
		get: func(c *config.Config) string {
			return strconv.Itoa(int(c.Gas.Addr))
		},
	},

	"gas_comport": {
		Name: "Газовый блок: СОМ порт",
		Type: "comport",
		set: func(c *config.Config, s string) error {
			c.Gas.Comport = s
			return nil
		},
		get: func(c *config.Config) string {
			return c.Gas.Comport
		},
	},

	"gas_type": {
		Name: "Газовый блок: тип",
		list: func() []string {
			return []string{string(gas.Mil82), string(gas.Lab73CO)}
		},
		get: func(c *config.Config) string {
			return string(c.Gas.Type)
		},
		set: func(c *config.Config, s string) error {
			c.Gas.Type = gas.DevType(s)
			return nil
		},
	},

	"warm_sheets_enable": {
		Name: "Подогрев плат: использовать",
		Type: "bool",
		get: func(c *config.Config) string {
			return strconv.FormatBool(c.WarmSheets.Enable)
		},
		set: func(c *config.Config, s string) error {
			v, err := strconv.ParseBool(s)
			if err != nil {
				return err
			}
			c.WarmSheets.Enable = v
			return nil
		},
	},
	"warm_sheets_address": {
		Name: "Подогрев плат: адрес устройства",
		Type: "int",
		get: func(c *config.Config) string {
			return strconv.Itoa(int(c.WarmSheets.Addr))
		},
		set: func(c *config.Config, s string) error {
			v, err := strconv.ParseInt(s, 10, 8)
			if err != nil {
				return err
			}
			c.WarmSheets.Addr = modbus.Addr(v)
			return nil
		},
	},

	"warm_sheets_temp_on": newFloatParam("Подогрев плат: температура включения",
		func(c *config.Config) *float64 {
			return &c.WarmSheets.TempOn
		}),

	"warm_sheets_temp_off": newFloatParam("Подогрев плат: температура выключения",
		func(c *config.Config) *float64 {
			return &c.WarmSheets.TempOff
		}),
}

func newFloatParam(name string, f func(c *config.Config) *float64) Param {
	return Param{
		Name: name,
		Type: "float",
		list: nil,
		set: func(c *config.Config, s string) error {
			v, err := parseFloat(s)
			if err != nil {
				return err
			}
			p := f(c)
			*p = v
			return nil
		},
		get: func(c *config.Config) string {
			p := f(c)
			return config.Get().FormatFloat(*p)
		},
	}
}

func newDurationParam(name string, f func(c *config.Config) *time.Duration) Param {
	return Param{
		Name: name,
		Type: "string",
		set: func(c *config.Config, s string) error {
			v, err := time.ParseDuration(s)
			if err != nil {
				return err
			}
			p := f(c)
			*p = v
			return nil
		},
		get: func(c *config.Config) string {
			return f(c).String()
		},
	}
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(strings.Replace(s, ",", ".", -1), 64)
}

func getCurrentPartyValues() (map[string]float64, error) {
	var xs []struct {
		Key   string  `db:"key"`
		Value float64 `db:"value"`
	}
	const q1 = `SELECT key, value FROM party_value WHERE party_id = (SELECT party_id FROM app_config)`
	if err := data.DB.Select(&xs, q1); err != nil {
		return nil, merry.Append(err, q1)
	}
	m := map[string]float64{}
	for _, x := range xs {
		m[x.Key] = x.Value
	}
	return m, nil
}
