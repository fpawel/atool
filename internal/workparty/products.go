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
	"github.com/fpawel/atool/internal/pkg/numeth"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"math"
	"sort"
)

type ErrorsOccurred map[string]struct{}

type WorkProduct = func(log comm.Logger, ctx context.Context, product Product) error

func ProcessEachActiveProduct(errs ErrorsOccurred, work WorkProduct) workgui.WorkFunc {
	return func(log comm.Logger, ctx context.Context) error {
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

			err := work(log, ctx, Product{
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
}

func ReadAndSaveProductParam(param modbus.Var, format modbus.FloatBitsFormat, dbKey string) workgui.WorkFunc {
	return workgui.NewFunc(fmt.Sprintf("📥 считывание регистр %d %v 💾 %s", param, format, dbKey),
		ProcessEachActiveProduct(nil, func(log comm.Logger, ctx context.Context, product Product) error {
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
				return merry.Appendf(err, "📥 считать %s регистр %d 💾 сохранить %s", format, param, dbKey)
			}
			workgui.NotifyInfo(log, fmt.Sprintf("%s считать регистр %d = %v 💾 сохранить %s", product, param, value, dbKey))
			return nil
		}))
}

func Write32(cmd modbus.DevCmd, format modbus.FloatBitsFormat, value float64) workgui.WorkFunc {
	return workgui.NewFunc(fmt.Sprintf("📥 отправка команды %d(%v) всем приборам сети", cmd, value),
		ProcessEachActiveProduct(nil, func(log comm.Logger, ctx context.Context, product Product) error {
			name := fmt.Sprintf("📥 команда %d(%v)", cmd, value)
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
			workgui.NotifyInfo(log, fmt.Sprintf("%s %s - успешно", product, name))
			return nil
		}))
}

func ReadProductsParams(ms *data.MeasurementCache, errorsOccurred ErrorsOccurred) workgui.WorkFunc {
	return ProcessEachActiveProduct(errorsOccurred, func(log comm.Logger, ctx context.Context, product Product) error {
		return product.readParams(log, ctx, ms)
	})
}

func WriteCfs(ks CfsList, format modbus.FloatBitsFormat) workgui.WorkFunc {
	return workgui.NewFunc("📥 запись коэффициентов %v %v",
		ProcessEachActiveProduct(nil, func(log comm.Logger, ctx context.Context, product Product) error {
			for _, k := range ks {
				var value float64
				err := data.DB.Get(&value,
					`SELECT value FROM product_value WHERE product_id = ? AND key = ?`,
					product.ProductID, data.KeyCoefficient(k))
				if err == sql.ErrNoRows {
					workgui.NotifyErr(log, fmt.Errorf("нет значения коэффициента %d", k))
					continue
				}
				if err := product.WriteKef(k, format, value)(log, ctx); err != nil {
					return err
				}
			}
			return nil
		}))
}

func ReadCfs(ks CfsList, format modbus.FloatBitsFormat) workgui.WorkFunc {
	return workgui.NewFunc(fmt.Sprintf("📥 💾 считывание коэффициентов %v %v", ks, format),
		ProcessEachActiveProduct(nil, func(log comm.Logger, ctx context.Context, product Product) error {
			for _, k := range ks {
				if _, err := product.ReadKef(log, ctx, k, format); err != nil {
					return err
				}
			}
			return nil
		}))
}

type ProductCoefficientValue struct {
	ProductID   int64
	Coefficient modbus.Coefficient
	Value       float64
}

type HandleProductCfErr = func(ProductCoefficientValue, error) error

func WriteProdsCfs(productCoefficientValues []ProductCoefficientValue, handleError HandleProductCfErr) workgui.WorkFunc {
	return func(log comm.Logger, ctx context.Context) error {
		party, err := data.GetCurrentParty()
		if err != nil {
			return err
		}

		device, err := appcfg.Cfg.Hardware.GetDevice(party.DeviceType)
		if err != nil {
			return err
		}

		noAnswer := map[int64]struct{}{}
		for _, x := range productCoefficientValues {

			if _, f := noAnswer[x.ProductID]; f {
				continue
			}

			if ctx.Err() != nil {
				return ctx.Err()
			}

			valFmt, err := device.GetCoefficientFormat(x.Coefficient)
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
			if err := p.WriteKef(x.Coefficient, valFmt, x.Value)(log, ctx); err != nil {
				if merry.Is(err, context.DeadlineExceeded) {
					noAnswer[x.ProductID] = struct{}{}
				}
				if handleError != nil {
					if err := handleError(x, err); err != nil {
						return err
					}
				}
				continue
			}

			// сохранить значение к-та
			if err := data.SaveProductKefValue(x.ProductID, x.Coefficient, x.Value); err != nil {
				return err
			}
		}
		return nil
	}
}

type ProductValues struct {
	Product            Product
	ProductIDKeyValues data.ProductIDKeyValues
}

func (x ProductValues) KefNaN(kef modbus.Coefficient) float64 {
	return x.GetNaN(data.KeyCoefficient(kef))
}

func (x ProductValues) GetNaN(dbKey string) float64 {
	v, ok := x.Get(dbKey)
	if !ok {
		return math.NaN()
	}
	return v
}

func (x ProductValues) Get(dbKey string) (float64, bool) {
	v, f := x.ProductIDKeyValues[data.ProductIDKey{
		ProductID: x.Product.ProductID,
		Key:       dbKey,
	}]
	return v, f
}

func (x ProductValues) GetValuesNaN(keys []string) (values []float64) {
	for _, key := range keys {
		values = append(values, x.GetNaN(key))
	}
	return
}

type InterpolateCfsFunc func(pv ProductValues) ([]numeth.Coordinate, error)

type InterpolateCfs struct {
	Name               string
	Coefficient        modbus.Coefficient
	Count              modbus.Coefficient
	Format             modbus.FloatBitsFormat
	InterpolateCfsFunc InterpolateCfsFunc
}

func (x InterpolateCfs) String() string {
	return fmt.Sprintf("📈 расчёт %s 📥 💾 запись K%d...K%d", x.Name, x.Coefficient, x.Coefficient+x.Count)
}

func (x InterpolateCfs) performProduct(productsValues data.ProductIDKeyValues, product Product) workgui.WorkFunc {
	return workgui.New(product.String(), func(log comm.Logger, ctx context.Context) error {
		pv := ProductValues{
			Product:            product,
			ProductIDKeyValues: productsValues,
		}
		dt, err := x.InterpolateCfsFunc(pv)
		if err != nil {
			workgui.NotifyErr(log, merry.Prepend(err, "нет данных для расчёта"))
			return nil
		}
		sort.Slice(dt, func(i, j int) bool {
			return dt[i][0] < dt[i][1]
		})
		workgui.NotifyInfo(log, fmt.Sprintf("📝 таблица интерполяции: %v", dt))

		r, ok := numeth.InterpolationCoefficients(dt)
		if !ok {
			r = make([]float64, len(dt))
			for i := range r {
				r[i] = math.NaN()
			}
			workgui.NotifyErr(log, merry.New("точки не интерполируемы"))
			return nil
		}
		for len(r) < int(x.Count) {
			r = append(r, 0)
		}

		workgui.NotifyInfo(log, fmt.Sprintf("📈 расчитано: %v", r))

		for i, value := range r {
			kef := x.Coefficient + modbus.Coefficient(i)
			if err := product.SaveKefValue(kef, value); err != nil {
				return err
			}
			_ = product.WriteKef(kef, x.Format, value)(log, ctx)
		}
		return nil
	}).Perform
}

func (x InterpolateCfs) Work() workgui.Work {
	return workgui.New(x.String(), func(log comm.Logger, ctx context.Context) error {
		productsValues, err := data.GetCurrentProductValues()
		if err != nil {
			return err
		}
		return ProcessEachActiveProduct(nil, func(log comm.Logger, ctx context.Context, product Product) error {
			return x.performProduct(productsValues, product)(log, ctx)
		})(log, ctx)
	})
}

type CfsList []modbus.Coefficient

func (x CfsList) String() string {
	var coefficients []int
	for _, k := range x {
		coefficients = append(coefficients, int(k))
	}
	return fmt.Sprintf("%v", intrng.IntRanges(coefficients))
}

func getCommProduct(comportName string, device devicecfg.Device) comm.T {
	return comm.New(comports.GetComport(comportName, device.Baud), device.CommConfig())
}
