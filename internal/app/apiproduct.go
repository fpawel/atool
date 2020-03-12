package app

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/guiwork"
	"github.com/fpawel/atool/internal/pkg/comports"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"strings"
	"time"
)

type productSvc struct{}

var _ api.ProductService = new(productSvc)

func (h *productSvc) SetNetAddr(_ context.Context, productID int64) error {
	var p data.Product
	err := data.DB.Get(&p, `SELECT * FROM product WHERE product_id=?`, productID)
	if err != nil {
		return err
	}

	party, err := data.GetCurrentParty()
	if err != nil {
		return err
	}

	device, f := config.Get().Hardware[party.DeviceType]
	if !f {
		return fmt.Errorf("не заданы параметры устройства %s для прибора %+v", party.DeviceType, p)
	}

	return guiwork.RunWork(log, appCtx, fmt.Sprintf("прибр %d: запись сетевого адреса %d", p.Serial, p.Addr),
		func(log logger, ctx context.Context) error {

			comPort := comports.GetComport(p.Comport, device.Baud)
			if err := comPort.Open(); err != nil {
				return err
			}

			r := modbus.RequestWrite32{
				Addr:      0,
				ProtoCmd:  0x10,
				DeviceCmd: device.NetAddr.Cmd,
				Format:    device.NetAddr.Format,
				Value:     float64(p.Addr),
			}
			if _, err := comPort.Write(r.Request().Bytes()); err != nil {
				return err
			}

			notifyComm(comm.Info{
				Request: r.Request().Bytes(),
				Port:    p.Comport,
			})

			pause(ctx.Done(), time.Second)
			_, err := modbus.RequestRead3{
				Addr:           p.Addr,
				FirstRegister:  0,
				RegistersCount: 2,
			}.GetResponse(log, ctx, getCommProduct(p.Comport, device))
			return err
		})
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
	var xs []data.ProductParam
	if err := data.DB.Select(&xs, `SELECT * FROM product_param WHERE chart=? AND series_active=TRUE`, r.Chart); err != nil {
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
