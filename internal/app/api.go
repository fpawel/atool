package app

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/gui"
	"github.com/fpawel/atool/internal/cfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/atool/internal/pkg/winapi"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/comm/modbus"
	"github.com/lxn/win"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type productsServiceHandler struct {
}

var _ api.ProductsService = new(productsServiceHandler)

func (h *productsServiceHandler) Connect(_ context.Context) error {
	if connected() {
		return nil
	}

	connect()
	return nil
}

func (h *productsServiceHandler) Disconnect(_ context.Context) error {
	if !connected() {
		return nil
	}
	disconnect()
	wgConnect.Wait()
	return nil
}

func (h *productsServiceHandler) Connected(_ context.Context) (bool, error) {
	return atomic.LoadInt32(&atomicConnected) != 0, nil
}

func (h *productsServiceHandler) OpenGuiClient(_ context.Context, hWnd int64) error {
	gui.SetHWndTargetSendMessage(win.HWND(hWnd))
	return nil
}

func (h *productsServiceHandler) CloseGuiClient(_ context.Context) error {
	gui.SetHWndTargetSendMessage(win.HWND_TOP)
	return nil
}

func (h *productsServiceHandler) ListParamAddresses(_ context.Context) (r []int32, _ error) {
	for _, n := range cfg.Get().Hardware.ParamAddresses() {
		r = append(r, int32(n))
	}
	return
}

func (h *productsServiceHandler) SetProductActive(_ context.Context, productID int64, active bool) error {
	_, err := db.Exec(`UPDATE product SET active = ? WHERE product_id=?`, active, productID)
	return err
}

func (h *productsServiceHandler) GetProductParam(ctx context.Context, productID int64, paramAddr int16) (*apitypes.ProductParam, error) {
	var d data.ProductParam
	err := db.Get(&d, `SELECT * FROM product_param WHERE product_id=? AND param_addr=?`, productID, paramAddr)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return &apitypes.ProductParam{
		ProductID:    productID,
		ParamAddr:    paramAddr,
		Chart:        d.Chart,
		SeriesActive: d.SeriesActive,
	}, nil
}

func (h *productsServiceHandler) SetProductParam(_ context.Context, p *apitypes.ProductParam) error {
	if p.Chart == "" {
		_, err := db.Exec(`DELETE FROM product_param WHERE product_id = ? AND param_addr = ?`, p.ProductID, p.ParamAddr)
		return err
	}
	return data.SetProductParam(db, data.ProductParam{
		ProductID:    p.ProductID,
		ParamAddr:    modbus.Var(p.ParamAddr),
		Chart:        p.Chart,
		SeriesActive: p.SeriesActive,
	})
}

