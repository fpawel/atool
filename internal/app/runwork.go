package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/gui"
	"github.com/fpawel/atool/internal/cfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/comm/modbus"
	"sync"
	"sync/atomic"
	"time"
)

var (
	comports        = map[string]*comport.Port{}
	atomicConnected int32
	disconnect      = func() {}
	wgConnect       sync.WaitGroup
)

var errNoInterrogateObjects = merry.New("не установлены объекты опроса")

func connected() bool {
	return atomic.LoadInt32(&atomicConnected) != 0
}

func runWork(work func(context.Context) error) {
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

func processEachActiveProduct(work func(data.Product, cfg.Device) error) error {
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
