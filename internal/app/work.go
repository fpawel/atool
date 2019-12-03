package app

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/gui"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/comm/modbus"
	"github.com/lxn/win"
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
		setComportLog()
		for {
			if err := processProductsParams(ctx); err != nil {
				if !merry.Is(err, context.Canceled) {
					go gui.MsgBox("Опрос", err.Error(), win.MB_OK|win.MB_ICONWARNING)
				}
				break
			}
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

func processProductsParams(ctx context.Context) error {
	bufferSize := getResponseBufferSize()
	if bufferSize == 0 {
		return errNoInterrogateObjects
	}
	for _, productID := range getCurrentPartyProductsIDs() {
		if err := processProduct(ctx, productID, bufferSize); err != nil {
			return err
		}
	}
	return nil
}

func processProduct(ctx context.Context, productID int64, bufferSize int) error {
	params := getReadParams(productID)

	data := make([]byte, bufferSize)

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
			go gui.NotifyCommTransaction(gui.CommTransaction{
				Comport: p.Comport,
				Request: formatBytes(p.request3().Bytes()),
				Result:  err.Error(),
				Ok:      false,
			})
			continue
		}
		copy(data[p.ParamAddr:2*p.SizeRead], response[3:3+2*p.SizeRead])

		go gui.NotifyCommTransaction(gui.CommTransaction{
			Comport: p.Comport,
			Request: formatBytes(p.request3().Bytes()),
			Result:  formatBytes(response),
			Ok:      true,
		})
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
	ReadOnce           bool          `db:"read_once"`
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

func formatBytes(xs []byte) string {
	return fmt.Sprintf("% X", xs)
}

func getReadParams(productID int64) (params []readParam) {
	err := db.Select(&params, `
SELECT comport, addr, device, baud, pause, timeout_get_responses, timeout_end_response, max_attempts_read,       
       param_addr, size_read, read_once, format
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
