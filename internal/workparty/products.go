package workparty

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config/appcfg"
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/comports"
	"github.com/fpawel/atool/internal/pkg/intrng"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/powerman/structlog"
)

type ErrorsOccurred map[string]struct{}

func ProcessEachActiveProduct(log comm.Logger, errs ErrorsOccurred, work func(Product) error) error {
	party, err := data.GetCurrentParty()
	if err != nil {
		return err
	}
	device, err := appcfg.Cfg.Hardware.GetDevice(party.DeviceType)
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
		workProduct := Product{
			Product: p,
			Device:  device,
			Party:   party,
		}

		processErr := func(err error) {
			if err == nil || merry.Is(err, context.Canceled) {
				return
			}
			if _, f := errs[err.Error()]; f {
				return
			}
			errs[err.Error()] = struct{}{}
			workgui.NotifyErr(log, merry.Prependf(err, "%s", workProduct))
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
		go gui.Popupf("опрашивается %s %s адрес %d %s", party.DeviceType, p.Comport, p.Addr, workProduct)

		err := work(Product{
			Product: p,
			Device:  device,
			Party:   party,
		})
		if merry.Is(err, context.Canceled) {
			return err
		}
		notifyConnection(err == nil)
		processErr(err)
	}
	return nil
}

func PerformWorkActiveEachProduct(log *structlog.Logger, ctx context.Context, name string, work func(Product) error) error {
	return workgui.Perform(log, ctx, name, func(log *structlog.Logger, ctx context.Context) error {
		return ProcessEachActiveProduct(log, nil, work)
	})
}

func ReadAndSaveProductParam(log *structlog.Logger, ctx context.Context, param modbus.Var, format modbus.FloatBitsFormat, dbKey string) error {
	what := fmt.Sprintf("📥 считать %s регистр %d 💾 сохранить %s", format, param, dbKey)
	return PerformWorkActiveEachProduct(log, ctx, what, func(product Product) error {
		value, err := modbus.Read3Value(log, ctx, product.Comm(), product.Addr, param, format)
		if err != nil {
			return err
		}
		const query = `
INSERT INTO product_value
VALUES (?, ?, ?)
ON CONFLICT (product_id,key) DO UPDATE
    SET value = ?`
		_, err = data.DB.Exec(query, product.ProductID, dbKey, value, value)
		if err != nil {
			return err
		}
		workgui.NotifyInfo(log, fmt.Sprintf("%s считать регистр %d = %v 💾 сохранить %s", product, param, value, dbKey))
		return nil
	})
}

func Write32(log comm.Logger, ctx context.Context, cmd modbus.DevCmd, format modbus.FloatBitsFormat, value float64) error {
	what := fmt.Sprintf("📥 команда %d(%v)", cmd, value)
	return PerformWorkActiveEachProduct(log, ctx, what, func(product Product) error {
		err := modbus.RequestWrite32{
			Addr:      product.Addr,
			ProtoCmd:  0x10,
			DeviceCmd: cmd,
			Format:    format,
			Value:     value,
		}.GetResponse(log, ctx, product.Comm())
		if err != nil {
			return err
		}
		workgui.NotifyInfo(log, fmt.Sprintf("%s %s - успешно", product, what))
		return nil
	})
}

func ReadProductsParams(log *structlog.Logger, ctx context.Context, ms *data.MeasurementCache, errorsOccurred ErrorsOccurred) error {
	return ProcessEachActiveProduct(log, errorsOccurred, func(product Product) error {
		return product.readParams(log, ctx, ms)
	})
}

func WriteCoefficients(log comm.Logger, ctx context.Context, ks CoefficientsList, format modbus.FloatBitsFormat) error {
	name := fmt.Sprintf("📥 запись коэффициентов %v %v", ks, format)
	return PerformWorkActiveEachProduct(log, ctx, name, func(product Product) error {
		for _, k := range ks {
			var value float64
			err := data.DB.Get(&value,
				`SELECT value FROM product_value WHERE product_id = ? AND key = ?`,
				product.ProductID, data.KeyCoefficient(int(k)))
			if err == sql.ErrNoRows {
				workgui.NotifyErr(log, fmt.Errorf("нет значения коэффициента %d", k))
				continue
			}
			if err := product.WriteKef(log, ctx, k, format, value); err != nil {
				return err
			}
		}
		return nil
	})
}

func ReadCoefficients(log comm.Logger, ctx context.Context, ks CoefficientsList, format modbus.FloatBitsFormat) error {
	name := fmt.Sprintf("📥 💾 считывание коэффициентов %v %v", ks, format)
	return PerformWorkActiveEachProduct(log, ctx, name, func(product Product) error {
		for _, k := range ks {
			if _, err := product.ReadKef(log, ctx, k, format); err != nil {
				return err
			}
		}
		return nil
	})
}

func getCommProduct(comportName string, device devicecfg.Device) comm.T {
	return comm.New(comports.GetComport(comportName, device.Baud), device.CommConfig())
}

func writeAllCoefficients(log *structlog.Logger, ctx context.Context, in []*apitypes.ProductCoefficientValue) error {

	party, err := data.GetCurrentParty()
	if err != nil {
		return err
	}

	device, err := appcfg.Cfg.Hardware.GetDevice(party.DeviceType)
	if err != nil {
		return err
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
		if err := p.WriteKef(log, ctx, modbus.Var(x.Coefficient), valFmt, x.Value); err != nil {
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
		return merry.New("не все коэффициенты записаны")
	}
	return nil
}

type CoefficientsList []modbus.Var

func (x CoefficientsList) String() string {
	var coefficients []int
	for _, k := range x {
		coefficients = append(coefficients, int(k))
	}
	return fmt.Sprintf("%v", intrng.IntRanges(coefficients))
}
