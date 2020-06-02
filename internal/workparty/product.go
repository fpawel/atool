package workparty

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/comports"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"math"
)

type Product struct {
	data.Product
	Party  data.Party
	Device devicecfg.Device
}

func (x Product) String() string {
	return fmt.Sprintf("%s №%d.id%d", x.Party.DeviceType, x.Serial, x.ProductID)
}

func (x Product) Write32(log comm.Logger, ctx context.Context, cmd modbus.DevCmd, format modbus.FloatBitsFormat, value float64) error {
	err := modbus.RequestWrite32{
		Addr:      x.Addr,
		ProtoCmd:  0x10,
		DeviceCmd: cmd,
		Format:    format,
		Value:     value,
	}.GetResponse(log, ctx, x.Comm())

	what := fmt.Sprintf("%s: команда %d(%v)", x, cmd, value)

	if err == nil {
		workgui.NotifyInfo(log, fmt.Sprintf("%s: успешно", what))
	} else {
		workgui.NotifyErr(log, merry.Prepend(err, what))
	}
	return err
}

func (x Product) WriteKef(log comm.Logger, ctx context.Context, kef int, format modbus.FloatBitsFormat, value float64) error {
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
		Coefficient: kef,
	}
	what := fmt.Sprintf("%s: запись K%d=%v", x, kef, value)
	if err == nil {
		kv.Result = config.Get().FormatFloat(value)
		kv.Ok = true
		workgui.NotifyInfo(log, fmt.Sprintf("%s: успешно", what))
	} else {
		err = merry.Prepend(err, what)
		kv.Result = err.Error()
		workgui.NotifyErr(log, err)
		kv.Ok = false
	}
	go gui.NotifyCoefficient(kv)

	return err
}

func (x Product) ReadAndSaveParamValue(log comm.Logger, ctx context.Context, param modbus.Var, format modbus.FloatBitsFormat, dbKey string) error {
	what := fmt.Sprintf("%s: считать рег.%d %s: сохранить %q", x, param, format, dbKey)

	value, err := x.ReadParamValue(log, ctx, param, format)
	if err != nil {
		workgui.NotifyErr(log, merry.Prepend(err, what))
		return nil
	}
	workgui.NotifyInfo(log, fmt.Sprintf("%s: сохранить рег.%d,%s = %v",
		x, param, dbKey, value))
	const query = `
INSERT INTO product_value
VALUES (?, ?, ?)
ON CONFLICT (product_id,key) DO UPDATE
    SET value = ?`
	_, err = data.DB.Exec(query, x.ProductID, dbKey, value, value)
	if err != nil {
		return merry.Prepend(err, what)
	}
	return nil
}

func (x Product) ReadParamValue(log comm.Logger, ctx context.Context, reg modbus.Var, format modbus.FloatBitsFormat) (float64, error) {
	v, err := modbus.Read3Value(log, ctx, x.Comm(), x.Addr, reg, format)
	if err != nil {
		return 0, merry.Prependf(err, "%s: считывание регистра %d %v", reg, format)
	}
	return v, nil
}

func (x Product) ReadKef(log comm.Logger, ctx context.Context, k modbus.Var, format modbus.FloatBitsFormat) (float64, error) {
	v, err := modbus.Read3Value(log, ctx, x.Comm(), x.Addr, 224+2*k, format)
	if err != nil {
		return 0, merry.Prependf(err, "считывание K%d %v", k, format)
	}
	if err := data.SaveProductValue(x.ProductID, data.KeyCoefficient(int(k)), v); err != nil {
		return v, merry.Prependf(err, "считывание K%d %v", k, format)
	}
	return v, nil
}

func (x Product) GetResponse(log comm.Logger, ctx context.Context, cmd modbus.ProtoCmd, data []byte) ([]byte, error) {
	return modbus.Request{
		Addr:     x.Addr,
		ProtoCmd: cmd,
		Data:     data,
	}.GetResponse(log, ctx, x.Comm())
}

