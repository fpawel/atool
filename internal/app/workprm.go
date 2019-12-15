package app

import (
	"context"
	"encoding/binary"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/gui"
	"github.com/fpawel/atool/internal/cfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/comm/modbus"
	"math"
	"strconv"
	"time"
)

type paramsReader struct {
	p   data.Product
	rdr modbus.ResponseReader
	dt  []byte
	rd  []bool
}

func newParamsReader(ctx context.Context, product data.Product, device cfg.Device) (paramsReader, error) {
	rdr, err := getResponseReader(ctx, product.Comport, device)
	if err != nil {
		return paramsReader{}, nil
	}
	r := paramsReader{
		p:   product,
		rdr: rdr,
		dt:  make([]byte, device.BufferSize()),
		rd:  make([]bool, device.BufferSize()),
	}
	for i := range r.dt {
		r.dt[i] = 0xFF
	}
	return r, nil
}

func (r paramsReader) read(prm cfg.Params) error {
	startTime := time.Now()
	request, response, err := r.getResponse(prm)
	if merry.Is(err, context.Canceled) {
		return err
	}
	ct := gui.CommTransaction{
		Addr:     r.p.Addr,
		Comport:  r.p.Comport,
		Request:  formatBytes(request),
		Response: formatBytes(response),
		Duration: time.Since(startTime).String(),
		Ok:       err == nil,
	}
	if err != nil {
		ct.Response = err.Error()
	}
	go gui.NotifyNewCommTransaction(ct)
	return err
}

func (r paramsReader) getResponse(prm cfg.Params) ([]byte, []byte, error) {
	regsCount := prm.Count * 2
	bytesCount := regsCount * 2

	req := modbus.RequestRead3(r.p.Addr, modbus.Var(prm.ParamAddr), uint16(regsCount))
	response, err := req.GetResponse(log, r.rdr, func(request, response []byte) (s string, e error) {
		if len(response) != bytesCount+5 {
			return "", merry.Errorf("длина ответа %d не равна %d", len(response), bytesCount+5)
		}
		return "", nil
	})
	if err == nil {
		offset := 2 * prm.ParamAddr
		copy(r.dt[offset:], response[3:][:bytesCount])
		for i := 0; i < bytesCount; i++ {
			r.rd[offset+i] = true
		}
	}
	return req.Bytes(), response, err
}

func (r paramsReader) processParamValueRead(p cfg.Params, i int, ms *measurements) {
	paramAddr := p.ParamAddr + 2*i
	offset := 2 * paramAddr
	if !r.rd[offset] {
		return
	}
	d := r.dt[offset:]

	setValue := func(value float64, str string) {
		if !math.IsNaN(value) {
			ms.add(r.p.ProductID, p.ParamAddr+2*i, value)
		}
		go gui.NotifyNewProductParamValue(gui.ProductParamValue{
			Addr:      r.p.Addr,
			Comport:   r.p.Comport,
			ParamAddr: p.ParamAddr + 2*i,
			Value:     str,
		})
	}

	setBytesNaN := func() {
		setValue(math.NaN(), formatBytes(d))
	}

	setFloat := func(x float64) {
		if math.IsNaN(x) {
			setBytesNaN()
			return
		}
		setValue(x, strconv.FormatFloat(x, 'f', -1, 64))
	}

	setFloatBits := func(endian binary.ByteOrder) {
		bits := endian.Uint32(d)
		setFloat(float64(math.Float32frombits(bits)))
	}
	setIntBits := func(endian binary.ByteOrder) {
		bits := endian.Uint32(d)
		setValue(float64(int32(bits)), strconv.Itoa(int(bits)))
	}

	var (
		be = binary.BigEndian
		le = binary.LittleEndian
	)

	switch p.Format {
	case "bcd":
		if x, ok := modbus.ParseBCD6(d); ok {
			setFloat(x)
		} else {
			setBytesNaN()
		}
	case "float_big_endian":
		setFloatBits(be)
	case "float_little_endian":
		setFloatBits(le)
	case "int_big_endian":
		setIntBits(be)
	case "int_little_endian":
		setIntBits(le)
	default:
		panic(p.Format)
	}
}
