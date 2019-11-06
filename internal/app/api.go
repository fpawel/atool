package app

import (
	"context"
	"github.com/fpawel/atool/internal/cfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/jmoiron/sqlx"
	"strconv"
	"strings"
	"time"
)

type productsServiceHandler struct {
	db *sqlx.DB
}

var _ api.ProductsService = new(productsServiceHandler)

func (h *productsServiceHandler) CreateNewParty(ctx context.Context, productsCount int8) error {
	return data.CreateNewParty(ctx, h.db, int(productsCount))
}

func (h *productsServiceHandler) GetLastParty(ctx context.Context) (r *apitypes.Party, err error) {
	partyID, err := data.GetLastPartyID(ctx, h.db)
	if err != nil {
		return nil, err
	}
	return h.GetParty(ctx, partyID)
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
	}

	for _, p := range dataParty.Products {
		party.Products = append(party.Products, &apitypes.Product{
			ProductID:      p.ProductID,
			PartyID:        p.PartyID,
			CreatedAt:      timeUnixMillis(p.CreatedAt),
			PartyCreatedAt: timeUnixMillis(p.PartyCreatedAt),
			Comport:        p.Comport,
			Addr:           int8(p.Addr),
			Checked:        p.Checked,
			Device:         p.Device,
		})
	}
	var params []struct {
		Var    int    `db:"var"`
		Format string `db:"format"`
	}
	if err := h.db.SelectContext(ctx, &params, `
SELECT DISTINCT var, format FROM product INNER JOIN param USING (device)
WHERE party_id IN (SELECT party_id FROM last_party)`); err != nil {
		return nil, err
	}
	for _, p := range params {
		party.Params = append(party.Params, &apitypes.Param{
			TheVar: int16(p.Var),
			Format: p.Format,
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
	var c appConfig
	if err := h.db.GetContext(ctx, &c, `SELECT * FROM app_config`); err != nil {
		return "", err
	}

	if err := h.db.SelectContext(ctx, nil, `
SELECT device, baud, pause,  
       timeout_get_responses,  
       timeout_end_response,
       var, count, format 
FROM param INNER JOIN hardware USING (device)`); err != nil {
		return "", err
	}

	return cfg.GetYaml(), nil
}

func (h *productsServiceHandler) SetAppConfig(ctx context.Context, appConfig string) error {
	return cfg.SetYaml(appConfig)
}

func (h *productsServiceHandler) ListDevices(ctx context.Context) (xs []string, err error) {
	err = h.db.SelectContext(ctx, &xs, `SELECT device FROM hardware`)
	return
}

func (h *productsServiceHandler) ListYearMonths(ctx context.Context) ([]*apitypes.YearMonth, error) {
	var xs []*apitypes.YearMonth
	if err := h.db.Select(&xs, `
SELECT DISTINCT year,  month
FROM measurement_ext
ORDER BY year DESC, month DESC`); err != nil {
		return nil, err
	}
	if len(xs) == 0 {
		t := time.Now()
		xs = append(xs, &apitypes.YearMonth{
			Year:  int32(t.Year()),
			Month: int32(t.Month()),
		})
	}
	return xs, nil
}

const timeLayout = "2006-01-02 15:04:05.000"

func timeUnixMillis(t time.Time) apitypes.TimeUnixMillis {

	return apitypes.TimeUnixMillis(t.UnixNano() / int64(time.Millisecond))
}

func unixMillisToTime(m apitypes.TimeUnixMillis) time.Time {
	t := int64(time.Millisecond) * int64(m)
	sec := t / int64(time.Second)
	ns := t % int64(time.Second)
	return time.Unix(sec, ns)
}
