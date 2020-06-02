package workparty

import (
	"context"
	"errors"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/comports"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/comm"
	"github.com/powerman/structlog"
)

type ErrorsOccurred map[string]struct{}

func ProcessEachActiveProduct(log comm.Logger, errs ErrorsOccurred, work func(Product) error) error {
	party, err := data.GetCurrentParty()
	if err != nil {
		return err
	}
	device, err := config.Get().Hardware.GetDevice(party.DeviceType)
	if err != nil {
		return err
	}
	products, err := data.GetActiveProducts()
	if err != nil {
		return err
	}

	if errs == nil {
		errs = ErrorsOccurred{}
	}

	for _, p := range products {
		p := p
		processErr := func(err error) {
			if err == nil || merry.Is(err, context.Canceled) {
				return
			}
			if _, f := errs[err.Error()]; f {
				return
			}
			errs[err.Error()] = struct{}{}
			workgui.NotifyErr(log, merry.Prependf(err, "ошибка связи с прибором №%d", p.Serial))
		}
		notifyConnection := func(ok bool) {
			go gui.NotifyProductConnection(gui.ProductConnection{
				ProductID: p.ProductID,
				Ok:        ok,
			})
		}

		if err := comports.GetComport(p.Comport, device.Baud).Open(); err != nil {
			processErr(err)
			notifyConnection(false)
			continue
		}
		go gui.Popupf("опрашивается прибор: №%d %s адр.%d", p.Serial, p.Comport, p.Addr)
		err := work(Product{
			Product: p,
			Device:  device,
		})
		if merry.Is(err, context.Canceled) {
			return err
		}
		notifyConnection(err == nil)
		processErr(err)
	}
	return nil
}

func writeAllCoefficients(log *structlog.Logger, ctx context.Context, in []*apitypes.ProductCoefficientValue) error {

	party, err := data.GetCurrentParty()
	if err != nil {
		return err
	}

	cfg := config.Get()

	device, f := cfg.Hardware[party.DeviceType]
	if !f {
		return merry.Errorf("не заданы параметры устройства %s", party.DeviceType)
	}

	noAnswer := map[int64]struct{}{}
	var errorsOccurred bool
	for _, x := range in {

		if _, f := noAnswer[x.ProductID]; f {
			continue
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		valFmt, err := device.GetCoefficientFormat(int(x.Coefficient))
		if err != nil {
			return err
		}

		product, productFound := party.GetProduct(x.ProductID)
		if !productFound {
			return merry.Errorf("product_id not found: %+v", x)
		}
		log := pkg.LogPrependSuffixKeys(log, "write_coefficient", x.Coefficient, "value", x.Value,
			"product", fmt.Sprintf("%+v", product))

		p := Product{
			Product: product,
			Device:  device,
		}
		if err := p.WriteKef(log, ctx, int(x.Coefficient), valFmt, x.Value); err != nil {
			if merry.Is(err, context.DeadlineExceeded) {
				noAnswer[x.ProductID] = struct{}{}
			}
			errorsOccurred = true
			continue
		}

		// сохранить значение к-та
		if err := data.SaveProductKefValue(x.ProductID, int(x.Coefficient), x.Value); err != nil {
			return err
		}
	}
	if errorsOccurred {
		return errors.New("не все коэффициенты записаны")
	}
	return nil
}

func readProductsParams(log *structlog.Logger, ctx context.Context, ms *data.MeasurementCache, errorsOccurred ErrorsOccurred) error {
	return ProcessEachActiveProduct(log, errorsOccurred, func(product Product) error {
		return product.readParams(log, ctx, ms)
	})
}

func getCommProduct(comportName string, device devicecfg.Device) comm.T {
	return comm.New(comports.GetComport(comportName, device.Baud), device.CommConfig())
}
