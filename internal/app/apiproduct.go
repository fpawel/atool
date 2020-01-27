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
	err := db.Get(&p, `SELECT * FROM product WHERE product_id=?`, productID)
	if err != nil {
		return err
	}

	return guiwork.RunWork(log, appCtx, fmt.Sprintf("прибр %d: запись сетевого адреса %d", p.Serial, p.Addr),
		func(log logger, ctx context.Context) error {
			d, f := config.Get().Hardware[p.Device]
			if !f {
				return fmt.Errorf("не заданы параметры устройства %s для прибора %+v", p.Device, p)
			}
			comPort := comports.GetComport(p.Comport, d.Baud)
			if err := comPort.Open(); err != nil {
				return err
			}

			r := modbus.RequestWrite32{
				Addr:      0,
				ProtoCmd:  0x10,
				DeviceCmd: d.NetAddr.Cmd,
				Format:    d.NetAddr.Format,
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
			}.GetResponse(log, ctx, getCommProduct(p.Comport, d))
			return err
		})
}

func (h *productSvc) SetProductSerial(_ context.Context, productID int64, serial int64) error {
	_, err := db.Exec(`UPDATE product SET serial = ? WHERE product_id = ?`,
		serial, productID)
	return err
}

func (h *productSvc) SetProductsComport(ctx context.Context, productIDs []int64, comport string) error {
	_, err := db.ExecContext(ctx, `UPDATE product SET comport = ? WHERE product_id IN (`+formatIDs(productIDs)+")", comport)
	return err
}

func (h *productSvc) SetProductsDevice(ctx context.Context, productIDs []int64, device string) error {
	_, err := db.ExecContext(ctx, `UPDATE product SET device = ? WHERE product_id IN (`+formatIDs(productIDs)+")", device)
	return err
}

func (h *productSvc) SetProductAddr(_ context.Context, productID int64, addr int16) error {
	_, err := db.Exec(`UPDATE product SET addr=? WHERE product_id = ?`, addr, productID)
	return err
}

func (h *productSvc) SetProductActive(_ context.Context, productID int64, active bool) error {
	_, err := db.Exec(`UPDATE product SET active = ? WHERE product_id=?`, active, productID)
	return err
}

func (h *productSvc) SetProductParamSeries(_ context.Context, p *apitypes.ProductParamSeries) error {
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

func (h *productSvc) GetProductParamSeries(ctx context.Context, productID int64, paramAddr int16) (*apitypes.ProductParamSeries, error) {
	var d data.ProductParam
	err := db.Get(&d, `SELECT * FROM product_param WHERE product_id=? AND param_addr=?`, productID, paramAddr)
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
	if err := db.Select(&xs, `SELECT * FROM product_param WHERE chart=? AND series_active=TRUE`, r.Chart); err != nil {
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
	res, err := db.Exec(sQ, r.ValueFrom, r.ValueTo)
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
