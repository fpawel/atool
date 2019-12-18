package app

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/gui"
	"github.com/fpawel/atool/internal/cfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/comm/modbus"
	"math"
)

type paramsReader struct {
	p   data.Product
	rdr modbus.ResponseReader
	dt  []byte
	rd  []bool
	dv  cfg.Device
}

func newParamsReader(product data.Product, device cfg.Device) (paramsReader, error) {
	rdr, err := wrk.getResponseReader(product.Comport, device)
	if err != nil {
		return paramsReader{}, nil
	}
	r := paramsReader{
		p:   product,
		rdr: rdr,
		dt:  make([]byte, device.BufferSize()),
		rd:  make([]bool, device.BufferSize()),
		dv:  device,
	}
	for i := range r.dt {
		r.dt[i] = 0xFF
	}
	return r, nil
}

func (r paramsReader) getResponse(ctx context.Context, prm cfg.Params) error {

	regsCount := prm.Count * 2
	bytesCount := regsCount * 2

	req := modbus.RequestRead3(r.p.Addr, modbus.Var(prm.ParamAddr), uint16(regsCount))

	ct := commTransaction{
		req:         req,
		device:      r.dv,
		comportName: r.p.Comport,
		prs: func(request, response []byte) (s string, e error) {
			if len(response) != bytesCount+5 {
				return "", merry.Errorf("длина ответа %d не равна %d", len(response), bytesCount+5)
			}
			return "", nil
		},
	}
	response, err := ct.getResponse(log, ctx)
	if err == nil {
		offset := 2 * prm.ParamAddr
		copy(r.dt[offset:], response[3:][:bytesCount])
		for i := 0; i < bytesCount; i++ {
			r.rd[offset+i] = true
		}
	}

	return err
}

func (r paramsReader) processParamValueRead(p cfg.Params, i int, ms *measurements) {
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
