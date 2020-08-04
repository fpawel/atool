package appcfg

import (
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/data"
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
		return v.Get(), nil
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

	device, _ := DeviceTypes[p.DeviceType]

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
			ValuesList: Cfg.Hardware.DeviceNames(),
		},
		{
			Key:        "product_type",
			Name:       "Приборы: исполнение",
			Value:      p.ProductType,
			ValuesList: device.ProductTypes,
		},
	}

	m, err := getCurrentPartyValues()
	if err != nil {
		return nil, err
	}

	dv, _ := DeviceTypes[p.DeviceType]
	for _, param := range dv.PartyParams {
		y := &apitypes.ConfigParamValue{
			Key:        param.Key,
			Name:       p.DeviceType + ": " + param.Name,
			ValuesList: param.ValuesList,
			Type:       "float",
		}
		if v, f := m[param.Key]; f {
			y.Value = Cfg.FormatFloat(v)
		}
		xs = append(xs, y)
	}

	var params []*apitypes.ConfigParamValue

	for k, x := range Params {
		params = append(params, &apitypes.ConfigParamValue{
			Key:        k,
			Name:       x.Name,
			Type:       x.Type,
			Value:      x.Get(),
			ValuesList: x.GetList(),
		})
	}

	sort.Slice(params, func(i, j int) bool {
		return params[i].Name < params[j].Name
	})
	xs = append(xs, params...)

	for i, x := range xs {
		for j, y := range xs {
			if i != j && x.Key == y.Key {
				return nil, merry.Errorf("дублирование значений ключа %s %d:%s %d:%s", x.Key, i, x.Name, j, y.Name)
			}
		}
	}
	return xs, nil
}

type Param struct {
	Name string
	Type string
	List func() []string
	Set  func(string) error
	Get  func() string
}

func (x Param) GetList() []string {
	if x.Type == "comport" {
		comports, _ := comport.Ports()
		return comports
	}
	if x.List == nil {
		return make([]string, 0)
	}
	return x.List()
}

var (
	Params = map[string]Param{

		"temperature_type": {
			Name: "Термокамера: тип",
			List: func() []string {
				return []string{string(config.T800), string(config.T2500), string(config.Ktx500)}
			},
			Get: func() string {
				return string(Cfg.Temperature.Type)
			},
			Set: func(s string) error {
				Cfg.Temperature.Type = config.TempDevType(s)
				return nil
			},
		},

		"temperature_comport": {
			Name: "Термокамера: СОМ порт",
			Type: "comport",
			Set: func(s string) error {
				Cfg.Temperature.Comport = s
				return nil
			},
			Get: func() string {
				return Cfg.Temperature.Comport
			},
		},

		"temperature_hold_duration": newDurationParam("Термокамера: длительность выдержки",
			func() *time.Duration {
				return &Cfg.Temperature.HoldDuration
			}),

		"gas_address": {
			Name: "Газовый блок: адрес",
			Type: "int",
			Set: func(s string) error {
				n, err := strconv.ParseInt(s, 10, 8)
				if err != nil {
					return err
				}
				Cfg.Gas.Addr = modbus.Addr(n)
				return nil
			},
			Get: func() string {
				return strconv.Itoa(int(Cfg.Gas.Addr))
			},
		},

		"gas_comport": {
			Name: "Газовый блок: СОМ порт",
			Type: "comport",
			Set: func(s string) error {
				Cfg.Gas.Comport = s
				return nil
			},
			Get: func() string {
				return Cfg.Gas.Comport
			},
		},

		"gas_type": {
			Name: "Газовый блок: тип",
			List: func() []string {
				return gasTypes
			},
			Get: func() string {
				return string(Cfg.Gas.Type)
			},
			Set: func(s string) error {
				for _, v := range gasTypes {
					if v == s {
						Cfg.Gas.Type = gas.DevType(s)
						return nil
					}
				}
				return fmt.Errorf("invalid gas type: %q", s)
			},
		},

		"warm_sheets_enable": {
			Name: "Подогрев плат: использовать",
			Type: "bool",
			Get: func() string {
				return strconv.FormatBool(Cfg.WarmSheets.Enable)
			},
			Set: func(s string) error {
				v, err := strconv.ParseBool(s)
				if err != nil {
					return err
				}
				Cfg.WarmSheets.Enable = v
				return nil
			},
		},
		"warm_sheets_address": {
			Name: "Подогрев плат: адрес устройства",
			Type: "int",
			Get: func() string {
				return strconv.Itoa(int(Cfg.WarmSheets.Addr))
			},
			Set: func(s string) error {
				v, err := strconv.ParseInt(s, 10, 8)
				if err != nil {
					return err
				}
				Cfg.WarmSheets.Addr = modbus.Addr(v)
				return nil
			},
		},

		"warm_sheets_temp_on": newFloatParam("Подогрев плат: температура включения",
			func() *float64 {
				return &Cfg.WarmSheets.TempOn
			}),

		"warm_sheets_temp_off": newFloatParam("Подогрев плат: температура выключения",
			func() *float64 {
				return &Cfg.WarmSheets.TempOff
			}),
	}
)

func newFloatParam(name string, f func() *float64) Param {
	return Param{
		Name: name,
		Type: "float",
		Set: func(s string) error {
			v, err := parseFloat(s)
			if err != nil {
				return err
			}
			p := f()
			*p = v
			return nil
		},
		Get: func() string {
			p := f()
			return Cfg.FormatFloat(*p)
		},
	}
}

func newDurationParam(name string, f func() *time.Duration) Param {
	return Param{
		Name: name,
		Type: "string",
		Set: func(s string) error {
			v, err := time.ParseDuration(s)
			if err != nil {
				return err
			}
			p := f()
			*p = v
			return nil
		},
		Get: func() string {
			return f().String()
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

func init() {
	for i := 0; i < 6; i++ {
		i := i
		Params[fmt.Sprintf("gas%d_duration", i+1)] = newDurationParam(fmt.Sprintf("ПГС%d: длительность продувки", i+1),
			func() *time.Duration {
				return &Cfg.Gas.BlowDuration[i]
			})
	}
}

var gasTypes = []string{string(gas.Mil82), string(gas.Lab73CO), string(gas.Ankat)}
