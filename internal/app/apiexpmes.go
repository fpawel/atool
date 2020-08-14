package app

import (
	"context"
	"fmt"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/pkg/winapi"
	"github.com/fpawel/comm/modbus"
	"os"
	"os/exec"
	"time"
)

type exportMeasurementsAsTextServiceSvc struct{}

func (exportMeasurementsAsTextServiceSvc) Export(_ context.Context, tmFrom unixMillis, tmTo unixMillis, filename string) error {

	timeFrom := unixMillisToTime(tmFrom)
	timeTo := unixMillisToTime(tmTo)

	var xs []struct {
		Tm        int64       `db:"tm"`
		ProductID int64       `db:"product_id"`
		Addr      modbus.Addr `db:"addr"`
		Serial    int         `db:"serial"`
		ParamAddr modbus.Var  `db:"param_addr"`
		Value     float64     `db:"value"`
	}

	const q = `
SELECT tm, product_id, addr, serial, param_addr, value FROM measurement
INNER JOIN product USING (product_id)
WHERE tm BETWEEN ? AND ?
ORDER BY created_at, created_order, param_addr, tm 

`
	if err := data.DB.Select(&xs, q, timeFrom, timeTo); err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}

	if _, err := f.WriteString("Дата,ID,Адресс,Сер.№,Регистр,Значение\n"); err != nil {
		return err
	}

	for _, x := range xs {
		_, err := fmt.Fprintf(f, "%s,%d,%d,%d,%d,%v\n",
			time.Unix(0, x.Tm).Format("2006-01-02 15:04:05.000"),
			x.ProductID,
			x.Addr,
			x.Serial, x.ParamAddr, x.Value)
		if err != nil {
			return err
		}
	}
	log.ErrIfFail(f.Close)

	cmd := exec.Command("./npp/notepad++.exe", filename)
	if err := cmd.Start(); err != nil {
		return err
	}
	winapi.ActivateWindowByPid(cmd.Process.Pid)

	return nil
}
