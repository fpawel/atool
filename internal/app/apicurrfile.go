package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config/appcfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/atool/internal/pkg/winapi"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/atool/internal/workgui"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"time"
)

type currentFileSvc struct{}

var _ api.CurrentFileService = new(currentFileSvc)

func (h *currentFileSvc) RequestChart(_ context.Context) error {
	go func() {
		processCurrentPartyChart()
		debug.FreeOSMemory()
	}()
	return nil
}

func (h *currentFileSvc) RenameChart(_ context.Context, oldName, newName string) error {
	_, err := data.DB.Exec(`
UPDATE product_param
SET chart = ?
WHERE chart = ?
  AND product_id IN (SELECT product_id FROM product WHERE party_id = (SELECT party_id FROM app_config))
`, newName, oldName)
	return err
}

func (h *currentFileSvc) AddNewProducts(_ context.Context, productsCount int8) error {
	for i := 0; i < int(productsCount); i++ {
		if _, err := data.AddNewProduct(i); err != nil {
			return err
		}
	}
	return nil
}

func (h *currentFileSvc) DeleteProducts(_ context.Context, productIDs []int64) error {
	_, err := data.DB.Exec(`DELETE FROM product WHERE product_id IN (` + formatIDs(productIDs) + ")")
	return err
}

func (h *currentFileSvc) ListDeviceParams(_ context.Context) ([]*apitypes.DeviceParam, error) {

	party, err := data.GetCurrentParty()
	if err != nil {
		return nil, err
	}

	device, _ := appcfg.Cfg.Hardware.GetDevice(party.DeviceType)

	r := make([]*apitypes.DeviceParam, 0)
	for _, x := range device.ParamAddresses() {
		r = append(r, &apitypes.DeviceParam{
			ParamAddr: int32(x),
			Name:      device.ParamName(x),
		})
	}
	return r, nil
}

func (h *currentFileSvc) RunEdit(_ context.Context) error {

	partyID, err := data.GetCurrentPartyID()
	if err != nil {
		return err
	}
	var partyValues data.PartyValues
	if err := data.GetPartyValues(partyID, &partyValues, -1); err != nil {
		return err
	}

	filename := filepath.Join(tmpDir, "party.yaml")
	if err := ioutil.WriteFile(filename, must.MarshalYaml(partyValues), 0644); err != nil {
		return err
	}
	cmd := exec.Command("./npp/notepad++.exe", filename)
	if err := cmd.Start(); err != nil {
		return err
	}
	winapi.ActivateWindowByPid(cmd.Process.Pid)

	save := func() error {
		if err := cmd.Wait(); err != nil {
			return err
		}
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(b, &partyValues); err != nil {
			return err
		}
		if err := data.SetCurrentPartyValues(partyValues); err != nil {
			return err
		}
		return nil
	}

	go func() {

		if err := save(); err != nil {
			workgui.NotifyErr(log, merry.Append(err, "Ошибка при сохранении данных"))
			return
		}
	}()
	return nil
}

func (h *currentFileSvc) OpenFile(_ context.Context, filename string) error {
	err := data.LoadFile(filename)
	if err == nil {
		go gui.NotifyCurrentPartyChanged()
	}
	return err
}

func processCurrentPartyChart() {

	t := time.Now()

	party, err := data.GetCurrentParty()
	if err != nil {
		err = merry.Append(err, "не удалось получить номер текущего файла")
		log.PrintErr(err)
		workgui.NotifyErr(log, err)
		return
	}

	cfg := appcfg.Cfg.Hardware

	paramsAddresses := cfg.GetDeviceParamAddresses(party.DeviceType)

	log := pkg.LogPrependSuffixKeys(log, "party",
		party.PartyID, "params", fmt.Sprintf("%d", paramsAddresses))

	printErr := func(err error) {
		workgui.NotifyWarnError(log, merry.Appendf(err, "график текущего файла %d: % d, %v",
			party.PartyID, paramsAddresses, time.Since(t)))
	}

	xs, err := data.GetPartyChart(party.PartyID, paramsAddresses)

	log = pkg.LogPrependSuffixKeys(log, "duration", time.Since(t))

	if err != nil {
		printErr(err)
		return
	}
	log.Debug("open chart", "measurements_count", len(xs), "duration", time.Since(t))
	t2 := time.Now()
	gui.NotifyChart(xs)
	gui.Popupf("открыт график %d, %d точек, %v", party.PartyID, len(xs), time.Since(t))
	log.Debug("load chart", "measurements_count", len(xs), "duration", time.Since(t2), "total_duration", time.Since(t))

}
