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
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/comm/modbus"
	"github.com/lxn/win"
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

		must.PanicIf(createNewChartIfUpdatedTooLong())

		go gui.NotifyStartWork()
		ms := new(measurements)
		for {
			if err := processProductsParams(ctx, ms); err != nil {
				if !merry.Is(err, context.Canceled) {
					go gui.MsgBox("Опрос", err.Error(), win.MB_OK|win.MB_ICONWARNING)
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

	s := fmt.Sprintf(
		"Опрос текущей партии выполнянлся более часа назад.\n\n%v, %v\n\nДля наглядности графичеких данных текущего опроса создан новый график.",
		t, time.Since(t))

	go gui.MsgBox("atool: создан новый график", s, win.MB_OK|win.MB_ICONINFORMATION)

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
	if err == sql.ErrNoRows {
		return errNoInterrogateObjects
	}
	if err != nil {
		return err
	}
	c := cfg.Get()
	for i, p := range products {
		d, f := c.Hardware.DeviceByName(p.Device)
		if !f {
			return fmt.Errorf("не заданы параметры устройства %s для прибора номер %d %+v",
				p.Device, i, p)
		}
		if err := processProduct(ctx, p, d, ms); err != nil {
			return err
		}
	}
	return nil
}

func processProduct(ctx context.Context, product data.Product, device cfg.Device, ms *measurements) error {
	port := getComport(product.Comport)
	if err := port.SetConfig(comport.Config{
		Name:        product.Comport,
		Baud:        device.Baud,
		ReadTimeout: time.Millisecond,
	}); err != nil {
		return merry.Append(err, "не удалось открыть СОМ порт")
	}
	reader := port.NewResponseReader(ctx, comm.Config{
		TimeoutGetResponse: device.TimeoutGetResponse,
		TimeoutEndResponse: device.TimeoutEndResponse,
		MaxAttemptsRead:    device.MaxAttemptsRead,
		Pause:              device.Pause,
	})

	dataBytes := make([]byte, device.BufferSize())
	readBytes := make([]bool, device.BufferSize())

	for i := range dataBytes {
		dataBytes[i] = 0xFF
	}
	for _, p := range device.Params {

		regsCount := p.Count * 2
		bytesCount := regsCount * 2

		req := modbus.RequestRead3(product.Addr, modbus.Var(p.ParamAddr), uint16(regsCount))

		ct := gui.CommTransaction{
			Addr:    product.Addr,
			Comport: product.Comport,
			Request: formatBytes(req.Bytes()),
		}

		response, err := req.GetResponse(log, reader, func(request, response []byte) (s string, e error) {
			if len(response) != bytesCount+5 {
				return "", merry.Errorf("длина ответа %d не равна %d", len(response), bytesCount+5)
			}
			return "", nil
		})
		if err != nil {

			if len(response) > 0 {
				err = merry.Appendf(err, "% X", response)
			}
			ct.Result = err.Error()
			go gui.NotifyNewCommTransaction(ct)

			if merry.Is(err, context.Canceled) {
				return nil
			}

			continue
		}
		offset := 2 * p.ParamAddr

		copy(dataBytes[offset:], response[3:][:bytesCount])

		for i := 0; i < bytesCount; i++ {
			readBytes[offset+i] = true
		}

		ct.Result = formatBytes(response)
		ct.Ok = true
		go gui.NotifyNewCommTransaction(ct)
	}

	for _, p := range device.Params {
		for i := 0; i < p.Count; i++ {
			paramAddr := p.ParamAddr + 2*i
			offset := 2 * paramAddr
			if !readBytes[offset] {
				continue
			}

			d := dataBytes[offset:]

			v := gui.ProductParamValue{
				Addr:      product.Addr,
				Comport:   product.Comport,
				ParamAddr: paramAddr,
			}

			value := math.NaN()
			switch p.Format {
			case "bcd":
				if x, ok := modbus.ParseBCD6(d); ok {
					value = x
				}
			case "float_big_endian":
				bits := binary.BigEndian.Uint32(d)
				value = float64(math.Float32frombits(bits))
			case "float_little_endian":
				bits := binary.LittleEndian.Uint32(d)
				value = float64(math.Float32frombits(bits))
			case "int_big_endian":
				bits := binary.BigEndian.Uint32(d)
				value = float64(int32(bits))
			case "int_little_endian":
				bits := binary.LittleEndian.Uint32(d)
				value = float64(int32(bits))
			default:
				panic(p.Format)
			}

			if !math.IsNaN(value) {
				ms.add(product.ProductID, paramAddr, value)
				v.Value = strconv.FormatFloat(value, 'f', -1, 64)
			} else {
				v.Value = formatBytes(d)
			}
			go gui.NotifyNewProductParamValue(v)
		}
	}

	return nil
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
