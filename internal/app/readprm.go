package app

import (
	"context"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/comm/modbus"
	"math"
)

type paramsReader struct {
	p  data.Product
	dt []byte
	rd []bool
	dv config.Device
}

func newParamsReader(product data.Product, device config.Device) paramsReader {
	r := paramsReader{
		p:  product,
		dt: make([]byte, device.BufferSize()),
		rd: make([]bool, device.BufferSize()),
		dv: device,
	}
	for i := range r.dt {
		r.dt[i] = 0xFF
	}
	return r
}

func (r paramsReader) getResponse(ctx context.Context, prm config.Params) error {

	regsCount := prm.Count * 2
	bytesCount := regsCount * 2

	req3 := modbus.RequestRead3{
		Addr:           r.p.Addr,
		FirstRegister:  modbus.Var(prm.ParamAddr),
		RegistersCount: uint16(regsCount),
	}
	cm := getCommProduct(r.p.Comport, r.dv)
	response, err := req3.GetResponse(log, ctx, cm)
	if err == nil {
		offset := 2 * prm.ParamAddr
		copy(r.dt[offset:], response[3:][:bytesCount])
		for i := 0; i < bytesCount; i++ {
			r.rd[offset+i] = true
		}
	}
	return err
}

func (r paramsReader) processParamValueRead(p config.Params, i int, ms *measurements) {
	paramAddr := p.ParamAddr + 2*i
	offset := 2 * paramAddr
	if !r.rd[offset] {
		return
	}
	d := r.dt[offset:]

	ct := gui.ProductParamValue{
		Addr:      r.p.Addr,
		Comport:   r.p.Comport,
		ParamAddr: p.ParamAddr + 2*i,
	}
	if v, err := p.Format.ParseFloat(d); err == nil {
		ct.Value = formatFloat(v)
		if !math.IsNaN(v) {
			ms.add(r.p.ProductID, p.ParamAddr+2*i, v)
		}
	} else {
		ct.Value = err.Error()
	}
	go gui.NotifyNewProductParamValue(ct)
}
