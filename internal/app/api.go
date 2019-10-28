package app

import (
	"context"
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

func (h *productsServiceHandler) GetParty(ctx context.Context, partyID int64) (*apitypes.Party, error) {
	var p data.Party
	if err := h.db.GetContext(ctx, &p, `SELECT * FROM party WHERE party_id=?`, partyID); err == nil {
		return nil, err
	}
	party := &apitypes.Party{
		PartyID:   p.PartyID,
		CreatedAt: timeUnixMillis(p.CreatedAt),
	}

	xs, err := data.ListProducts(ctx, h.db, p.PartyID)
	if err != nil {
		return nil, err
	}

	for _, p := range xs {
		party.Products = append(party.Products, &apitypes.Product{
			ProductID: p.ProductID,
			PartyID:   p.PartyID,
			Port:      int32(p.Port),
			Addr:      int8(p.Addr),
			Serial:    int32(p.Serial),
			Checked:   p.Checked,
		})
	}
	return party,nil
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
	SET serial=?, addr=?, port=?, checked=?
WHERE product_id = ?`, p.Serial, p.Addr, p.Port, p.Checked, p.ProductID)
	return err
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
