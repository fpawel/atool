package app

import (
	"context"
	"database/sql"
	"encoding/binary"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/gui"
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
		setComportLog()
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

	bufferSize := getResponseBufferSize()
	if bufferSize == 0 {
		return errNoInterrogateObjects
	}
	for _, productID := range getCurrentPartyProductsIDs() {
		if err := processProduct(ctx, productID, bufferSize, ms); err != nil {
			return err
		}
	}
	return nil
}

func processProduct(ctx context.Context, productID int64, bufferSize int, ms *measurements) error {
	params := getReadParams(productID)

	data := make([]byte, bufferSize)
	for i := range data {
		data[i] = 0xFF
	}

	for _, p := range params {
		if p.SizeRead == 0 {
			continue
		}
		response, err := p.getResponse3(ctx)
		if err != nil {
			if merry.Is(err, context.Canceled) {
				return err
			}
			if len(response) > 0 {
				err = merry.Appendf(err, "% X", response)
			}
			go gui.NotifyNewCommTransaction(gui.CommTransaction{
				Addr:    p.Addr,
				Comport: p.Comport,
				Request: formatBytes(p.request3().Bytes()),
				Result:  err.Error(),
				Ok:      false,
			})
			continue
		}
		copy(data[p.ParamAddr:2*p.SizeRead], response[3:3+2*p.SizeRead])

		go gui.NotifyNewCommTransaction(gui.CommTransaction{
			Addr:    p.Addr,
			Comport: p.Comport,
			Request: formatBytes(p.request3().Bytes()),
			Result:  formatBytes(response),
			Ok:      true,
		})
	}

	for _, p := range params {
		p.processValueRead(data, ms)
	}
	return nil
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

func getResponseBufferSize() int {
	var bufferSize *int
	err := db.Get(&bufferSize, `
SELECT 2 * max(param_addr + size_read)
FROM product INNER JOIN param USING (device)
WHERE party_id = (SELECT party_id FROM app_config)
  AND active
  AND param.size_read > 0`)
	if err == sql.ErrNoRows {
		return 0
	}
	if err != nil {
		panic(err)
	}
	if bufferSize == nil {
		return 0
	}
	return *bufferSize
}

func getCurrentPartyProductsIDs() (productIDs []int64) {
	err := db.Select(&productIDs,
		`SELECT product_id FROM product WHERE party_id = (SELECT party_id FROM app_config) AND active`)
	if err != nil {
		panic(err)
	}
	return
}

type readParam struct {
	ProductID          int64         `db:"product_id"`
	Comport            string        `db:"comport"`
	Addr               modbus.Addr   `db:"addr"`
	Device             string        `db:"device"`
	Baud               int           `db:"baud"`
	Pause              time.Duration `db:"pause"`
	TimeoutGetResponse time.Duration `db:"timeout_get_responses"`
	TimeoutEndResponse time.Duration `db:"timeout_end_response"`
	MaxAttemptsRead    int           `db:"max_attempts_read"`
	ParamAddr          modbus.Var    `db:"param_addr"`
	Format             string        `db:"format"`
	SizeRead           uint16        `db:"size_read"`
}

func (p readParam) getResponse3(ctx context.Context) ([]byte, error) {
	rdr, err := p.getResponseReader(ctx)
	if err != nil {
		return nil, err
	}
	return p.request3().GetResponse(log, rdr, func(request, response []byte) (s string, e error) {
		lenMustBe := int(p.SizeRead)*2 + 5
		if len(response) != lenMustBe {
			return "", merry.Errorf("длина ответа %d не равна %d", len(response), lenMustBe)
		}
		return "", nil
	})
}

func (p readParam) getResponseReader(ctx context.Context) (modbus.ResponseReader, error) {
	port := getComport(p.Comport)
	if err := port.SetConfig(comport.Config{
		Name:        p.Comport,
		Baud:        p.Baud,
		ReadTimeout: time.Millisecond,
	}); err != nil {
		return nil, merry.Append(err, "не удалось открыть СОМ порт")
	}
	return port.NewResponseReader(ctx, comm.Config{
		TimeoutGetResponse: p.TimeoutGetResponse,
		TimeoutEndResponse: p.TimeoutEndResponse,
		MaxAttemptsRead:    p.MaxAttemptsRead,
		Pause:              p.Pause,
	}), nil
}

func (p readParam) request3() modbus.Request {
	return modbus.RequestRead3(p.Addr, p.ParamAddr, p.SizeRead)
}

func (p readParam) processValueRead(d []byte, ms *measurements) {
	if int(p.ParamAddr*2) >= len(d) {
		return
	}
	v := gui.ProductParamValue{
		Addr:      p.Addr,
		Comport:   p.Comport,
		ParamAddr: p.ParamAddr,
	}
	d = d[p.ParamAddr*2 : p.ParamAddr*2+4]
	value := math.NaN()
	switch p.Format {
	case "bcd":
		if x, ok := modbus.ParseBCD6(d); ok {
			value = x
			v.Value = strconv.FormatFloat(x, 'f', -1, 64)
		}
	case "float_big_endian":
		bits := binary.BigEndian.Uint32(d[:4])
		value = float64(math.Float32frombits(bits))
		v.Value = strconv.FormatFloat(value, 'f', -1, 64)
	case "float_little_endian":
		bits := binary.LittleEndian.Uint32(d[:4])
		value = float64(math.Float32frombits(bits))
		v.Value = strconv.FormatFloat(value, 'f', -1, 64)
	case "int_big_endian":
		bits := binary.BigEndian.Uint32(d[:4])
		value = float64(int32(bits))
		v.Value = strconv.Itoa(int(int32(bits)))
	case "int_little_endian":
		bits := binary.LittleEndian.Uint32(d[:4])
		value = float64(int32(bits))
		v.Value = strconv.Itoa(int(int32(bits)))
	default:
		panic(p.Format)
	}
	go gui.NotifyNewProductParamValue(v)
	if !math.IsNaN(value) {
		ms.add(p.ProductID, p.ParamAddr, value)
	}

}

func formatBytes(xs []byte) string {
	return fmt.Sprintf("% X", xs)
}

func getReadParams(productID int64) (params []readParam) {
	err := db.Select(&params, `
SELECT product_id, comport, addr, device, baud, pause, timeout_get_responses, timeout_end_response, max_attempts_read,       
       param_addr, size_read, format
FROM product
         INNER JOIN hardware USING (device)
         INNER JOIN param USING (device)
WHERE product_id = ?`, productID)
	if err != nil {
		panic(err)
	}
	return
}

func setComportLog() {
	var logComport bool
	if err := db.Get(&logComport, `SELECT log_comport FROM app_config WHERE id=1`); err != nil {
		panic(merry.Append(err, "get config from db"))
	}
	comm.SetEnableLog(logComport)
}

var errNoInterrogateObjects = merry.New("не установлены объекты опроса")
