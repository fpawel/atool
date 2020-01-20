package app

import (
	"context"
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

func (h *appConfigSvc) SetConfig(_ context.Context, c *apitypes.AppConfig) (err error) {
	x := config.Get()
	x.Gas.Comport = c.Gas.Comport
	x.Gas.Type = gas.DevType(c.Gas.DeviceType)
	x.Temperature.Comport = c.Temperature.Comport
	x.Temperature.Type = config.TempDevType(c.Temperature.DeviceType)
	if err := x.Validate(); err != nil {
		return err
	}
	return config.Set(x)
}

func (h *appConfigSvc) GetConfig(_ context.Context) (*apitypes.AppConfig, error) {
	c := config.Get()
	return &apitypes.AppConfig{
		Gas: &apitypes.GasDeviceConfig{
			DeviceType: int8(c.Gas.Type),
			Comport:    c.Gas.Comport,
		},
		Temperature: &apitypes.TemperatureDeviceConfig{
			DeviceType: int8(c.Temperature.Type),
			Comport:    c.Temperature.Comport,
		},
	}, nil
}

func (h *appConfigSvc) GetParamValues(_ context.Context) ([]*apitypes.ConfigParamValue, error) {

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

	for k, x := range configParams {
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

func (h *appConfigSvc) GetParamValue(_ context.Context, key string) (r string, err error) {
	if v, f := configParams[key]; f {
		c := config.Get()
		r = v.get(c)
		return
	}
	const q1 = `SELECT value FROM party_value WHERE party_id = (SELECT party_id FROM app_config) AND key = ?`
	err = db.Select(&r, q1, key)
	return
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

func (x configParam) List() []string {
	if x.list == nil {
		return make([]string, 0)
	}
	return x.list()
}

func listComportsNames() []string {
	comports, _ := comport.Ports()
	return comports
}

var configParams = map[string]configParam{
	"config_temperature_comport": {
		Name: "Термокамера: СОМ порт",
		Type: "comport",
		list: listComportsNames,
		set: func(c *config.Config, s string) error {
			c.Temperature.Comport = s
			return nil
		},
		get: func(c config.Config) string {
			return c.Temperature.Comport
		},
	},

	"config_gas_address": {
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
	"config_gas_comport": {
		Name: "Газовый блок: СОМ порт",
		Type: "comport",
		list: listComportsNames,
		set: func(c *config.Config, s string) error {
			c.Gas.Comport = s
			return nil
		},
		get: func(c config.Config) string {
			return c.Gas.Comport
		},
	},
	"config_gas_type": {
		Name: "Газовый блок: тип",
		list: func() []string {
			return []string{gas.Mil82.String(), gas.Lab73CO.String()}
		},
		get: func(c config.Config) string {
			return c.Gas.Type.String()
		},
		set: func(c *config.Config, s string) error {
			switch s {
			case gas.Mil82.String():
				c.Gas.Type = gas.Mil82
			case gas.Lab73CO.String():
				c.Gas.Type = gas.Lab73CO
			default:
				return fmt.Errorf("wrong gas type: %q", s)
			}
			return nil
		},
	},
}
