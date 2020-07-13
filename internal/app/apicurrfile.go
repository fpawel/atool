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
	"github.com/fpawel/comm"
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
	return runWithNotifyPartyChanged(fmt.Sprintf("добавление приборов в партию: %d", productsCount), func(log comm.Logger, ctx context.Context) error {
		for i := 0; i < int(productsCount); i++ {
			if _, err := data.AddNewProduct(i); err != nil {
				return err
			}
		}
		return nil
	})
}

func (h *currentFileSvc) DeleteProducts(_ context.Context, productIDs []int64) error {
	return runWithNotifyPartyChanged("удаление приборов "+formatIDs(productIDs), func(log comm.Logger, ctx context.Context) error {
		sql := `DELETE FROM product WHERE product_id IN (` + formatIDs(productIDs) + ")"
		_, err := data.DB.Exec(sql)
		if err != nil {
			return merry.Appendf(err, sql)
		}
		return nil
	})
}

func (h *currentFileSvc) ListDeviceParams(_ context.Context) ([]*apitypes.DeviceParam, error) {

	party, err := data.GetCurrentParty()
	if err != nil {
		return nil, err
	}
	device := appcfg.GetDeviceByNameOrDefault(party.DeviceType)

	r := make([]*apitypes.DeviceParam, 0)
	for _, x := range device.Vars(party.ProductType) {
		r = append(r, &apitypes.DeviceParam{
			ParamAddr: int32(x),
			Name:      device.VarName(x),
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

func (*currentFileSvc) ListWorkLogRecords(context.Context) ([]*apitypes.WorkLogRecord, error) {
	var xs []struct {
		StartedAt   time.Time `db:"started_at"`
		CompletedAt time.Time `db:"complete_at"`
		WorkName    string    `db:"work_name"`
	}
	const q = `
SELECT started_at, complete_at, work_name 
FROM work_log 
WHERE party_id = (SELECT app_config.party_id FROM app_config) AND (complete_at IS NOT NULL) 
ORDER BY started_at`
	if err := data.DB.Select(&xs, q); err != nil {
		return nil, err
	}
	r := make([]*apitypes.WorkLogRecord, 0)
	for _, x := range xs {
		r = append(r, &apitypes.WorkLogRecord{
			WorkName:    x.WorkName,
			StrtedAt:    x.StartedAt.Format(time.RFC3339),
			CompletedAt: x.CompletedAt.Format(time.RFC3339),
		})
	}
	return r, nil
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

	device := appcfg.GetDeviceByNameOrDefault(party.DeviceType)

	paramsAddresses := device.Vars(party.ProductType)

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
