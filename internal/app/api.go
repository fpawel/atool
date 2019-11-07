package app

import (
	"context"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/jmoiron/sqlx"
	"gopkg.in/yaml.v3"
	"strconv"
	"strings"
	"time"
)

type productsServiceHandler struct {
	db *sqlx.DB
}

var _ api.ProductsService = new(productsServiceHandler)

func (h *productsServiceHandler) SetCurrentParty(ctx context.Context, partyID int64) error {
	_, err := h.db.ExecContext(ctx, `UPDATE app_config SET party_id=? WHERE id=1`, partyID)
	return err
}

func (h *productsServiceHandler) CreateNewParty(ctx context.Context, productsCount int8, note string) error {
	return data.CreateNewParty(ctx, h.db, int(productsCount), note)
}

func (h *productsServiceHandler) SetPartyNote(ctx context.Context, partyID int64, note string) error {
	_, err := h.db.ExecContext(ctx, `UPDATE party SET note=? WHERE party_id=?`, note, partyID)
	return err
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
			Note:      x.Note,
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
		Note:      dataParty.Note,
		Products:  []*apitypes.Product{},
		Params:    []int16{},
		Charts:    []int32{},
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
		for _, x := range dataProduct.Series {
			p.Series = append(p.Series, &apitypes.Series{
				TheVar:  int16(x.Var),
				ChartID: int32(x.ChartID),
				Color:   x.Color,
			})
		}
		party.Products = append(party.Products, p)
	}
	for _, p := range dataParty.Params {
		party.Params = append(party.Params, int16(p))
	}
	for _, p := range dataParty.Charts {
		party.Charts = append(party.Charts, int32(p))
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
