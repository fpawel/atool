package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/gui"
	"github.com/fpawel/atool/internal/cfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/comm/modbus"
	"time"
)

func runInterrogate() {
	runWork("опрос приборов", func(ctx context.Context) (string, error) {
		must.PanicIf(createNewChartIfUpdatedTooLong())
		ms := new(measurements)
		defer func() {
			saveMeasurements(ms.xs)
		}()
		for {
			if err := processProductsParams(ctx, ms); err != nil {
				if merry.Is(err, context.Canceled) {
					return "", nil
				}
				return "", err
			}
		}
	})
}

func processProductsParams(ctx context.Context, ms *measurements) error {
	return processEachActiveProduct(func(product data.Product, device cfg.Device) error {
		rdr, err := newParamsReader(product, device)
		if err != nil {
			return err
		}
		for _, prm := range device.Params {
			if err := rdr.read(ctx, prm); err != nil {
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

func runReadAllCoefficients() {
	runWork("считывание коэффициентов", func(ctx context.Context) (string, error) {
		c := cfg.Get()
		log := pkg.LogPrependSuffixKeys(log, "work", "считывание_коэффициентов")
		var xs []gui.CoefficientValue
		hasFormatErrors := false
		err := processEachActiveProduct(func(product data.Product, device cfg.Device) error {
			for _, k := range c.Coefficients {
				count := k.Range[1] - k.Range[0] + 1
				ct := commTransaction{
					what:        fmt.Sprintf("считывание коэффициентов:%d...%d,%s", k.Range[0], k.Range[1], k.Format),
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
					return err
				}
				n := k.Range[0]
				d := response[3 : len(response)-2]
				for i := 0; i < len(d); i, n = i+4, n+1 {
					if _, f := c.InactiveCoefficients[n]; f {
						continue
					}

					x := gui.CoefficientValue{
						ProductID:   product.ProductID,
						What:        "read",
						Coefficient: n,
					}
					d := response[i:][:4]
					if v, err := parseFloatBits(string(k.Format), d); err == nil {
						x.Result = fmt.Sprintf("%v", v)
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
		go gui.NotifyCoefficients(xs)
		if hasFormatErrors {
			return "", errors.New("один или несколько к-тов имеют не верный формат")
		}
		return "см. таблицу во вкладке коэффициентов", nil

	})
}

func runRawCommand(c modbus.ProtoCmd, b []byte) {
	runTask(fmt.Sprintf("отправка команды XX %X % X", c, b), func() (string, error) {
		err := processEachActiveProduct(func(p data.Product, device cfg.Device) error {
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
	log.Printf("last party updated at: %v, %v", t, time.Since(t))
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
