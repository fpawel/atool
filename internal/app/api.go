package app

import (
	"context"
	"github.com/fpawel/atool/gui"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/atool/internal/pkg/winapi"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
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

func (h *productsServiceHandler) EditConfig(ctx context.Context) error {
	filename := filepath.Join(tmpDir, "app-config.yaml")
	c, err := openAppConfig(h.db, ctx)
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
		var c appConfig
		if err := yaml.Unmarshal(b, &c); err != nil {
			return err
		}
		if err := c.save(h.db, ctx); err != nil {
			return err
		}
		return nil
	}

	go func() {

		if err := applyConfig(); err != nil {
			log.PrintErr(err)
			gui.MsgBox("Конфигурация", err.Error(), win.MB_OK|win.MB_ICONERROR)
			return
		}
		gui.NotifyCurrentPartyChanged()
	}()
	return nil
}

func (h *productsServiceHandler) SetProductVarSeriesChart(ctx context.Context, productID int64, theVar int16, chartName string) error {
	_, err := h.db.ExecContext(ctx, `
INSERT INTO series(product_id, var, chart) VALUES (?,?,?) 
	ON CONFLICT (product_id,var) DO 
	    UPDATE SET chart = ?`,
		productID, theVar, chartName, chartName)
	return err
}

func (h *productsServiceHandler) SetProductVarSeriesActive(ctx context.Context, productID int64, theVar int16, active bool) error {
	_, err := h.db.ExecContext(ctx, `UPDATE series SET active=? WHERE product_id=? AND var=?`,
		active, productID, theVar)
	return err
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
		PartyID:   dataParty.PartyID,
		CreatedAt: timeUnixMillis(dataParty.CreatedAt),
		Products:  []*apitypes.Product{},
		Params:    []int16{},
		Series:    []*apitypes.ProductVarSeries{},
	}

	for _, dataProduct := range dataParty.Products {
		p := &apitypes.Product{
			ProductID:      dataProduct.ProductID,
			PartyID:        dataProduct.PartyID,
			PartyCreatedAt: timeUnixMillis(dataProduct.PartyCreatedAt),
			Comport:        dataProduct.Comport,
			Addr:           int8(dataProduct.Addr),
			Checked:        dataProduct.Checked,
			Device:         dataProduct.Device,
		}

		party.Products = append(party.Products, p)
	}
	for _, p := range dataParty.Params {
		party.Params = append(party.Params, int16(p))
	}
	for _, p := range dataParty.Series {
		party.Series = append(party.Series, &apitypes.ProductVarSeries{
			ProductID: p.ProductID,
			TheVar:    int16(p.Var),
			Chart:     p.Chart,
			Active:    p.Active,
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
	_, err := h.db.Exec(`
UPDATE product 
	SET addr=?, comport=?, checked=?, device=?
WHERE product_id = ?`, p.Addr, p.Comport, p.Checked, p.Device, p.ProductID)
	return err
}

func (h *productsServiceHandler) GetAppConfig(ctx context.Context) (string, error) {
	c, err := openAppConfig(h.db, ctx)
	if err != nil {
		return "", err
	}
	return string(must.MarshalYaml(&c)), nil
}

func (h *productsServiceHandler) SetAppConfig(ctx context.Context, appConfigYaml string) error {
	var c appConfig
	if err := yaml.Unmarshal([]byte(appConfigYaml), &c); err != nil {
		return err
	}
	return c.save(h.db, ctx)
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
