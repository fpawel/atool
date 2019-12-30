package app

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/atool/internal/pkg/winapi"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/hardware/gas"
	"io/ioutil"
	"os/exec"
	"path/filepath"
)

type appConfigSvc struct{}

var _ api.AppConfigService = new(appConfigSvc)

func (h *appConfigSvc) ListDevices(_ context.Context) (xs []string, err error) {
	for _, d := range config.Get().Hardware {
		xs = append(xs, d.Name)
	}
	return
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
			gui.JournalError(log, merry.Append(err, "Ошибка при сохранении конфигурации"))
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
