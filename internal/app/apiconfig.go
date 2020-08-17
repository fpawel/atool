package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/config/appcfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg/winapi"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/atool/internal/workgui"
	"os/exec"
	"strconv"
	"strings"
)

type appConfigSvc struct{}

var _ api.AppConfigService = new(appConfigSvc)

func (h *appConfigSvc) ListDevices(_ context.Context) ([]string, error) {
	return appcfg.Cfg.Hardware.DeviceNames(), nil
}

func (*appConfigSvc) CurrentDeviceInfo(context.Context) (*apitypes.DeviceInfo, error) {
	party, err := data.GetCurrentParty()
	if err != nil {
		return nil, err
	}
	device := appcfg.GetDeviceByNameOrDefault(party.DeviceType)

	r := &apitypes.DeviceInfo{
		ProductTypes: device.ProductTypes,
		Commands:     []string{},
		Coefficients: []*apitypes.Coefficient{},
	}
	for _, i := range device.Config.ListCoefficients() {
		_, inactive := appcfg.Sets.InactiveCoefficients[i]

		kef := &apitypes.Coefficient{
			N:      int32(i),
			Active: !inactive,
			Name:   fmt.Sprintf("%d", i),
		}
		if device.Config.CfsNames != nil {
			name, fName := device.Config.CfsNames[i]
			if fName {
				kef.Name = fmt.Sprintf("%d %s", i, name)
			}
		}
		r.Coefficients = append(r.Coefficients, kef)
	}

	for _, s := range device.Config.Commands {
		r.Commands = append(r.Commands, s)
	}
	return r, nil
}

func (h *appConfigSvc) EditConfig(_ context.Context) error {
	if err := appcfg.Cfg.Save(); err != nil {
		return err
	}

	filename := config.Filename()
	cmd := exec.Command("./npp/notepad++.exe", filename)
	if err := cmd.Start(); err != nil {
		return err
	}
	winapi.ActivateWindowByPid(cmd.Process.Pid)

	applyConfig := func() error {
		if err := cmd.Wait(); err != nil {
			return err
		}
		if workgui.IsConnected() {
			return errors.New("нельзя менять конфигурации при выполнении настройки")
		}
		return appcfg.Reload()
	}

	go func() {
		if err := applyConfig(); err != nil {
			workgui.NotifyErr(log, merry.Prepend(err, "не удалось сохранить конфигурацию"))
			return
		}
		gui.NotifyCurrentPartyChanged()
	}()
	return nil
}

func (h *appConfigSvc) GetParamValues(_ context.Context) ([]*apitypes.ConfigParamValue, error) {
	return appcfg.GetParamsValues()
}

func (h *appConfigSvc) GetParamValue(_ context.Context, key string) (string, error) {
	return appcfg.GetParamValue(key)
}

func (h *appConfigSvc) SetParamValue(_ context.Context, key string, value string) error {
	wrapErr := func(err error) error {
		return merry.Appendf(err, "%q = %q", key, value)
	}

	if workgui.IsConnected() {
		return wrapErr(merry.New("нельзя менять конфигурации при выполнении настройки"))
	}

	if v, f := appcfg.Params[key]; f {
		if err := v.Set(value); err != nil {
			return wrapErr(err)
		}
		return nil
	}

	switch key {

	case "device_type":

		var productType string
		d, _ := appcfg.DeviceTypes[value]
		if len(d.ProductTypes) > 0 {
			productType = d.ProductTypes[0]
		}

		_, err := data.DB.Exec(`UPDATE party SET device_type = ?, product_type = ? WHERE party_id = (SELECT party_id FROM app_config)`, value, productType)
		if err != nil {
			return wrapErr(err)
		}

		d, f := appcfg.DeviceTypes[value]
		if f && d.InitParty != nil {
			if err := d.InitParty(); err != nil {
				return wrapErr(err)
			}
		}

		go gui.NotifyCurrentPartyChanged()
		return nil

	case "product_type":
		_, err := data.DB.Exec(`UPDATE party SET product_type = ? WHERE party_id = (SELECT party_id FROM app_config)`, value)
		if err != nil {
			return wrapErr(err)
		}

		party, err := data.GetCurrentPartyInfo()
		if err != nil {
			return wrapErr(err)
		}

		d, f := appcfg.DeviceTypes[party.DeviceType]
		if f && d.InitParty != nil {
			if err := d.InitParty(); err != nil {
				return wrapErr(err)
			}
		}

		go gui.NotifyCurrentPartyChanged()
		return nil

	case "name":
		_, err := data.DB.Exec(`UPDATE party SET name = ? WHERE party_id = (SELECT party_id FROM app_config)`, value)
		return wrapErr(err)

	default:
		value, err := strconv.ParseFloat(strings.ReplaceAll(value, ",", "."), 64)
		if err != nil {
			return wrapErr(err)
		}
		return wrapErr(data.SetCurrentPartyValue(key, value))
	}
}
