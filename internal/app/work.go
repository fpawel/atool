package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/guiwork"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/comports"
	"github.com/fpawel/atool/internal/pkg/intrng"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/powerman/structlog"
	"time"
)

var errNoInterrogateObjects = merry.New("не установлены объекты опроса")

func searchProducts(log comm.Logger, ctx context.Context, comportName string) error {
	party, err := data.GetCurrentParty()
	if err != nil {
		return err
	}
	device, err := config.Get().Hardware.GetDevice(party.DeviceType)
	if err != nil {
		return err
	}

	if len(device.Params) == 0 {
		return fmt.Errorf("нет параметров устройства %q", party.DeviceType)
	}

	cm := comm.New(comports.GetComport(comportName, device.Baud), comm.Config{
		TimeoutGetResponse: 500 * time.Millisecond,
		TimeoutEndResponse: 50 * time.Millisecond,
	})

	ans, notAns := make(intrng.Bytes), make(intrng.Bytes)
	param := device.Params[0]

	go gui.NotifyProgressShow(127, "MODBUS: сканирование сети")
	defer func() {
		go gui.NotifyProgressHide()
	}()

	for addr := modbus.Addr(1); addr <= 127; addr++ {
		go gui.NotifyProgress(int(addr), fmt.Sprintf("MODBUS: сканирование сети: %d, ответили [%s], не ответили [%s]",
			addr, ans.Format(), notAns.Format()))
		_, err := modbus.Read3Value(log, ctx, cm, addr, modbus.Var(param.ParamAddr), param.Format)
		if merry.Is(err, context.DeadlineExceeded) || merry.Is(err, modbus.Err) {
			notAns.Push(byte(addr))
			continue
		}
		if err != nil {
			return err
		}
		ans.Push(byte(addr))
	}

	if len(ans) == 0 {
		go gui.NotifyStatus(gui.Status{
			Text:       "сканирование сети: приборы не найдены",
			Ok:         true,
			PopupLevel: gui.LWarn,
		})
		return nil
	}

	if err := data.SetNewCurrentParty(len(ans)); err != nil {
		return err
	}
	party, err = data.GetCurrentParty()
	if err != nil {
		return err
	}

	for i, addr := range ans.Slice() {
		p := party.Products[i]
		p.Addr = modbus.Addr(addr)
		if err := data.UpdateProduct(p); err != nil {
			return err
		}
	}
	go func() {
		gui.NotifyCurrentPartyChanged()
		gui.NotifyStatus(gui.Status{
			Text: fmt.Sprintf("сканирование сети: создана новая партия %d. Ответили [%s], не ответили [%s]",
				party.PartyID, ans.Format(), notAns.Format()),
			Ok:         true,
			PopupLevel: gui.LWarn,
		})
	}()

	return nil

}

func runInterrogate() error {
	return guiwork.RunWork(log, appCtx, "опрос приборов", func(log *structlog.Logger, ctx context.Context) error {
		ms := new(measurements)
		defer ms.Save()

		errorsOccurred := make(errorsOccurred)
		for {
			if err := readProductsParams(ctx, ms, errorsOccurred); err != nil {
				if merry.Is(err, context.Canceled) {
					return nil
				}
				return err
			}
		}
	})
}

