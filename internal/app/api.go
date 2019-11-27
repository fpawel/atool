package app

import (
	"context"
	"github.com/fpawel/atool/gui"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/atool/internal/pkg/winapi"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/comm/modbus"
	"github.com/jmoiron/sqlx"
	"github.com/lxn/win"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type productsServiceHandler struct {
	db *sqlx.DB
}

var _ api.ProductsService = new(productsServiceHandler)

func (h *productsServiceHandler) SetClientWindow(ctx context.Context, hWnd int64) error {
	gui.SetHWndTargetSendMessage(win.HWND(hWnd))
	return nil
}

func (h *productsServiceHandler) SetProductParam(ctx context.Context, p *apitypes.ProductParam) (err error) {
	_, err = h.db.NamedExecContext(ctx, `
INSERT INTO product_param (product_id, param_addr, chart, series_active) 
VALUES (:product_id, :param_addr, :chart, :series_active)
ON CONFLICT DO UPDATE SET series_active=:series_active, chart=:chart 
WHERE product_id=:product_id AND param_addr=:param_addr`, data.ProductParam{
		ProductID:    p.ProductID,
		ParamAddr:    modbus.Var(p.ParamAddr),
		Chart:        p.Chart,
		SeriesActive: p.SeriesActive,
	})
	return
}

func (h *productsServiceHandler) EditConfig(ctx context.Context) error {
	filename := filepath.Join(tmpDir, "app-config.yaml")
	c, err := data.OpenAppConfig(h.db, ctx)
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
		return data.SaveAppConfig(h.db, ctx, c)
	}

	go func() {

		if err := applyConfig(); err != nil {
			log.PrintErr(err)
			gui.MsgBox("Ошибка при сохранении конфигурации", err.Error(), win.MB_OK|win.MB_ICONERROR)
		}
		gui.NotifyCurrentPartyChanged()
	}()
	return nil
}

func (h *productsServiceHandler) SetCurrentParty(ctx context.Context, partyID int64) error {
	_, err := h.db.ExecContext(ctx, `UPDATE app_config SET party_id=? WHERE id=1`, partyID)
	return err
}

func (h *productsServiceHandler) CreateNewParty(ctx context.Context, productsCount int8) error {
	return data.CreateNewParty(ctx, h.db, int(productsCount))
}

func (h *productsServiceHandler) GetCurrentParty(ctx context.Context) (r *apitypes.Party, err error) {
	partyID, err := data.GetCurrentPartyID(ctx, h.db)
	if err != nil {
		return nil, err
	}
	return h.GetParty(ctx, partyID)
}

func (h *productsServiceHandler) ListParties(ctx context.Context) (parties []*apitypes.PartyInfo, err error) {
	var xs []data.PartyInfo
	if err = h.db.SelectContext(ctx, &xs, `SELECT * FROM party ORDER BY created_at`); err != nil {
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

func (h *productsServiceHandler) GetParty(ctx context.Context, partyID int64) (*apitypes.Party, error) {
	dataParty, err := data.GetParty(ctx, h.db, partyID)
	if err != nil {
		return nil, err
	}
	party := &apitypes.Party{
		PartyID:        dataParty.PartyID,
		CreatedAt:      timeUnixMillis(dataParty.CreatedAt),
		Products:       []*apitypes.Product{},
		ProductParams:  []*apitypes.ProductParam{},
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
	for _, p := range dataParty.ProductParams {
		party.ProductParams = append(party.ProductParams, &apitypes.ProductParam{
			ProductID:    p.ProductID,
			ParamAddr:    int16(p.ParamAddr),
			Chart:        p.Chart,
			SeriesActive: p.SeriesActive,
		})
	}
	return party, nil
}

func (h *productsServiceHandler) AddNewProducts(ctx context.Context, productsCount int8) error {
	for i := productsCount; i > 0; i-- {
		if err := data.AddNewProduct(ctx, h.db); err != nil {
			return err
		}
	}
	return nil

}

func (h *productsServiceHandler) SetProductsComport(ctx context.Context, productIDs []int64, comport string) error {
	_, err := h.db.ExecContext(ctx, `UPDATE product SET comport = ? WHERE product_id IN (`+formatIDs(productIDs)+")", comport)
	return err
}

func (h *productsServiceHandler) SetProductsDevice(ctx context.Context, productIDs []int64, device string) error {
	_, err := h.db.ExecContext(ctx, `UPDATE product SET device = ? WHERE product_id IN (`+formatIDs(productIDs)+")", device)
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
	_, err := h.db.ExecContext(ctx, `DELETE FROM product WHERE product_id IN (`+formatIDs(productIDs)+")")
	return err
}

func (h *productsServiceHandler) SetProduct(ctx context.Context, p *apitypes.Product) error {
	_, err := h.db.NamedExecContext(ctx, `
UPDATE product 
	SET addr=:addr, comport=:comport, device=:device, active=:active 
WHERE product_id = ?`, p)
	return err
}

func (h *productsServiceHandler) ListDevices(ctx context.Context) (xs []string, err error) {
	err = h.db.SelectContext(ctx, &xs, `SELECT device FROM hardware`)
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
