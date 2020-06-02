package workparty

import (
	"context"
	"errors"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg/comports"
	"github.com/fpawel/atool/internal/pkg/intrng"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/powerman/structlog"
	"time"
)

func Delay(log *structlog.Logger, ctx context.Context, duration time.Duration, name string) error {
	// измерения, полученные в процесе опроса приборов во время данной задержки
	ms := new(data.MeasurementCache)
	defer ms.Save()

	errorsOccurred := make(ErrorsOccurred)

	return workgui.Delay(log, ctx, duration, name, func(_ *structlog.Logger, ctx context.Context) error {
		return readProductsParams(log, ctx, ms, errorsOccurred)
	})
}

func RunInterrogate(log comm.Logger, appCtx context.Context) error {
	return workgui.RunWork(log, appCtx, "опрос приборов", func(log *structlog.Logger, ctx context.Context) error {
		ms := new(data.MeasurementCache)
		defer ms.Save()
		errorsOccurred := make(ErrorsOccurred)
		for {
			if err := readProductsParams(log, ctx, ms, errorsOccurred); err != nil {
				if merry.Is(err, context.Canceled) {
					return nil
				}
				return err
			}
		}
	})
}

func RunReadAllCoefficients(log comm.Logger, appCtx context.Context) error {

	return workgui.RunWork(log, appCtx, "считывание коэффициентов", func(log *structlog.Logger, ctx context.Context) error {
		errs := make(ErrorsOccurred)
		err := ProcessEachActiveProduct(log, errs, func(p Product) error {
			return p.readAllCoefficients(log, ctx)
		})
		if err != nil {
			return err
		}
		if len(errs) > 0 {
			return errors.New("не все коэффициенты считаны")
		}
		return nil
	})
}

func RunWriteAllCoefficients(log comm.Logger, appCtx context.Context, in []*apitypes.ProductCoefficientValue) error {

	return workgui.RunWork(log, appCtx, "запись коэффициентов", func(log *structlog.Logger, ctx context.Context) error {
		return writeAllCoefficients(log, ctx, in)
	})
}

func RunRawCommand(log comm.Logger, appCtx context.Context, c modbus.ProtoCmd, b []byte) {
	workgui.RunTask(log, fmt.Sprintf("отправка команды XX %X % X", c, b), func() error {
		err := ProcessEachActiveProduct(log, nil, func(p Product) error {
			_, err := p.GetResponse(log, appCtx, c, b)
			return err
		})
		if err != nil {
			return err
		}
		return nil
	})
}

func RunSetNetAddr(log comm.Logger, appCtx context.Context, productID int64, notifyComm func(comm.Info)) error {
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
		return merry.Errorf("не заданы параметры устройства %s для прибора %+v", party.DeviceType, p)
	}

	return workgui.RunWork(log, appCtx, fmt.Sprintf("прибр %d: запись сетевого адреса %d", p.Serial, p.Addr),
		func(log comm.Logger, ctx context.Context) error {

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

func RunSearchProducts(log comm.Logger, appCtx context.Context, comportName string) error {

	return workgui.RunWork(log, appCtx, "поиск приборов сети", func(log *structlog.Logger, ctx context.Context) error {
		party, err := data.GetCurrentParty()
		if err != nil {
			return err
		}
		device, err := config.Get().Hardware.GetDevice(party.DeviceType)
		if err != nil {
			return err
		}

		if len(device.Params) == 0 {
			return merry.Errorf("нет параметров устройства %q", party.DeviceType)
		}

		cm := comm.New(comports.GetComport(comportName, device.Baud), comm.Config{
			TimeoutGetResponse: 500 * time.Millisecond,
			TimeoutEndResponse: 50 * time.Millisecond,
		})

		ans, notAns := make(intrng.Bytes), make(intrng.Bytes)
		param := device.Params[0]

		go gui.NotifyProgressShow(127, "модбас: сканирование сети")
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
	})
}

func pause(chDone <-chan struct{}, d time.Duration) {
	timer := time.NewTimer(d)
	for {
		select {
		case <-timer.C:
			return
		case <-chDone:
			timer.Stop()
			return
		}
	}
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
