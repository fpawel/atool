package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/guiwork"
	"github.com/fpawel/atool/internal/journal"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/comports"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/powerman/structlog"
	"time"
)

var errNoInterrogateObjects = merry.New("не установлены объекты опроса")

func runInterrogate() error {
	return guiwork.RunWork(log, appCtx, "опрос приборов", func(log *structlog.Logger, ctx context.Context) error {
		ms := new(measurements)
		defer ms.Save()
		for {
			if err := readProductsParams(ctx, ms); err != nil {
				if merry.Is(err, context.Canceled) {
					return nil
				}
				return err
			}
		}
	})
}

func readProductsParams(ctx context.Context, ms *measurements) error {
	return processEachActiveProduct(func(product data.Product, device config.Device) error {
		rdr := newParamsReader(product, device)
		for _, prm := range device.Params {
			err := rdr.getResponse(ctx, prm)
			if err != nil {
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

func runReadAllCoefficients() error {
	return guiwork.RunWork(log, appCtx, "считывание коэффициентов", func(log *structlog.Logger, ctx context.Context) error {
		var xs []gui.CoefficientValue
		hasFormatErrors := false
		err := processEachActiveProduct(func(product data.Product, device config.Device) error {
			log := pkg.LogPrependSuffixKeys(log, "product", product.String())
			for _, k := range config.Get().Coefficients {
				count := k.Range[1] - k.Range[0] + 1
				log := pkg.LogPrependSuffixKeys(log, "range", fmt.Sprintf("%d...%d", k.Range[0], k.Range[1]))

				req := modbus.RequestRead3{
					Addr:           product.Addr,
					FirstRegister:  modbus.Var(224 + k.Range[0]*2),
					RegistersCount: uint16(count * 2),
				}
				cm := getCommProduct(product.Comport, device)
				response, err := req.GetResponse(log, ctx, cm)
				if err != nil {
					return err
				}
				n := k.Range[0]
				d := response[3 : len(response)-2]
				for i := 0; i < len(d); i, n = i+4, n+1 {
					d := d[i:][:4]
					if _, f := config.Get().InactiveCoefficients[n]; f {
						continue
					}
					x := gui.CoefficientValue{
						ProductID:   product.ProductID,
						What:        "read",
						Coefficient: n,
					}

					if v, err := k.Format.ParseFloat(d); err == nil {
						x.Result = formatFloat(v)
						x.Ok = true
					} else {
						x.Result = fmt.Sprintf("% X: %v", d, err)
						x.Ok = false
						hasFormatErrors = true
					}
					xs = append(xs, x)
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
		if len(xs) > 0 {
			go gui.NotifyCoefficients(xs)
		}

		if hasFormatErrors {
			return merry.New("один или несколько к-тов имеют не верный формат")
		}
		return nil
	})
}

func runWriteAllCoefficients(in []*apitypes.ProductCoefficientValue) error {
	return guiwork.RunWork(log, appCtx, "запись коэффициентов", func(log *structlog.Logger, ctx context.Context) error {

		for _, x := range in {
			valFmt, err := config.Get().GetCoefficientFormat(int(x.Coefficient))
			if err != nil {
				return err
			}

			var product data.Product
			if err := db.Get(&product, `SELECT * FROM product WHERE product_id = ?`, x.ProductID); err != nil {
				return err
			}
			journal.Info(log, fmt.Sprintf("%s %s адр.%d K%d=%v", product.Device,
				product.Comport, product.Addr, x.Coefficient, x.Value))

			device, f := config.Get().Hardware[product.Device]
			if !f {
				return fmt.Errorf("не заданы параметры устройства %s для прибора %+v",
					product.Device, product)
			}

			log := pkg.LogPrependSuffixKeys(log, "write_coefficient", x.Coefficient, "value", x.Value,
				"product", fmt.Sprintf("%+v", product))

			req := modbus.RequestWrite32{
				Addr:      product.Addr,
				ProtoCmd:  0x10,
				DeviceCmd: modbus.DevCmd((0x80 << 8) + x.Coefficient),
				Format:    valFmt,
				Value:     x.Value,
			}
			cm := getCommProduct(product.Comport, device)
			err = req.GetResponse(log, ctx, cm)
			if err != nil && !merry.Is(err, comm.Err) {
				return err
			}

		}
		return nil
	})
}

func runRawCommand(c modbus.ProtoCmd, b []byte) {
	guiwork.RunTask(log, fmt.Sprintf("отправка команды XX %X % X", c, b), func() error {
		err := processEachActiveProduct(func(p data.Product, device config.Device) error {
			cm := getCommProduct(p.Comport, device)
			req := modbus.Request{
				Addr:     p.Addr,
				ProtoCmd: c,
				Data:     b,
			}
			_, err := req.GetResponse(log, context.Background(), cm)
			return err
		})
		if err != nil {
			return err
		}
		return nil
	})
}

func readAndSaveProductValue(log logger, ctx context.Context, product data.Product, device config.Device, param modbus.Var, format modbus.FloatBitsFormat, dbKey string) error {
	wrapErr := func(err error) error {
		return merry.Appendf(err, "прибор %d.%d: считать рег.%d %s: сохранить %q",
			product.Serial, product.ProductID, param, format, dbKey)
	}
	cm := getCommProduct(product.Comport, device)
	value, err := modbus.Read3Value(log, ctx, cm, product.Addr, modbus.Var(param), format)
	if err != nil {
		journal.Err(log, wrapErr(err))
		return nil
	}
	journal.Info(log, fmt.Sprintf("прибор %d.%d: сохранить рег.%d,%s = %v",
		product.Serial, product.ProductID, param, dbKey, value))
	const query = `
INSERT INTO product_value
VALUES (?, ?, ?)
ON CONFLICT (product_id,key) DO UPDATE
    SET value = ?`
	_, err = db.Exec(query, product.ProductID, dbKey, value, value)
	return wrapErr(err)
}

//func createNewChartIfUpdatedTooLong() error {
//	t, err := data.GetCurrentPartyUpdatedAt(db)
//	if err == sql.ErrNoRows {
//		log.Info("last party has no measurements")
//		return nil
//	}
//	if err != nil {
//		return err
//	}
//	//log.Printf("last party updated at: %v, %v", t, time.Since(t))
//	if time.Since(t) <= time.Hour {
//		return nil
//	}
//
//	go gui.Popup(true, "Для наглядности графичеких данных текущего опроса создан новый график.")
//
//	log.Info("copy current party for new chart")
//	if err := data.CopyCurrentParty(db); err != nil {
//		return err
//	}
//	gui.NotifyCurrentPartyChanged()
//
//	return nil
//}

func getActiveProducts() ([]data.Product, error) {
	var products []data.Product
	err := db.Select(&products,
		`SELECT * FROM product WHERE party_id = (SELECT party_id FROM app_config) AND active`)
	if err != nil {
		return nil, err
	}
	if len(products) == 0 {
		return nil, errNoInterrogateObjects
	}
	return products, nil
}

func processEachActiveProduct(work func(data.Product, config.Device) error) error {
	products, err := getActiveProducts()
	if err != nil {
		return err
	}
	for _, p := range products {
		d, f := config.Get().Hardware[p.Device]
		if !f {
			return fmt.Errorf("не заданы параметры устройства %s для прибора %+v",
				p.Device, p)
		}
		gui.Popupf("опрашивается прибор: %s %s адр.%d", p.Device, p.Comport, p.Addr)
		err := work(p, d)
		if merry.Is(err, context.Canceled) {
			return err
		}
		go gui.NotifyProductConnection(gui.ProductConnection{
			ProductID: p.ProductID,
			Ok:        err == nil,
		})
		if err != nil {
			journal.Err(log, merry.Errorf("ошибка связи с прибором %s %s адр.%d", p.Device, p.Comport, p.Addr).WithCause(err))
		}
	}
	return nil
}

func getCommProduct(comportName string, device config.Device) comm.T {
	return comm.New(comports.GetComport(comportName, device.Baud), device.CommConfig())
}

func delay(log *structlog.Logger, ctx context.Context, duration time.Duration, name string) error {
	// измерения, полученные в процесе опроса приборов во время данной задержки
	ms := new(measurements)
	defer ms.Save()

	return guiwork.Delay(log, ctx, duration, name, func(_ *structlog.Logger, ctx context.Context) error {
		return readProductsParams(ctx, ms)
	})
}
