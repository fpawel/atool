package app

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/journal"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/atool/internal/pkg/winapi"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/hardware/gas"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type appConfigSvc struct{}

var _ api.AppConfigService = new(appConfigSvc)

func (h *appConfigSvc) ListDevices(_ context.Context) ([]string, error) {
	return config.Get().Hardware.ListDevices(), nil
}

func (h *appConfigSvc) EditConfig(_ context.Context) error {

	filename := filepath.Join(tmpDir, "config.yaml")

	if err := ioutil.WriteFile(filename, must.MarshalYaml(config.Get()), 0644); err != nil {
		return err
	}
	cmd := exec.Command("./npp/notepad++.exe", filename)
	if err := cmd.Start(); err != nil {
		return err
	}
	winapi.ActivateWindowByPid(cmd.Process.Pid)

	applyConfig := func() error {
		if err := cmd.Wait(); err != nil {
			return err
		}
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		return config.SetYaml(b)
	}

	go func() {
		if err := applyConfig(); err != nil {
			log.PrintErr(err)
			journal.Err(log, merry.Append(err, "Ошибка при сохранении конфигурации"))
			return
		}
		gui.NotifyCurrentPartyChanged()
	}()
	return nil
}

func (h *appConfigSvc) GetParamValues(_ context.Context) ([]*apitypes.ConfigParamValue, error) {
	return getConfigParamsValues()
}

func (h *appConfigSvc) GetParamValue(_ context.Context, key string) (string, error) {
	return getConfigParamValue(key)
}

func (h *appConfigSvc) SetParamValue(_ context.Context, key string, value string) error {
	wrapErr := func(err error) error {
		return merry.Appendf(err, "%q = %q", key, value)
	}

	if v, f := configParams[key]; f {
		c := config.Get()
		if err := v.set(&c, value); err != nil {
			return wrapErr(err)
		}
		return wrapErr(config.Set(c))
	}

	switch key {

	case "product_type":
		_, err := db.Exec(`UPDATE party SET product_type = ? WHERE party_id = (SELECT party_id FROM app_config)`, value)
		return wrapErr(err)

	case "name":
		_, err := db.Exec(`UPDATE party SET name = ? WHERE party_id = (SELECT party_id FROM app_config)`, value)
		return wrapErr(err)

	default:
		value, err := strconv.ParseFloat(strings.ReplaceAll(value, ",", "."), 64)
		if err != nil {
			return wrapErr(err)
		}
		_, err = db.Exec(`
INSERT INTO party_value (party_id, key, value)
  VALUES ((SELECT party_id FROM app_config), ?, ?)
  ON CONFLICT (party_id,key) DO UPDATE SET value = ?`, key, value, value)
		return wrapErr(err)
	}
}

type configParam struct {
	Name string
	Type string
	list func() []string
	set  func(*config.Config, string) error
	get  func(config.Config) string
}

func getConfigParamValue(key string) (string, error) {
	if v, f := configParams[key]; f {
		c := config.Get()
		return v.get(c), nil
	}
	const q1 = `SELECT value FROM party_value WHERE party_id = (SELECT party_id FROM app_config) AND key = ?`
	var r string
	err := db.Get(&r, q1, key)
	if err == sql.ErrNoRows {
		err = fmt.Errorf("значение ключа партии %q не задано", key)
	}
	return r, err
}

func (x configParam) List() []string {
	if x.Type == "comport" {
		comports, _ := comport.Ports()
		return comports
	}
	if x.list == nil {
		return make([]string, 0)
	}
	return x.list()
}

func getConfigParamsValues() ([]*apitypes.ConfigParamValue, error) {
	p, err := data.GetCurrentParty(db)
	if err != nil {
		return nil, err
	}

	cfg := config.Get()

	xs := []*apitypes.ConfigParamValue{
		{
			Key:   "name",
			Name:  "Имя файла",
			Value: p.Name,
		},
		{
			Key:        "product_type",
			Name:       "Исполнение",
			Value:      p.ProductType,
			ValuesList: cfg.ProductTypes,
		},
	}

	checkKey := func(k string) error {
		for _, x := range xs {
			if x.Key == k {
				return fmt.Errorf("дублирование значений ключа %q", k)
			}
		}
		return nil
	}

	for k, x := range configParams {
		if err := checkKey(k); err != nil {
			return nil, err
		}
		xs = append(xs, &apitypes.ConfigParamValue{
			Key:        k,
			Name:       x.Name,
			Type:       x.Type,
			ValuesList: x.List(),
			Value:      x.get(cfg),
		})
	}

	m, err := getCurrentPartyValues()
	if err != nil {
		return nil, err
	}
	for key, name := range config.Get().PartyParams {
		if err := checkKey(key); err != nil {
			return nil, err
		}
		y := &apitypes.ConfigParamValue{
			Key:  key,
			Name: name,
			Type: "float",
		}
		if v, f := m[key]; f {
			y.Value = formatFloat(v)
		}
		xs = append(xs, y)
	}
	sort.Slice(xs, func(i, j int) bool {
		return xs[i].Name < xs[j].Name
	})
	return xs, nil
}

var configParams = map[string]configParam{

	"temperature_type": {
		Name: "Термокамера: тип",
		list: func() []string {
			return []string{string(config.T800), string(config.T2500), string(config.Ktx500)}
		},
		get: func(c config.Config) string {
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
		get: func(c config.Config) string {
			return c.Temperature.Comport
		},
	},

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
		get: func(c config.Config) string {
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
		get: func(c config.Config) string {
			return c.Gas.Comport
		},
	},
	"gas_type": {
		Name: "Газовый блок: тип",
		list: func() []string {
			return []string{string(gas.Mil82), string(gas.Lab73CO)}
		},
		get: func(c config.Config) string {
			return string(c.Gas.Type)
		},
		set: func(c *config.Config, s string) error {
			c.Gas.Type = gas.DevType(s)
			return nil
		},
	},
}
