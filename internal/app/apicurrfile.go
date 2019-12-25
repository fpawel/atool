package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"time"
)

type currentFileSvc struct{}

var _ api.CurrentFileService = new(currentFileSvc)

func (h *currentFileSvc) RequestChart(_ context.Context) error {
	go getCurrentPartyChart()
	return nil
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

func (h *currentFileSvc) SetName(_ context.Context, name string) error {
	_, err := db.Exec(`UPDATE party SET name = ? WHERE party_id = (SELECT party_id FROM app_config)`, name)
	return err
}

func (h *currentFileSvc) AddNewProducts(_ context.Context, productsCount int8) error {
	for i := productsCount; i > 0; i-- {
		if err := data.AddNewProduct(db); err != nil {
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
		go gui.PopupError(true, err)
	}

	paramAddresses, err := getParamAddresses()
	if err != nil {
		err = merry.Append(err, "не удалось получить номера параметров текущего файла")
		log.PrintErr(err)
		go gui.PopupError(true, err)
		return
	}

	t := time.Now()
	log := pkg.LogPrependSuffixKeys(log, "party", partyID, "params", fmt.Sprintf("% d", paramAddresses))

	printErr := func(err error) {
		err = merry.Appendf(err, "график текущего файла %d: % d, %v", partyID, paramAddresses, time.Since(t))
		log.PrintErr(err)
		go gui.PopupError(true, err)
	}

	log.Debug("getting current party chart")

	go gui.Popup(false, fmt.Sprintf("открывается график файла %d", partyID))

	xs, err := data.GetCurrentPartyChart(db, paramAddresses)

	log = pkg.LogPrependSuffixKeys(log, "duration", time.Since(t))

	if err != nil {
		printErr(err)
		return
	}
	log.Debug("current party chart", "measurements_count", len(xs), "duration", time.Since(t))
	go gui.Popup(false, fmt.Sprintf("открыт график текущего файла %d, %d точек, %v",
		partyID, len(xs), time.Since(t)))
	go gui.NotifyChart(xs)
}
