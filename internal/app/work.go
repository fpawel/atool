package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/gui/guiwork"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"time"
)

var errNoInterrogateObjects = merry.New("не установлены объекты опроса")

func runInterrogate() error {
	return guiwork.RunWork(appCtx, "опрос приборов", func(ctx context.Context) (string, error) {
		must.PanicIf(createNewChartIfUpdatedTooLong())
		ms := new(measurements)
		defer func() {
			saveMeasurements(ms.xs)
		}()
		for {
			if err := readProductsParams(ctx, ms); err != nil {
				if merry.Is(err, context.Canceled) {
					return "", nil
				}
				return "", err
			}
		}
	})
}

func readProductsParams(ctx context.Context, ms *measurements) error {
	return processEachActiveProduct(func(product data.Product, device config.Device) error {
		rdr := newParamsReader(product, device)
		for _, prm := range device.Params {
			if err := rdr.getResponse(ctx, prm); err != nil {
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
	return guiwork.RunWork(appCtx, "считывание коэффициентов", func(ctx context.Context) (string, error) {
		log := pkg.LogPrependSuffixKeys(log, "work", "считывание_коэффициентов")
		var xs []gui.CoefficientValue
		hasFormatErrors := false
		err := processEachActiveProduct(func(product data.Product, device config.Device) error {
			log := pkg.LogPrependSuffixKeys(log, "product", product.String())
			for _, k := range config.Get().Coefficients {
				count := k.Range[1] - k.Range[0] + 1
				ct := commTransaction{
					what:        fmt.Sprintf("read K:%d...%d,%s", k.Range[0], k.Range[1], k.Format),
					req:         modbus.RequestRead3(product.Addr, modbus.Var(224+k.Range[0]*2), uint16(count*2)),
					device:      device,
					comportName: product.Comport,
					prs: func(request, response []byte) (s string, e error) {
						if len(response) != count*4+5 {
							return "", merry.Errorf("длина ответа %d не равна %d", len(response), count*4+5)
						}
						return "", nil
					},
				}
				log := pkg.LogPrependSuffixKeys(log, "range", fmt.Sprintf("%d...%d", k.Range[0], k.Range[1]))
				response, err := ct.getResponse(log, ctx)
				if err != nil {
					//go gui.PopupError(true, fmt.Errorf("%s: %w", product, err))
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
			return "", err
		}
		if len(xs) > 0 {
			go gui.NotifyCoefficients(xs)
		}

		if hasFormatErrors {
			return "", errors.New("один или несколько к-тов имеют не верный формат")
		}
		return "см. таблицу во вкладке коэффициентов", nil

	})
}

func runWriteAllCoefficients(in []*apitypes.ProductCoefficientValue) error {
	return guiwork.RunWork(appCtx, "запись коэффициентов", func(ctx context.Context) (string, error) {
		log := pkg.LogPrependSuffixKeys(log, "work", "запись_коэффициентов")

		for _, x := range in {
			valFmt, err := config.Get().GetCoefficientFormat(int(x.Coefficient))
			if err != nil {
				return "", err
			}
			var product data.Product
			if err := db.Get(&product, `SELECT * FROM product WHERE product_id = ?`, x.ProductID); err != nil {
				return "", err
			}
			device, okDevice := config.Get().Hardware.DeviceByName(product.Device)
			if !okDevice {
				return "", fmt.Errorf("не найден тип проибора %q", product.Device)
			}
			go gui.Popup(false, fmt.Sprintf("%s %s адр.%d K%d=%v", product.Device,
				product.Comport, product.Addr, x.Coefficient, x.Value))
			ct := commTransaction{
				what: fmt.Sprintf("write K%d=%v", x.Coefficient, x.Value),
				req: modbus.Request{
					Addr:     product.Addr,
					ProtoCmd: 0x10,
					Data:     requestWrite32Bytes(uint16((0x80<<8)+x.Coefficient), x.Value, valFmt),
				},
				device:      device,
				comportName: product.Comport,
			}
			log := pkg.LogPrependSuffixKeys(log, "write_coefficient", x.Coefficient, "value", x.Value,
				"product", fmt.Sprintf("%+v", product))
			_, err = ct.getResponse(log, ctx)
			if err != nil && !merry.Is(err, comm.Err) {
				return "", err
			}
		}
		return "", nil
	})
}

func runRawCommand(c modbus.ProtoCmd, b []byte) {
	guiwork.RunTask(fmt.Sprintf("отправка команды XX %X % X", c, b), func() (string, error) {
		err := processEachActiveProduct(func(p data.Product, device config.Device) error {
			ct := commTransaction{
				comportName: p.Comport,
				device:      device,
				req: modbus.Request{
					Addr:     p.Addr,
					ProtoCmd: c,
					Data:     b,
				},
			}
			_, err := ct.getResponse(log, context.Background())
			return err
		})
		if err != nil {
			return "", err
		}
		return "", nil
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
	//log.Printf("last party updated at: %v, %v", t, time.Since(t))
	if time.Since(t) <= time.Hour {
		return nil
	}

	go gui.Popup(true, "Для наглядности графичеких данных текущего опроса создан новый график.")

	log.Info("copy current party for new chart")
	if err := data.CopyCurrentParty(db); err != nil {
		return err
	}
	gui.NotifyCurrentPartyChanged()

	return nil
}

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
		d, f := config.Get().Hardware.DeviceByName(p.Device)
		if !f {
			return fmt.Errorf("не заданы параметры устройства %s для прибора %+v",
				p.Device, p)
		}
		go gui.Popup(false, fmt.Sprintf("опрашивается прибор: %s %s адр.%d", d.Name, p.Comport, p.Addr))
		if err := work(p, d); merry.Is(err, context.Canceled) {
			return err
		}
	}
	return nil
}
