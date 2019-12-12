package app

import (
	"context"
	"github.com/fpawel/atool/gui"
	"github.com/fpawel/atool/internal/cfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/thriftgen/api"
)

type currentFileSvc struct{}

var _ api.CurrentFileService = new(currentFileSvc)

func (h *currentFileSvc) RequestCurrentPartyChart(_ context.Context) error {
	xs, err := data.GetCurrentPartyChart(db, cfg.Get().Hardware.ParamAddresses())
	if err != nil {
		return err
	}
	go gui.NotifyChart(xs)
	return nil
}

func (h *currentFileSvc) SetPartyName(_ context.Context, name string) error {
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
