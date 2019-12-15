package app

import (
	"context"
	"database/sql"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/gui"
	"github.com/fpawel/atool/internal/cfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/hardware/gas"
	"time"
)

func runInterrogate() {
	runWork(func(ctx context.Context) error {
		must.PanicIf(createNewChartIfUpdatedTooLong())
		ms := new(measurements)
		defer func() {
			saveMeasurements(ms.xs)
		}()
		for {
			if err := processProductsParams(ctx, ms); err != nil {
				if merry.Is(err, context.Canceled) {
					return nil
				}
				return err
			}
		}
	})
}

func processProductsParams(ctx context.Context, ms *measurements) error {
	return processEachActiveProduct(func(product data.Product, device cfg.Device) error {
		rdr, err := newParamsReader(ctx, product, device)
		if err != nil {
			return err
		}
		for _, prm := range device.Params {
			if err := rdr.read(prm); err != nil {
				return err
			}
		}
		for _, p := range device.Params {
			for i := 0; i < p.Count; i++ {
				rdr.processParamValueRead(p, i, ms)
			}
		}
		return nil
	})
}

func runRawCommand(c modbus.ProtoCmd, b []byte) {
	runWork(func(ctx context.Context) error {
		return processEachActiveProduct(func(p data.Product, d cfg.Device) error {
			startTime := time.Now()
			rdr, err := getResponseReader(ctx, p.Comport, d)
			if err != nil {
				return err
			}
			req := modbus.Request{
				Addr:     p.Addr,
				ProtoCmd: c,
				Data:     b,
			}
			response, err := rdr.GetResponse(req.Bytes(), log, nil)
			if merry.Is(err, context.Canceled) {
				return err
			}
			ct := gui.CommTransaction{
				Addr:     p.Addr,
				Comport:  p.Comport,
				Request:  formatBytes(req.Bytes()),
				Response: formatBytes(response),
				Duration: time.Since(startTime).String(),
				Ok:       err == nil,
			}
			if err != nil {
				ct.Response = err.Error()
			}
			go gui.NotifyNewCommTransaction(ct)
			return err
		})
	})
}

func createNewChartIfUpdatedTooLong() error {
	t, err := data.GetCurrentPartyUpdatedAt(db)
	if err == sql.ErrNoRows {
		log.Info("last party has no measurements")
		return nil
	}
	if err != nil {
		return err
	}
	log.Printf("last party updated at: %v, %v", t, time.Since(t))
	if time.Since(t) <= time.Hour {
		return nil
	}

	go gui.Popup("Для наглядности графичеких данных текущего опроса создан новый график.")

	log.Info("copy current party for new chart")
	if err := data.CopyCurrentParty(db); err != nil {
		return err
	}
	gui.NotifyCurrentPartyChanged()

	return nil
}

func switchGas(ctx context.Context, c cfg.Gas, valve byte) error {
	port := getComport(c.Comport)
	if err := port.SetConfig(comport.Config{
		Name:        c.Comport,
		Baud:        9600,
		ReadTimeout: time.Millisecond,
	}); err != nil {
		return merry.Append(err, "COM порт газового блока")
	}
	rdr := port.NewResponseReader(ctx, comm.Config{
		TimeoutGetResponse: c.TimeoutGetResponse,
		TimeoutEndResponse: c.TimeoutEndResponse,
		MaxAttemptsRead:    c.MaxAttemptsRead,
		Pause:              0,
	})
	return gas.Switch(c.Type, log, rdr, c.Addr, valve)
}
