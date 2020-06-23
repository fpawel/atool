package workparty

import (
	"context"
	"errors"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config/appcfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg/comports"
	"github.com/fpawel/atool/internal/pkg/intrng"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/hashicorp/go-multierror"
	"time"
)

type Work = workgui.Work

func Delay(duration time.Duration, name string) workgui.WorkFunc {
	// измерения, полученные в процесе опроса приборов во время данной задержки
	ms := new(data.MeasurementCache)
	defer ms.Save()
	return workgui.Delay(duration, name, ReadProductsParams(ms, nil))
}

func NewWorkInterrogate() Work {
	return workgui.New("📤 опрос приборов", func(log comm.Logger, ctx context.Context) error {
		ms := new(data.MeasurementCache)
		defer ms.Save()
		errorsOccurred := make(ErrorsOccurred)
		for {
			if err := ReadProductsParams(ms, errorsOccurred)(log, ctx); err != nil {
				if merry.Is(err, context.Canceled) {
					return nil
				}
				return err
			}
		}
	})
}

func NewWorkReadAllCfs() Work {
	return workgui.New("📤 считывание коэффициентов", func(log comm.Logger, ctx context.Context) error {
		errs := make(ErrorsOccurred)
		err := ProcessEachActiveProduct(errs, func(log comm.Logger, ctx context.Context, p Product) error {
			return p.readAllCoefficients(log, ctx)
		})(log, ctx)
		if err != nil {
			return err
		}
		if len(errs) > 0 {
			return errors.New("не все коэффициенты считаны")
		}
		return nil
	})
}

func NewWorkWriteAllCfs(in []*apitypes.ProductCoefficientValue) Work {
	return workgui.New("запись коэффициентов", func(log comm.Logger, ctx context.Context) error {
		var (
			mulErr *multierror.Error
			xs     []ProductCoefficientValue
		)
		for _, p := range in {
			xs = append(xs, ProductCoefficientValue{
				ProductID:   p.ProductID,
				Coefficient: modbus.Coefficient(p.Coefficient),
				Value:       p.Value,
			})
		}
		//return merry.New("не все коэффициенты записаны")
		if err := WriteProdsCfs(xs, func(v ProductCoefficientValue, err error) error {
			mulErr = multierror.Append(mulErr, err)
			return nil
		})(log, ctx); err != nil {
			return err
		}
		return mulErr
	})
}

func NewWorkRawCmd(c modbus.ProtoCmd, b []byte) Work {
	what := fmt.Sprintf("📥 отправка команды %X(% X)", c, b)
	return Work{
		Name: what,
		Func: ProcessEachActiveProduct(nil, func(log comm.Logger, ctx context.Context, p Product) error {
			_, err := modbus.Request{
				Addr:     p.Addr,
				ProtoCmd: c,
				Data:     b,
			}.GetResponse(log, ctx, p.Comm())
			if err != nil {
				return merry.Prepend(err, what)
			}
			workgui.NotifyInfo(log, fmt.Sprintf("%s %s - успешно", p, what))
			return nil
		}),
	}
}

func SetNetAddr(productID int64, notifyComm func(comm.Info)) workgui.WorkFunc {
	return func(log comm.Logger, ctx context.Context) error {
		var p data.Product
		err := data.DB.Get(&p, `SELECT * FROM product WHERE product_id=?`, productID)
		if err != nil {
			return err
		}

		party, err := data.GetCurrentParty()
		if err != nil {
			return err
		}

		device, err := appcfg.Cfg.Hardware.GetDevice(party.DeviceType)
		if err != nil {
			return err
		}

		workProduct := Product{
			Product: p,
			Device:  device,
			Party:   party,
		}

		return Work{
			Name: fmt.Sprintf("%s: запись сетевого адреса %d", workProduct, p.Addr),
			Func: func(log comm.Logger, ctx context.Context) error {
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
				req := modbus.RequestRead3{
					Addr:           p.Addr,
					FirstRegister:  0,
					RegistersCount: 2,
				}
				answer, err := req.GetResponse(log, ctx, getCommProduct(p.Comport, device))
				if err == nil {
					workgui.NotifyInfo(log, fmt.Sprintf("установка сетевого адреса: % 02X -> % 02X", req.Request().Bytes(), answer))
				} else {
					workgui.NotifyErr(log, fmt.Errorf("установка сетевого адреса: % 02X: %v", req.Request().Bytes(), err))
				}
				return err
			},
		}.Run(log, ctx)
	}
}

func NewWorkScanModbus(comportName string) workgui.Work {
	return workgui.New("сканирование сети модбас", func(log comm.Logger, ctx context.Context) error {
		party, err := data.GetCurrentParty()
		if err != nil {
			return err
		}
		device, err := appcfg.Cfg.Hardware.GetDevice(party.DeviceType)
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
