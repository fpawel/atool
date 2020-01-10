package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/journal"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/atool/internal/pkg/winapi"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type currentFileSvc struct{}

var _ api.CurrentFileService = new(currentFileSvc)

func (h *currentFileSvc) RequestChart(_ context.Context) error {
	go getCurrentPartyChart()
	return nil
}

func (h *currentFileSvc) SetParamValues(_ context.Context, xs []*apitypes.PartyParamValue) error {

	var err error
	appendErr := func(e error) {
		if e == nil {
			return
		}
		if err == nil {
			err = e
			return
		}
		err = merry.WithCause(err, e)
	}

	for _, x := range xs {
		x := x
		appendErr := func(err error) {
			appendErr(merry.Appendf(err, "%s: %q = %q", x.Name, x.Key, x.Value))
		}
		switch x.Key {
		case "product_type":
			_, err := db.Exec(`UPDATE party SET product_type = ? WHERE party_id = (SELECT party_id FROM app_config)`, x.Value)
			appendErr(err)
		case "name":
			_, err := db.Exec(`UPDATE party SET name = ? WHERE party_id = (SELECT party_id FROM app_config)`, x.Value)
			appendErr(err)
		default:
			value, err := strconv.ParseFloat(strings.ReplaceAll(x.Value, ",", "."), 64)
			if err != nil {
				appendErr(err)
				continue
			}
			_, err = db.Exec(`
INSERT INTO party_value (party_id, key, value)
  VALUES ((SELECT party_id FROM app_config), ?, ?)
  ON CONFLICT (party_id,key) DO UPDATE SET value = ?`, x.Key, value, value)
			appendErr(err)
		}
	}
	return err
}

func (h *currentFileSvc) GetParamValues(_ context.Context) ([]*apitypes.PartyParamValue, error) {
	p, err := data.GetCurrentParty(db)
	if err != nil {
		return nil, err
	}
	xs := []*apitypes.PartyParamValue{
		{
			Key:   "name",
			Name:  "Имя файла",
			Value: p.Name,
		},
		{
			Key:   "product_type",
			Name:  "Исполнение",
			Value: p.ProductType,
		},
	}
	var dxs []struct {
		Key   string `db:"key"`
		Value string `db:"value"`
	}
	const q1 = `SELECT key, value FROM party_value WHERE party_id = (SELECT party_id FROM app_config)`
	if err := db.Select(&dxs, q1); err != nil {
		return nil, err
	}
	for key, x := range config.Get().PartyParams {
		y := &apitypes.PartyParamValue{
			Key:  key,
			Name: x.Name,
		}
		for _, z := range dxs {
			if z.Key == key {
				y.Value = z.Value
			}
		}
		xs = append(xs, y)
	}
	return xs, nil

}

func (h *currentFileSvc) RenameChart(_ context.Context, oldName, newName string) error {
	_, err := db.Exec(`
UPDATE product_param
SET chart = ?
WHERE chart = ?
  AND product_id IN (SELECT product_id FROM product WHERE party_id = (SELECT party_id FROM app_config))
`, newName, oldName)
	return err
}

func (h *currentFileSvc) AddNewProducts(_ context.Context, productsCount int8) error {
	for i := 0; i < int(productsCount); i++ {
		if err := data.AddNewProduct(db, i); err != nil {
			return err
		}
	}
	return nil
}

func (h *currentFileSvc) DeleteProducts(ctx context.Context, productIDs []int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM product WHERE product_id IN (`+formatIDs(productIDs)+")")
	return err
}

func (h *currentFileSvc) ListDeviceParams(_ context.Context) ([]*apitypes.DeviceParam, error) {
	xs, err := getParamAddresses()
	if err != nil {
		return nil, err
	}
	r := make([]*apitypes.DeviceParam, 0)
	m := config.Get().ParamsNames
	for _, x := range xs {
		name, _ := m[x]
		r = append(r, &apitypes.DeviceParam{
			ParamAddr: int32(x),
			Name:      name,
		})
	}
	return r, nil
}
func (h *currentFileSvc) CreateNewCopy(_ context.Context) error {
	go func() {
		if err := data.CopyCurrentParty(db); err != nil {
			journal.Err(log, merry.Append(err, "копирование текущего файла"))
			return
		}
		gui.NotifyCurrentPartyChanged()
	}()
	return nil

}

func (h *currentFileSvc) RunEdit(_ context.Context) error {

	var partyValues data.PartyValues
	if err := data.GetCurrentPartyValues(db, &partyValues); err != nil {
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
		if err := data.SetCurrentPartyValues(db, partyValues); err != nil {
			return err
		}
		return nil
	}

	go func() {

		if err := save(); err != nil {
			journal.Err(log, merry.Append(err, "Ошибка при сохранении данных"))
			return
		}
	}()
	return nil
}

func getParamAddresses() ([]int, error) {
	var xs []string
	if err := db.Select(&xs, `SELECT DISTINCT device FROM product WHERE party_id IN (SELECT party_id FROM app_config)`); err != nil {
		return nil, err
	}
	m := make(map[string]struct{})
	for _, x := range xs {
		m[x] = struct{}{}
	}
	var r []int
	for _, n := range config.Get().Hardware.ParamAddresses(m) {
		r = append(r, n)
	}
	return r, nil
}

func getCurrentPartyChart() {

	partyID, err := data.GetCurrentPartyID(db)
	if err != nil {
		err = merry.Append(err, "не удалось получить номер текущего файла")
		log.PrintErr(err)
		journal.Err(log, err)
	}

	paramAddresses, err := getParamAddresses()
	if err != nil {
		err = merry.Append(err, "не удалось получить номера параметров текущего файла")
		log.PrintErr(err)
		journal.Err(log, err)
		return
	}

	t := time.Now()
	log := pkg.LogPrependSuffixKeys(log, "party", partyID, "params", fmt.Sprintf("%d", paramAddresses))

	printErr := func(err error) {
		journal.WarnError(log, merry.Appendf(err, "график текущего файла %d: % d, %v", partyID, paramAddresses, time.Since(t)))
	}

	gui.Popupf("открывается график файла %d", partyID)

	xs, err := data.GetCurrentPartyChart(db, paramAddresses)

	log = pkg.LogPrependSuffixKeys(log, "duration", time.Since(t))

	if err != nil {
		printErr(err)
		return
	}
	//log.Debug("current party chart", "measurements_count", len(xs), "duration", time.Since(t))
	gui.Popupf("открыт график текущего файла %d, %d точек, %v", partyID, len(xs), time.Since(t))
	go gui.NotifyChart(xs)
}