func readProductsParams(ctx context.Context, ms *measurements, errorsOccurred errorsOccurred) error {
	return processEachActiveProduct(errorsOccurred, func(product data.Product, device devicecfg.Device) error {
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

func notifyReadCoefficient(p data.Product, n int, value float64, err error) {
	x := gui.CoefficientValue{
		ProductID:   p.ProductID,
		Read:        true,
		Coefficient: n,
	}
	if err == nil {
		x.Result = formatFloat(value)
		x.Ok = true
		guiwork.JournalInfo(log, fmt.Sprintf("считано: №%d K%d=%v", p.Serial, n, value))
	} else {
		err = fmt.Errorf("считывание №%d K%d: %w", p.Serial, n, err)
		x.Result = err.Error()
		guiwork.JournalErr(log, err)
		x.Ok = false
	}
	go gui.NotifyCoefficient(x)
}

func runReadAllCoefficients() error {

	return guiwork.RunWork(log, appCtx, "считывание коэффициентов", func(log *structlog.Logger, ctx context.Context) error {

		inactiveCoefficients := config.Get().InactiveCoefficients
		err := processEachActiveProduct(nil, func(product data.Product, device devicecfg.Device) error {
			log := pkg.LogPrependSuffixKeys(log, "product", product.String())
			cm := getCommProduct(product.Comport, device)
			for _, Kr := range device.Coefficients {
				log := pkg.LogPrependSuffixKeys(log, "range", fmt.Sprintf("%d...%d", Kr.Range[0], Kr.Range[1]))
				for kef := Kr.Range[0]; kef <= Kr.Range[1]; kef++ {
					if ctx.Err() != nil {
						return ctx.Err()
					}
					if _, f := inactiveCoefficients[kef]; f {
						continue
					}
					value, err := modbus.Read3Value(log, ctx, cm, product.Addr, 224+2*modbus.Var(kef), Kr.Format)
					notifyReadCoefficient(product, kef, value, err)
					if err != nil {
						continue
					}
					// сохранить значение к-та
					if err := saveProductKefValue(product.ProductID, kef, value); err != nil {
						return err
					}
				}
			}
			return nil
		})
		return err
	})
}

func runWriteAllCoefficients(in []*apitypes.ProductCoefficientValue) error {

	party, err := data.GetCurrentParty()
	if err != nil {
		return err
	}

	cfg := config.Get()

	device, f := cfg.Hardware[party.DeviceType]
	if !f {
		return fmt.Errorf("не заданы параметры устройства %s", party.DeviceType)
	}

	return guiwork.RunWork(log, appCtx, "запись коэффициентов", func(log *structlog.Logger, ctx context.Context) error {
		for _, x := range in {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			valFmt, err := device.GetCoefficientFormat(int(x.Coefficient))
			if err != nil {
				return err
			}

			product, productFound := party.GetProduct(x.ProductID)
			if !productFound {
				return fmt.Errorf("product_id not found: %+v", x)
			}

			log := pkg.LogPrependSuffixKeys(log, "write_coefficient", x.Coefficient, "value", x.Value,
				"product", fmt.Sprintf("%+v", product))

			if err := writeKefProduct(log, ctx, product, device, int(x.Coefficient), valFmt, x.Value); err != nil {
				continue
			}

			// сохранить значение к-та
			if err := saveProductKefValue(x.ProductID, int(x.Coefficient), x.Value); err != nil {
				return err
			}
		}
		return nil
	})
}

func runRawCommand(c modbus.ProtoCmd, b []byte) {
	guiwork.RunTask(log, fmt.Sprintf("отправка команды XX %X % X", c, b), func() error {
		err := processEachActiveProduct(nil, func(p data.Product, device devicecfg.Device) error {
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

func readAndSaveProductValue(log logger, ctx context.Context, product data.Product, device devicecfg.Device, param modbus.Var, format modbus.FloatBitsFormat, dbKey string) error {
	wrapErr := func(err error) error {
		return merry.Appendf(err, "прибор %d.%d: считать рег.%d %s: сохранить %q",
			product.Serial, product.ProductID, param, format, dbKey)
	}
	cm := getCommProduct(product.Comport, device)
	value, err := modbus.Read3Value(log, ctx, cm, product.Addr, param, format)
	if err != nil {
		guiwork.JournalErr(log, wrapErr(err))
		return nil
	}
	guiwork.JournalInfo(log, fmt.Sprintf("прибор %d.%d: сохранить рег.%d,%s = %v",
		product.Serial, product.ProductID, param, dbKey, value))
	const query = `
INSERT INTO product_value
VALUES (?, ?, ?)
ON CONFLICT (product_id,key) DO UPDATE
    SET value = ?`
	_, err = data.DB.Exec(query, product.ProductID, dbKey, value, value)
	return wrapErr(err)
}

func write32Product(log logger, ctx context.Context, product data.Product, device devicecfg.Device, cmd modbus.DevCmd, format modbus.FloatBitsFormat, value float64) error {
	err := modbus.RequestWrite32{
		Addr:      product.Addr,
		ProtoCmd:  0x10,
		DeviceCmd: cmd,
		Format:    format,
		Value:     value,
	}.GetResponse(log, ctx, getCommProduct(product.Comport, device))

	if err == nil {
		guiwork.JournalInfo(log, fmt.Sprintf("прибор №%d: команда %d(%v)", product.Serial, cmd, value))
	} else {
		guiwork.JournalErr(log, fmt.Errorf("прибор №%d: команда %d(%v): %w", product.Serial, cmd, value, err))
	}
	return err
}

func writeKefProduct(log logger, ctx context.Context, product data.Product, device devicecfg.Device, kef int, format modbus.FloatBitsFormat, value float64) error {
	err := modbus.RequestWrite32{
		Addr:      product.Addr,
		ProtoCmd:  0x10,
		DeviceCmd: (0x80 << 8) + modbus.DevCmd(kef),
		Format:    format,
		Value:     value,
	}.GetResponse(log, ctx, getCommProduct(product.Comport, device))

	x := gui.CoefficientValue{
		ProductID:   product.ProductID,
		Read:        false,
		Coefficient: kef,
	}
	if err == nil {
		x.Result = formatFloat(value)
		x.Ok = true
		guiwork.JournalInfo(log, fmt.Sprintf("№%d.id%d: записано: K%d=%v", product.Serial, product.ProductID, kef, value))
	} else {
		err = fmt.Errorf("запись №%d K%d=%v: %w", product.Serial, kef, value, err)
		x.Result = err.Error()
		guiwork.JournalErr(log, err)
		x.Ok = false
	}
	go gui.NotifyCoefficient(x)

	return err
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

type errorsOccurred map[string]struct{}

func processEachActiveProduct(errorsOccurred errorsOccurred, work func(data.Product, devicecfg.Device) error) error {
	party, err := data.GetCurrentParty()
	if err != nil {
		return err
	}
	device, err := config.Get().Hardware.GetDevice(party.DeviceType)
	if err != nil {
		return err
	}
	products, err := getActiveProducts()
	if err != nil {
		return err
	}
	for _, p := range products {
		go gui.Popupf("опрашивается прибор: №%d %s адр.%d", p.Serial, p.Comport, p.Addr)
		err := work(p, device)
		if merry.Is(err, context.Canceled) {
			return err
		}
		go gui.NotifyProductConnection(gui.ProductConnection{
			ProductID: p.ProductID,
			Ok:        err == nil,
		})
		if err != nil {
			if errorsOccurred == nil {
				guiwork.JournalErr(log, merry.Errorf("ошибка связи с прибором №%d", p.Serial).WithCause(err))
			} else {
				if _, f := errorsOccurred[err.Error()]; !f {
					errorsOccurred[err.Error()] = struct{}{}
					guiwork.JournalErr(log, merry.Errorf("ошибка связи с прибором №%d", p.Serial).WithCause(err))
				}
			}
		}
	}
	return nil
}

func getCommProduct(comportName string, device devicecfg.Device) comm.T {
	return comm.New(comports.GetComport(comportName, device.Baud), device.CommConfig())
}

func delay(log *structlog.Logger, ctx context.Context, duration time.Duration, name string) error {
	// измерения, полученные в процесе опроса приборов во время данной задержки
	ms := new(measurements)
	defer ms.Save()

	errorsOccurred := make(errorsOccurred)

	return guiwork.Delay(log, ctx, duration, name, func(_ *structlog.Logger, ctx context.Context) error {
		return readProductsParams(ctx, ms, errorsOccurred)
	})
}
