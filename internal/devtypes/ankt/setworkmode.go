package ankt

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/devtypes/devdata"
	"github.com/fpawel/atool/internal/pkg/comports"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"time"
)

var (
	setWorkModeTime = make(map[int64]time.Time)
)

func onReadProduct(log comm.Logger, ctx context.Context, product devdata.Product) error {
	intervalSec, ok := product.PartyValues[keySendCmdSetWorkModeIntervalSec]
	if !ok || intervalSec == 0 {
		return nil
	}
	t, _ := setWorkModeTime[product.ProductID]
	if time.Since(t) < time.Second*time.Duration(intervalSec) {
		return nil
	}
	err := setProductWorkMode(log, ctx, product, 2)
	if err == nil {
		setWorkModeTime[product.ProductID] = time.Now()
	}
	return err
}

func setProductWorkMode(log comm.Logger, ctx context.Context, product devdata.Product, mode float64) error {
	_, err := modbus.Request{
		Addr:     product.Addr,
		ProtoCmd: 0x16,
		Data:     append([]byte{0xA0, 0, 0, 2, 4}, modbus.BCD6(mode)...),
	}.GetResponse(log, ctx, comports.Comm(product.Comport, deviceConfig))
	return merry.Prepend(err, "установка режима работы 2")
}
