package app

import (
	"context"
	"database/sql"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/gui"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/atool/internal/pkg/winapi"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/comm/modbus"
	"github.com/lxn/win"
	"gopkg.in/yaml.v3"
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
	if connected() {
		return merry.New("нельзя менять конфигурацию пока выполняется опрос")
	}

	filename := filepath.Join(tmpDir, "app-config.yaml")
	c, err := data.OpenAppConfig(db, ctx)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filename, must.MarshalYaml(&c), 0644); err != nil {
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
		var c data.AppConfig
		if err := yaml.Unmarshal(b, &c); err != nil {
			return err
		}
		return data.SaveAppConfig(db, ctx, c)
	}

	go func() {

		if connected() {
			go gui.MsgBox("Ошибка при сохранении конфигурации",
				"Нельзя сменить активную партию пока выполняется опрос",
				win.MB_OK|win.MB_ICONERROR)
			return
		}

		if err := applyConfig(); err != nil {
			log.PrintErr(err)
			gui.MsgBox("Ошибка при сохранении конфигурации", err.Error(), win.MB_OK|win.MB_ICONERROR)
		}
		go gui.NotifyCurrentPartyChanged()
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

func (h *productsServiceHandler) CreateNewParty(ctx context.Context, productsCount int8) error {
	if connected() {
		return merry.New("нельзя создать новую партию пока выполняется опрос")
	}
	return data.CreateNewParty(ctx, db, int(productsCount))
}

func (h *productsServiceHandler) GetCurrentParty(ctx context.Context) (r *apitypes.Party, err error) {
	partyID, err := data.GetCurrentPartyID(db)
	if err != nil {
		return nil, err
	}
	return h.GetParty(ctx, partyID)
}

func (h *productsServiceHandler) RequestCurrentPartyChart(_ context.Context) error {
	xs, err := data.GetCurrentPartyChart(db)
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
		PartyID:        dataParty.PartyID,
		CreatedAt:      timeUnixMillis(dataParty.CreatedAt),
		Products:       []*apitypes.Product{},
		ParamAddresses: []int16{},
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
		})
	}
	for _, p := range dataParty.ParamAddresses {
		party.ParamAddresses = append(party.ParamAddresses, int16(p))
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
	err = db.SelectContext(ctx, &xs, `SELECT device FROM hardware`)
	return
}

//const timeLayout = "2006-01-02 15:04:05.000"

func timeUnixMillis(t time.Time) apitypes.TimeUnixMillis {

	return apitypes.TimeUnixMillis(t.UnixNano() / int64(time.Millisecond))
}

//func unixMillisToTime(m apitypes.TimeUnixMillis) time.Time {
//	t := int64(time.Millisecond) * int64(m)
//	sec := t / int64(time.Second)
//	ns := t % int64(time.Second)
//	return time.Unix(sec, ns)
//}
