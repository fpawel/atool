package app

import (
	"context"
	"github.com/fpawel/atool/internal/cfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/jmoiron/sqlx"
	"time"
)

type productsServiceHandler struct {
	db *sqlx.DB
}

var _ api.ProductsService = new(productsServiceHandler)

func (h *productsServiceHandler) CreateNewParty(ctx context.Context) error {
	return data.CreateNewParty(ctx, h.db)
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
			Port:           int32(p.Port),
			Addr:           int8(p.Addr),
			Serial:         int32(p.Serial),
			Checked:        p.Checked,
			Device:         p.Device,
		})
	}
	return party, nil
}

func (h *productsServiceHandler) AddNewProduct(ctx context.Context) error {
	return data.AddNewProduct(ctx, h.db)
}

func (h *productsServiceHandler) DeleteProduct(ctx context.Context, productID int64) error {
	_, err := h.db.ExecContext(ctx, `DELETE FROM product WHERE product_id=?`, productID)
	return err
}

func (h *productsServiceHandler) SetProduct(ctx context.Context, p *apitypes.Product) error {
	_, err := h.db.Exec(`
UPDATE product 
	SET serial=?, addr=?, port=?, checked=?, device=?
WHERE product_id = ?`, p.Serial, p.Addr, p.Port, p.Checked, p.Device, p.ProductID)
	return err
}

func (h *productsServiceHandler) GetAppConfig(ctx context.Context) (string, error) {
	return cfg.GetYaml(), nil
}

func (h *productsServiceHandler) SetAppConfig(ctx context.Context, appConfig string) error {
	return cfg.SetYaml(appConfig)
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
