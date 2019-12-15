package app

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/gui"
	"github.com/fpawel/atool/internal/cfg"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/atool/internal/pkg/winapi"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"io/ioutil"
	"os/exec"
	"path/filepath"
)

type appConfigSvc struct{}

var _ api.AppConfigService = new(appConfigSvc)

func (h *appConfigSvc) ListDevices(_ context.Context) (xs []string, err error) {
	for _, d := range cfg.Get().Hardware {
		xs = append(xs, d.Name)
	}
	return
}

func (h *appConfigSvc) EditConfig(_ context.Context) error {

	filename := filepath.Join(tmpDir, "config.yaml")

	if err := ioutil.WriteFile(filename, must.MarshalYaml(cfg.Get()), 0644); err != nil {
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
		return cfg.SetYaml(b)
	}

	go func() {
		if err := applyConfig(); err != nil {
			log.PrintErr(err)
			go gui.PopupError(merry.Append(err, "Ошибка при сохранении конфигурации"))
			return
		}
		gui.NotifyCurrentPartyChanged()
	}()
	return nil
}

func (h *appConfigSvc) Get(_ context.Context) (*apitypes.AppConfig, error) {
	c := cfg.Get()
	return &apitypes.AppConfig{
		Gas: &apitypes.GasDeviceConfig{
			DeviceType: int8(c.Gas.Type),
			Comport:    c.Gas.Comport,
		},
	}, nil
}
