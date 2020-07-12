package workparty

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config/appcfg"
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/devtypes/devdata"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"math"
)

type Product struct {
	data.Product
	Party  data.Party
	Device devdata.Device
}

func (x Product) String() string {
	return fmt.Sprintf("🔌%d🔑%d", x.Serial, x.ProductID)
}

func (x Product) Write32(cmd modbus.DevCmd, format modbus.FloatBitsFormat, value float64) workgui.WorkFunc {
	return func(log comm.Logger, ctx context.Context) error {
		what := fmt.Sprintf("%s 📥 команда %d(%v)", x, cmd, value)
		if math.IsNaN(value) {
			return merry.Errorf("%s: NaN", what)
		}
		err := modbus.RequestWrite32{
			Addr:      x.Addr,
			ProtoCmd:  0x10,
			DeviceCmd: cmd,
			Format:    format,
			Value:     value,
		}.GetResponse(log, ctx, x.Comm())
		if err != nil {
			return merry.Prepend(err, what)
		}
		return nil
	}
}

func (x Product) WriteKef(kef devicecfg.Kef, format modbus.FloatBitsFormat, value float64) workgui.WorkFunc {
	return func(log comm.Logger, ctx context.Context) error {
		what := fmt.Sprintf("%s 📥 запись K%d=%v %s", x, kef, value, format)
		err := func() error {
			if math.IsNaN(value) {
				return merry.Errorf("%s: NaN", what)
			}

			err := modbus.RequestWrite32{
				Addr:      x.Addr,
				ProtoCmd:  0x10,
				DeviceCmd: (0x80 << 8) + modbus.DevCmd(kef),
				Format:    format,
				Value:     value,
			}.GetResponse(log, ctx, x.Comm())

			kv := gui.CoefficientValue{
				ProductID:   x.ProductID,
				Read:        false,
				Coefficient: int(kef),
			}
			if err == nil {
				kv.Result = appcfg.Cfg.FormatFloat(value)
				kv.Ok = true
			} else {
				kv.Result = err.Error()
				kv.Ok = false
			}
			go gui.NotifyCoefficient(kv)
			return err
		}()
		if err != nil {
			return merry.Append(err, what)
		}
		workgui.NotifyInfo(log, what+": успешно")
		return nil
	}
}

func (x Product) ReadKef(log comm.Logger, ctx context.Context, k devicecfg.Kef, format modbus.FloatBitsFormat) (float64, error) {
	what := fmt.Sprintf("%s 📥 💾 %s K%d", x, format, k)
	return workgui.WithNotifyValue(log, what, func() (float64, error) {
		value, err := modbus.Read3Value(log, ctx, x.Comm(), x.Addr, 224+2*modbus.Var(k), format)

		i := gui.CoefficientValue{
			ProductID:   x.ProductID,
			Read:        true,
			Coefficient: int(k),
		}

		if err != nil {
			i.Result = err.Error()
			go gui.NotifyCoefficient(i)
			return math.NaN(), err
		}

		i.Ok = true
		i.Result = appcfg.Cfg.FormatFloat(value)

		go gui.NotifyCoefficient(i)

		// сохранить значение к-та
		if err := x.SaveKefValue(k, value); err != nil {
			return math.NaN(), err
		}

		return value, nil
	})
}

func (x Product) SaveKefValue(k devicecfg.Kef, value float64) error {
	return data.SaveProductKefValue(x.ProductID, k, value)
}

func (x Product) Comm() comm.T {
	//return comm.New(comports.GetComport(x.Comport, x.Device.Baud), x.Device.CommConfig()).WithLockPort(x.Comport)
	return getCommProduct(x.Comport, x.Device.Config)
}

func (x Product) readAllCoefficients(log comm.Logger, ctx context.Context) error {
	for _, Kr := range x.Device.Config.CfsList {
		log := pkg.LogPrependSuffixKeys(log,
			"product", x.Product.String(),
			"range", fmt.Sprintf("%d...%d", Kr[0], Kr[1]))
		for kef := Kr[0]; kef <= Kr[1]; kef++ {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if _, f := appcfg.Sets.InactiveCoefficients[kef]; f {
				continue
			}

			_, err := x.ReadKef(log, ctx, kef, x.Device.Config.FloatFormat)
			if err != nil {
				if merry.Is(err, context.DeadlineExceeded) {
					return err
				}
				continue
			}
		}
	}
	return nil
}

func (x Product) readParams(log comm.Logger, ctx context.Context, ms *data.MeasurementCache) error {
	rdr := x.newParamsReader()
	for _, prm := range x.Device.VarsRng(x.Party.ProductType) {
		err := rdr.getResponse(log, ctx, prm)
		if err != nil {
			return err
		}
	}
	for _, p := range x.Device.VarsRng(x.Party.ProductType) {
		for i := modbus.Var(0); i < p[1]; i++ {
			rdr.processParamValueRead(p, i, ms)
		}
	}
	return nil
}

func (x Product) newParamsReader() productParamsReader {
	bufferSize := x.Device.BufferSize(x.Party.ProductType)
	r := productParamsReader{
		Product: x,
		dt:      make([]byte, bufferSize),
		rd:      make([]bool, bufferSize),
	}
	for i := range r.dt {
		r.dt[i] = 0xFF
	}
	return r
}

type productParamsReader struct {
	Product
	dt []byte
	rd []bool
}

func (r productParamsReader) getResponse(log comm.Logger, ctx context.Context, prm devicecfg.Vars) error {

	regsCount := prm.Count() * 2
	bytesCount := regsCount * 2

	req3 := modbus.RequestRead3{
		Addr:           r.Addr,
		FirstRegister:  prm.Var(),
		RegistersCount: uint16(regsCount),
	}
	response, err := req3.GetResponse(log, ctx, r.Comm())
	if err == nil {
		offset := 2 * int(prm.Var())
		copy(r.dt[offset:], response[3:][:bytesCount])
		for i := 0; i < bytesCount; i++ {
			r.rd[offset+i] = true
		}
	}
	return err
}

func (r productParamsReader) processParamValueRead(p devicecfg.Vars, i modbus.Var, ms *data.MeasurementCache) {
	paramAddr := p.Var() + 2*i
	offset := 2 * paramAddr
	if !r.rd[offset] {
		return
	}
	d := r.dt[offset:]

	ct := gui.ProductParamValue{
		Addr:      r.Addr,
		Comport:   r.Comport,
		ParamAddr: p.Var() + 2*i,
	}
	fFmt := r.Device.Config.VarFormat(p.Var())
	if v, err := fFmt.ParseFloat(d); err == nil {
		ct.Value = appcfg.Cfg.FormatFloat(v)
		if !math.IsNaN(v) {
			ms.Add(r.ProductID, p.Var()+2*i, v)
		}
	} else {
		ct.Value = err.Error()
	}
	go gui.NotifyNewProductParamValue(ct)
}
