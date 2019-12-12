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

func connect() {
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

		must.PanicIf(createNewChartIfUpdatedTooLong())

		ms := new(measurements)
		for {
			if err := processProductsParams(ctx, ms); err != nil {
				if !merry.Is(err, context.Canceled) {
					go gui.PopupError(err)
				}
				break
			}
		}
		saveMeasurements(ms.xs)
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
	var products []data.Product
	err := db.Select(&products,
		`SELECT * FROM product WHERE party_id = (SELECT party_id FROM app_config) AND active`)
	if err != nil {
		return err
	}
	if len(products) == 0 {
		return errNoInterrogateObjects
	}
	c := cfg.Get()
	for i, p := range products {
		d, f := c.Hardware.DeviceByName(p.Device)
		if !f {
			return fmt.Errorf("не заданы параметры устройства %s для прибора номер %d %+v",
				p.Device, i, p)
		}
		if err := processProduct(ctx, p, d, ms); merry.Is(err, context.Canceled) {
			return err
		}
	}
	return nil
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

type paramsReader struct {
	p   data.Product
	rdr modbus.ResponseReader
	dt  []byte
	rd  []bool
}

func newParamsReader(ctx context.Context, product data.Product, device cfg.Device) (paramsReader, error) {

	port := getComport(product.Comport)
	err := port.SetConfig(comport.Config{
		Name:        product.Comport,
		Baud:        device.Baud,
		ReadTimeout: time.Millisecond,
	})
	if err != nil {
		return paramsReader{}, merry.Append(err, "не удалось открыть СОМ порт")
	}

	r := paramsReader{
		p:   product,
		rdr: port.NewResponseReader(ctx, device.CommConfig()),
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
