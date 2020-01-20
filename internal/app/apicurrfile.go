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
	"runtime/debug"
	"sort"
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

type mapIntFloat map[int64]float64

func (h *currentFileSvc) GetSectionsProductsParamsValues(_ context.Context) ([]*apitypes.SectionProductParamsValues, error) {
	const q2 = `
SELECT product_id, key, value
FROM product_value
WHERE product_id IN (SELECT product_id FROM product WHERE party_id IN (SELECT party_id FROM app_config))`
	var values1 []struct {
		ProductID int64   `db:"product_id"`
		Key       string  `db:"key"`
		Value     float64 `db:"value"`
	}
	if err := db.Select(&values1, q2); err != nil {
		return nil, err
	}

	values := map[string]mapIntFloat{}

	for _, x := range values1 {
		if values[x.Key] == nil {
			values[x.Key] = mapIntFloat{}
		}
		values[x.Key][x.ProductID] = x.Value
	}

	var result []*apitypes.SectionProductParamsValues

	party, err := data.GetCurrentParty(db)
	if err != nil {
		return nil, err
	}

	cfg := config.Get()

	for section, m := range cfg.ProductParams {
		y := &apitypes.SectionProductParamsValues{Section: section, Values: [][]string{{"Прибор"}}}

		for key := range m {
			y.Keys = append(y.Keys, key)
		}
		sort.Slice(y.Keys, func(i, j int) bool {
			return m[y.Keys[i]] < m[y.Keys[j]]
		})

		for _, p := range party.Products {
			y.Values[0] = append(y.Values[0], fmt.Sprintf("№%d ID%d", p.Serial, p.ProductID))
		}
		for _, key := range y.Keys {
			xs := []string{m[key]}
			for _, p := range party.Products {
				var s string
				if m, f := values[key]; f {
					if v, f := m[p.ProductID]; f {
						s = fmt.Sprintf("%v", v)
					}
				}
				xs = append(xs, s)
			}
			y.Values = append(y.Values, xs)
		}
		result = append(result, y)
	}

	addAdditionalProductParamsSectionValues(party, values, &result)

	sort.Slice(result, func(i, j int) bool {
		return result[i].Section > result[j].Section
	})
	return result, nil
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

func addAdditionalProductParamsSectionValues(party data.Party, values map[string]mapIntFloat, result *[]*apitypes.SectionProductParamsValues) {
	sect := &apitypes.SectionProductParamsValues{
		Values:  [][]string{{"Прибор"}},
		Section: fmt.Sprintf("%d. Дополнительно", len(*result)+1),
	}
	cfgProductParamsKeys := config.Get().ListProductParamKeys()

	for _, p := range party.Products {
		sect.Values[0] = append(sect.Values[0], fmt.Sprintf("№%d ID%d", p.Serial, p.ProductID))
	}
	for k, m := range values {
		if _, f := cfgProductParamsKeys[k]; f {
			continue
		}
		xs := []string{k}
		for _, p := range party.Products {
			s := ""
			if v, f := m[p.ProductID]; f {
				s = fmt.Sprintf("%v", v)
			}
			xs = append(xs, s)
		}
		if len(xs) > 1 {
			sect.Values = append(sect.Values, xs)
			sect.Keys = append(sect.Keys, k)
		}
	}

	if len(sect.Values) > 1 {
		sort.Slice(sect.Keys, func(i, j int) bool {
			return sect.Keys[i] < sect.Keys[j]
		})
		vs := sect.Values[1:]
		sort.Slice(vs, func(i, j int) bool {
			return vs[i][0] < vs[j][0]
		})
		*result = append(*result, sect)
	}
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

func processCurrentPartyChart() {

	t := time.Now()

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

	log := pkg.LogPrependSuffixKeys(log, "party", partyID, "params", fmt.Sprintf("%d", paramAddresses))

	printErr := func(err error) {
		journal.WarnError(log, merry.Appendf(err, "график текущего файла %d: % d, %v", partyID, paramAddresses, time.Since(t)))
	}

	gui.Popupf("открывается график файла %d", partyID)

	xs, err := data.GetPartyChart(db, partyID, paramAddresses)

	log = pkg.LogPrependSuffixKeys(log, "duration", time.Since(t))

	if err != nil {
		printErr(err)
		return
	}
	log.Debug("open chart", "measurements_count", len(xs), "duration", time.Since(t))
	t2 := time.Now()
	gui.NotifyChart(xs)
	gui.Popupf("открыт график текущего файла %d, %d точек, %v", partyID, len(xs), time.Since(t))
	log.Debug("load chart", "measurements_count", len(xs), "duration", time.Since(t2), "total_duration", time.Since(t))

}