func (h *productsServiceHandler) EditConfig(ctx context.Context) error {

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

func (h *productsServiceHandler) SetCurrentParty(ctx context.Context, partyID int64) error {
	if connected() {
		return merry.New("нельзя сменить активную партию пока выполняется опрос")
	}
	_, err := db.ExecContext(ctx, `UPDATE app_config SET party_id=? WHERE id=1`, partyID)
	return err
}

func (h *productsServiceHandler) CreateNewParty(ctx context.Context, productsCount int8, name string) error {
	if connected() {
		return merry.New("нельзя создать новую партию пока выполняется опрос")
	}
	return data.CreateNewParty(ctx, db, int(productsCount), name)
}

func (h *productsServiceHandler) GetCurrentParty(ctx context.Context) (r *apitypes.Party, err error) {
	partyID, err := data.GetCurrentPartyID(db)
	if err != nil {
		return nil, err
	}
	return h.GetParty(ctx, partyID)
}

func (h *productsServiceHandler) RequestCurrentPartyChart(_ context.Context) error {
	xs, err := data.GetCurrentPartyChart(db, cfg.Get().Hardware.ParamAddresses())
	if err != nil {
		return err
	}
	go gui.NotifyChart(xs)
	return nil
}

func (h *productsServiceHandler) ListParties(ctx context.Context) (parties []*apitypes.PartyInfo, err error) {
	var xs []data.PartyInfo
	if err = db.SelectContext(ctx, &xs, `SELECT * FROM party ORDER BY created_at`); err != nil {
		return
	}
	for _, x := range xs {
		parties = append(parties, &apitypes.PartyInfo{
			PartyID:   x.PartyID,
			CreatedAt: timeUnixMillis(x.CreatedAt),
			Name:      x.Name,
		})
	}
	return
}

func (h *productsServiceHandler) GetParty(_ context.Context, partyID int64) (*apitypes.Party, error) {
	dataParty, err := data.GetParty(db, partyID)
	if err != nil {
		return nil, err
	}
	party := &apitypes.Party{
		PartyID:   dataParty.PartyID,
		CreatedAt: timeUnixMillis(dataParty.CreatedAt),
		Name:      dataParty.Name,
		Products:  []*apitypes.Product{},
	}

	for _, p := range dataParty.Products {
		party.Products = append(party.Products, &apitypes.Product{
			ProductID:      p.ProductID,
			PartyID:        p.PartyID,
			PartyCreatedAt: timeUnixMillis(p.PartyCreatedAt),
			Comport:        p.Comport,
			Addr:           int8(p.Addr),
			Device:         p.Device,
			Active:         p.Active,
			Serial:         int64(p.Serial),
		})
	}
	return party, nil
}

func (h *productsServiceHandler) AddNewProducts(_ context.Context, productsCount int8) error {
	for i := productsCount; i > 0; i-- {
		if err := data.AddNewProduct(db); err != nil {
			return err
		}
	}
	return nil

}

func (h *productsServiceHandler) SetProductsComport(ctx context.Context, productIDs []int64, comport string) error {
	_, err := db.ExecContext(ctx, `UPDATE product SET comport = ? WHERE product_id IN (`+formatIDs(productIDs)+")", comport)
	return err
}

func (h *productsServiceHandler) SetProductsDevice(ctx context.Context, productIDs []int64, device string) error {
	_, err := db.ExecContext(ctx, `UPDATE product SET device = ? WHERE product_id IN (`+formatIDs(productIDs)+")", device)
	return err
}

func formatIDs(ids []int64) string {
	var ss []string
	for _, id := range ids {
		ss = append(ss, strconv.FormatInt(id, 10))
	}
	return strings.Join(ss, ",")
}

func (h *productsServiceHandler) DeleteProducts(ctx context.Context, productIDs []int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM product WHERE product_id IN (`+formatIDs(productIDs)+")")
	return err
}

func (h *productsServiceHandler) SetProductAddr(_ context.Context, productID int64, addr int16) error {
	_, err := db.Exec(`UPDATE product SET addr=? WHERE product_id = ?`, addr, productID)
	return err
}

func (h *productsServiceHandler) ListDevices(ctx context.Context) (xs []string, err error) {
	for _, d := range cfg.Get().Hardware {
		xs = append(xs, d.Name)
	}
	return
}

func (h *productsServiceHandler) SetPartyName(_ context.Context, name string) error {
	_, err := db.Exec(`UPDATE party SET name = ? WHERE party_id = (SELECT party_id FROM app_config)`, name)
	return err
}

func (h *productsServiceHandler) SetProductSerial(_ context.Context, productID int64, serial int64) error {
	_, err := db.Exec(`UPDATE product SET serial = ? WHERE product_id = ?`,
		serial, productID)
	return err
}

func (h *productsServiceHandler) DeleteChartPoints(_ context.Context, r *apitypes.DeleteChartPointsRequest) error {
	var xs []data.ProductParam
	if err := db.Select(&xs, `SELECT * FROM product_param WHERE chart=? AND series_active=TRUE`, r.Chart); err != nil {
		return err
	}
	var qProductsXs, qParamsXs []string
	mProducts := map[int64]struct{}{}
	for _, p := range xs {
		if _, f := mProducts[p.ProductID]; !f {
			mProducts[p.ProductID] = struct{}{}
			qProductsXs = append(qProductsXs, fmt.Sprintf("%d", p.ProductID))
		}
		qParamsXs = append(qParamsXs, fmt.Sprintf("%d", p.ParamAddr))
	}
	qProducts := strings.Join(qProductsXs, ",")
	qParams := strings.Join(qParamsXs, ",")

	timeFrom := unixMillisToTime(r.TimeFrom)
	timeTo := unixMillisToTime(r.TimeTo)

	sQ := fmt.Sprintf(`
DELETE FROM measurement 
WHERE product_id IN (%s) 
  AND param_addr IN (%s) 
  AND tm >= %s
  AND tm <= %s
  AND value >= ?
  AND value <= ?`,
		qProducts, qParams, formatTimeAsQuery(timeFrom), formatTimeAsQuery(timeTo))
	log.Printf("delete chart points %q, products:%s, params:%s, time:%v...%v, value:%v...%v, sql:%s",
		r.Chart, qProducts, qParams, timeFrom, timeTo, r.ValueFrom, r.ValueTo, sQ)
	res, err := db.Exec(sQ, r.ValueFrom, r.ValueTo)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	log.Println(n, "rows deleted")
	return nil

}

func timeUnixMillis(t time.Time) apitypes.TimeUnixMillis {

	return apitypes.TimeUnixMillis(t.UnixNano() / int64(time.Millisecond))
}

func unixMillisToTime(m apitypes.TimeUnixMillis) time.Time {
	t := int64(time.Millisecond) * int64(m)
	sec := t / int64(time.Second)
	ns := t % int64(time.Second)
	return time.Unix(sec, ns)
}
