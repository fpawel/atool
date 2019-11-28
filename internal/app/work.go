package app

import (
	"context"
	"database/sql"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/gui"
	"github.com/fpawel/atool/internal/data"
	"github.com/lxn/win"
	"sync"
	"sync/atomic"
)

func connected() bool {
	return atomic.LoadInt32(&atomicConnected) != 0
}

func connect() {
	if connected() {
		log.Debug("connect: connected")
		return
	}
	var wgConnect sync.WaitGroup
	wgConnect.Add(1)
	ctx, interrupt := context.WithCancel(context.Background())
	disconnect = func() {
		if !connected() {
			log.Debug("disconnect: disconnected")
			return
		}
		interrupt()
		wgConnect.Wait()
	}
	go func() {
		if err := processProductsParams(ctx); err != nil {
			gui.MsgBox("Опрос", err.Error(), win.MB_OK|win.MB_ICONWARNING)
		}
		interrupt()
		wgConnect.Done()
	}()

}

func processProductsParams(ctx context.Context) error {
	var bufferSize int
	if err := getResponseBufferSize(&bufferSize); err != nil {
		return err
	}
	var products []data.Product
	if err := db.Select(&products,
		`SELECT * FROM product WHERE party_id = (SELECT party_id FROM app_config) AND active`); err != nil {
		return err
	}
	if len(products) == 0 {
		return errNoInterrogateObjects
	}
	for _, product := range products {
		type param struct {
			data.Product
			data.Hardware
			data.Param
		}
		var params []param
		if err := db.Select(&params, `
SELECT product_id,
       comport,
       addr,
       baud,
       timeout_get_responses,
       timeout_end_response,
       pause,
       max_attempts_read,
       device,
       param_addr,
       size_read,
       read_once,
       format
FROM product
         INNER JOIN hardware USING (device)
         INNER JOIN param USING (device)
WHERE product_id = ?  AND active  AND param.size_read > 0`, product.ProductID); err != nil {
			return err
		}
		log.Printf("%+v", params)
	}
	return nil
}

func getResponseBufferSize(bufferSize *int) error {
	err := db.Get(bufferSize, `
SELECT 2 * max(param_addr + size_read)
FROM product INNER JOIN param USING (device)
WHERE party_id = (SELECT party_id FROM app_config)
  AND active
  AND param.size_read > 0`)
	if err != nil {
		if err == sql.ErrNoRows {
			return errNoInterrogateObjects
		}
		return err
	}
	if *bufferSize == 0 {
		return errNoInterrogateObjects
	}
	return nil
}

var errNoInterrogateObjects = merry.New("не установлены объекты опроса")
