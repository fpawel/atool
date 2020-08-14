package app

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/atool/internal/workparty"
	"github.com/fpawel/comm/modbus"
)

type productSvc struct{}

var _ api.ProductService = new(productSvc)

func (h *productSvc) SetNetAddr(_ context.Context, productID int64) error {
	return workparty.SetNetAddr(productID, notifyComm)(log, appCtx)
}

func (h *productSvc) SetProductSerial(_ context.Context, productID int64, serial int64) error {
	_, err := data.DB.Exec(`UPDATE product SET serial = ? WHERE product_id = ?`,
		serial, productID)
	return err
}

func (h *productSvc) SetProductsComport(ctx context.Context, productIDs []int64, comport string) error {
	_, err := data.DB.ExecContext(ctx, `UPDATE product SET comport = ? WHERE product_id IN (`+formatIDs(productIDs)+")", comport)
	return err
}

func (h *productSvc) SetProductAddr(_ context.Context, productID int64, addr int16) error {
	_, err := data.DB.Exec(`UPDATE product SET addr=? WHERE product_id = ?`, addr, productID)
	return err
}

func (h *productSvc) SetProductActive(_ context.Context, productID int64, active bool) error {
	_, err := data.DB.Exec(`UPDATE product SET active = ? WHERE product_id=?`, active, productID)
	return err
}

func (h *productSvc) SetProductParamSeries(_ context.Context, p *apitypes.ProductParamSeries) error {
	if p.Chart == "" {
		_, err := data.DB.Exec(`DELETE FROM product_param WHERE product_id = ? AND param_addr = ?`, p.ProductID, p.ParamAddr)
		return err
	}
	return data.SetProductParam(data.ProductParam{
		ProductID:    p.ProductID,
		ParamAddr:    modbus.Var(p.ParamAddr),
		Chart:        p.Chart,
		SeriesActive: p.SeriesActive,
	})
}

func (h *productSvc) GetProductParamSeries(_ context.Context, productID int64, paramAddr int16) (*apitypes.ProductParamSeries, error) {
	var d data.ProductParam
	err := data.DB.Get(&d, `SELECT * FROM product_param WHERE product_id=? AND param_addr=?`, productID, paramAddr)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return &apitypes.ProductParamSeries{
		ProductID:    productID,
		ParamAddr:    paramAddr,
		Chart:        d.Chart,
		SeriesActive: d.SeriesActive,
	}, nil
}

func (h *productSvc) DeleteChartPoints(_ context.Context, r *apitypes.DeleteChartPointsRequest) error {
	qProducts, qParams, err := selectProductParamsChart(r.Chart)
	if err != nil {
		return err
	}

	timeFrom := unixMillisToTime(r.TimeFrom)
	timeTo := unixMillisToTime(r.TimeTo)

	sQ := fmt.Sprintf(`
DELETE FROM measurement 
WHERE product_id IN (%s) 
  AND param_addr IN (%s) 
  AND tm >= %d
  AND tm <= %d
  AND value >= ?
  AND value <= ?`,
		qProducts, qParams, timeFrom.UnixNano(), timeTo.UnixNano())
	log.Printf("delete chart points %q, products:%s, params:%s, time:%v...%v, value:%v...%v, sql:%s",
		r.Chart, qProducts, qParams, timeFrom, timeTo, r.ValueFrom, r.ValueTo, sQ)
	res, err := data.DB.Exec(sQ, r.ValueFrom, r.ValueTo)
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
