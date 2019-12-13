package app

import (
	"context"
	"database/sql"
	"encoding/binary"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/gui"
	"github.com/fpawel/atool/internal/cfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/comm/modbus"
	"math"
	"strconv"
	"sync/atomic"
	"time"
)

func connected() bool {
	return atomic.LoadInt32(&atomicConnected) != 0
}

func runInterrogate() {
	run(func(ctx context.Context) error {
		must.PanicIf(createNewChartIfUpdatedTooLong())
		ms := new(measurements)
		defer func() {
			saveMeasurements(ms.xs)
		}()
		for {
			if err := processProductsParams(ctx, ms); err != nil {
				if merry.Is(err, context.Canceled) {
					return nil
				}
				return err
			}
		}
	})
}

func run(work func(context.Context) error) {
	if connected() {
		log.Debug("connect: connected")
		return
	}

	wgConnect.Add(1)
	atomic.StoreInt32(&atomicConnected, 1)

	ctx, interrupt := context.WithCancel(appCtx)
	disconnect = func() {
		interrupt()
	}
	go func() {

		go gui.NotifyStartWork()

		if err := work(ctx); err != nil {
			go gui.PopupError(err)
		}
		interrupt()
		atomic.StoreInt32(&atomicConnected, 0)

		for _, port := range comports {
			log.ErrIfFail(port.Close)
		}

		wgConnect.Done()
		go gui.NotifyStopWork()
	}()

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

	go gui.Popup("Для наглядности графичеких данных текущего опроса создан новый график.")

	log.Info("copy current party for new chart")
	if err := data.CopyCurrentParty(db); err != nil {
		return err
	}
	gui.NotifyCurrentPartyChanged()

	return nil
}

func processProductsParams(ctx context.Context, ms *measurements) error {
	return processEachActiveProduct(ctx, func(p data.Product, d cfg.Device) error {
		return processProduct(ctx, p, d, ms)
	})
}

func processEachActiveProduct(ctx context.Context, work func(data.Product, cfg.Device) error) error {
	products, err := getActiveProducts()
	if err != nil {
		return nil
	}
	c := cfg.Get()
	for i, p := range products {
		d, f := c.Hardware.DeviceByName(p.Device)
		if !f {
			return fmt.Errorf("не заданы параметры устройства %s для прибора номер %d %+v",
				p.Device, i, p)
		}
		if err := work(p, d); merry.Is(err, context.Canceled) {
			return err
		}
	}
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

func processProduct(ctx context.Context, product data.Product, device cfg.Device, ms *measurements) error {
	rdr, err := newParamsReader(ctx, product, device)
	if err != nil {
		return err
	}
	for _, prm := range device.Params {
		if err := rdr.read(prm); err != nil {
			return err
		}
	}
	for _, p := range device.Params {
		for i := 0; i < p.Count; i++ {
			rdr.processParamValueRead(p, i, ms)
		}
	}
	return nil
}

func getResponseReader(ctx context.Context, comportName string, device cfg.Device) (modbus.ResponseReader, error) {
	port := getComport(comportName)
	err := port.SetConfig(comport.Config{
		Name:        comportName,
		Baud:        device.Baud,
		ReadTimeout: time.Millisecond,
	})
	if err != nil {
		return nil, merry.Append(err, "не удалось открыть СОМ порт")
	}
	return port.NewResponseReader(ctx, device.CommConfig()), nil
}

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

func formatBytes(xs []byte) string {
	return fmt.Sprintf("% X", xs)
}

func getComport(name string) *comport.Port {
	if p, f := comports[name]; f {
		return p
	}
	comports[name] = comport.NewPort(comport.Config{
		Name:        name,
		Baud:        9600,
		ReadTimeout: time.Millisecond,
	})
	return comports[name]
}

var errNoInterrogateObjects = merry.New("не установлены объекты опроса")

func runRawCommand(c modbus.ProtoCmd, b []byte) {
	run(func(ctx context.Context) error {
		return processEachActiveProduct(ctx, func(p data.Product, d cfg.Device) error {
			startTime := time.Now()
			rdr, err := getResponseReader(ctx, p.Comport, d)
			if err != nil {
				return err
			}
			req := modbus.Request{
				Addr:     p.Addr,
				ProtoCmd: c,
				Data:     b,
			}
			response, err := rdr.GetResponse(req.Bytes(), log, nil)
			if merry.Is(err, context.Canceled) {
				return err
			}
			ct := gui.CommTransaction{
				Addr:     p.Addr,
				Comport:  p.Comport,
				Request:  formatBytes(req.Bytes()),
				Response: formatBytes(response),
				Duration: time.Since(startTime).String(),
				Ok:       err == nil,
			}
			if err != nil {
				ct.Response = err.Error()
			}
			go gui.NotifyNewCommTransaction(ct)
			return err
		})
	})

}
