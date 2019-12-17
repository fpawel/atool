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

func runWork(what string, work func(context.Context) (string, error)) {
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
		go gui.Popup(false, what+": выполняется")

		result, err := work(ctx)
		if err != nil {
			go gui.PopupError(false, merry.Append(err, what))
		} else {
			if len(what) == 0 {
				gui.Popup(false, what+": "+result)
				return
			}
			go gui.Popup(false, what+": выполнено")
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

func runTask(what string, task func() (string, error)) {
	go func() {
		gui.Popup(false, what+": выполняется")
		str, err := task()
		if err != nil {
			gui.PopupError(false, merry.Append(err, what))
			return
		}
		if len(what) == 0 {
			gui.Popup(false, what+": "+str)
			return
		}
		gui.Popup(false, what+": выполнено")

	}()
}

func processEachActiveProduct(work func(data.Product, cfg.Device) error) error {
	products, err := getActiveProducts()
	if err != nil {
		return err
	}
	c := cfg.Get()
	for _, p := range products {
		d, f := c.Hardware.DeviceByName(p.Device)
		if !f {
			return fmt.Errorf("не заданы параметры устройства %s для прибора %+v",
				p.Device, p)
		}
		go gui.Popup(false, fmt.Sprintf("опрашивается прибор: %s %s адр.%d", d.Name, p.Comport, p.Addr))
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

func getResponseReader(comportName string, device cfg.Device) (modbus.ResponseReader, error) {
	port, err := getComport(comport.Config{
		Name:        comportName,
		Baud:        device.Baud,
		ReadTimeout: time.Millisecond,
	})
	if err != nil {
		return nil, merry.Append(err, "не удалось открыть СОМ порт")
	}
	return modbus.NewResponseReader(port, device.CommConfig()), nil
}

func getComport(c comport.Config) (*comport.Port, error) {
	if p, f := comports[c.Name]; f {
		if err := p.SetConfig(c); err != nil {
			return nil, err
		}
		return p, nil
	}
	comports[c.Name] = comport.NewPort(c)
	return comports[c.Name], nil
}