func (x Product) Comm() comm.T {
	return comm.New(comports.GetComport(x.Comport, x.Device.Baud), x.Device.CommConfig())
}

func (x Product) readAllCoefficients(log comm.Logger, ctx context.Context) error {
	for _, Kr := range x.Device.Coefficients {
		log := pkg.LogPrependSuffixKeys(log,
			"product", x.Product.String(),
			"range", fmt.Sprintf("%d...%d", Kr.Range[0], Kr.Range[1]))
		for kef := Kr.Range[0]; kef <= Kr.Range[1]; kef++ {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if _, f := config.Get().InactiveCoefficients[kef]; f {
				continue
			}
			value, err := modbus.Read3Value(log, ctx, x.Comm(), x.Addr, 224+2*modbus.Var(kef), Kr.Format)
			notifyReadCoefficient(log, x, kef, value, err)
			if err != nil {
				if merry.Is(err, context.DeadlineExceeded) {
					return err
				}
				continue
			}
			// сохранить значение к-та
			if err := data.SaveProductKefValue(x.ProductID, kef, value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (x Product) readParams(log comm.Logger, ctx context.Context, ms *data.MeasurementCache) error {
	rdr := x.newParamsReader()
	for _, prm := range x.Device.Params {
		err := rdr.getResponse(log, ctx, prm)
		if err != nil {
			return err
		}
	}
	for _, p := range x.Device.Params {
		for i := 0; i < p.Count; i++ {
			rdr.processParamValueRead(p, i, ms)
		}
	}
	return nil
}

func (x Product) newParamsReader() productParamsReader {
	r := productParamsReader{
		Product: x,
		dt:      make([]byte, x.Device.BufferSize()),
		rd:      make([]bool, x.Device.BufferSize()),
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

func (r productParamsReader) getResponse(log comm.Logger, ctx context.Context, prm devicecfg.Params) error {

	regsCount := prm.Count * 2
	bytesCount := regsCount * 2

	req3 := modbus.RequestRead3{
		Addr:           r.Addr,
		FirstRegister:  modbus.Var(prm.ParamAddr),
		RegistersCount: uint16(regsCount),
	}
	response, err := req3.GetResponse(log, ctx, r.Comm())
	if err == nil {
		offset := 2 * prm.ParamAddr
		copy(r.dt[offset:], response[3:][:bytesCount])
		for i := 0; i < bytesCount; i++ {
			r.rd[offset+i] = true
		}
	}
	return err
}

func (r productParamsReader) processParamValueRead(p devicecfg.Params, i int, ms *data.MeasurementCache) {
	paramAddr := p.ParamAddr + 2*i
	offset := 2 * paramAddr
	if !r.rd[offset] {
		return
	}
	d := r.dt[offset:]

	ct := gui.ProductParamValue{
		Addr:      r.Addr,
		Comport:   r.Comport,
		ParamAddr: p.ParamAddr + 2*i,
	}
	if v, err := p.Format.ParseFloat(d); err == nil {
		ct.Value = config.Get().FormatFloat(v)
		if !math.IsNaN(v) {
			ms.Add(r.ProductID, p.ParamAddr+2*i, v)
		}
	} else {
		ct.Value = err.Error()
	}
	go gui.NotifyNewProductParamValue(ct)
}

func notifyReadCoefficient(log comm.Logger, p Product, n int, value float64, err error) {
	x := gui.CoefficientValue{
		ProductID:   p.ProductID,
		Read:        true,
		Coefficient: n,
	}

	if err == nil {
		x.Result = config.Get().FormatFloat(value)
		x.Ok = true
		workgui.NotifyInfo(log, fmt.Sprintf("%s: считано K%d=%v", p, n, value))
	} else {
		err = merry.Prependf(err, "%s, считывание K%d", p, n)
		x.Result = err.Error()
		workgui.NotifyErr(log, err)
		x.Ok = false
	}
	go gui.NotifyCoefficient(x)
}
