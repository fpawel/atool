package app

import (
	"context"
	"fmt"
	"github.com/fpawel/atool/internal/config/appcfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/tealeg/xlsx/v3"
	"os"
	"os/exec"
	"time"
)

type chartPoint struct {
	Tm        int64       `db:"tm"`
	ProductID int64       `db:"product_id"`
	Addr      modbus.Addr `db:"addr"`
	Serial    int         `db:"serial"`
	ParamAddr modbus.Var  `db:"param_addr"`
	Value     float64     `db:"value"`
}

type exportChartFunc = func(filename string, chartPoints []chartPoint) error

func runWorkExportChart(filename string, utmFrom unixMillis, utmTo unixMillis, chart string, exportChartFunc exportChartFunc) error {
	what := fmt.Sprintf("экспорт графика %q в файл %s", chart, filename)
	return runWorkFunc(what, func(log comm.Logger, ctx context.Context) error {
		chartPoints, err := getChartPoints(utmFrom, utmTo, chart)
		if err != nil {
			return err
		}
		if err := exportChartFunc(filename, chartPoints); err != nil {
			return err
		}
		const layout = "01.02-15:04"
		workgui.NotifyInfo(log, fmt.Sprintf("%s: успешно. %s-%s, ", what,
			unixMillisToTime(utmFrom).Format(layout),
			unixMillisToTime(utmTo).Format(layout),
		))

		s1 := fmt.Sprintf("/select,%s", filename)
		//fmt.Println("explorer.exe", s1)
		cmd := exec.Command("explorer.exe", s1)
		if err := cmd.Start(); err != nil {
			return err
		}

		return nil
	})
}

func getChartPoints(utmFrom unixMillis, utmTo unixMillis, chart string) ([]chartPoint, error) {
	qProducts, qParams, err := selectProductParamsChart(chart)
	if err != nil {
		return nil, err
	}

	timeFrom := unixMillisToTime(utmFrom)
	timeTo := unixMillisToTime(utmTo)

	sQ := fmt.Sprintf(`
SELECT tm, product_id, addr, serial, param_addr, value FROM measurement 
INNER JOIN product USING (product_id)
WHERE product_id IN (%s) 
  AND param_addr IN (%s) 
  AND tm >= %d
  AND tm <= %d
ORDER BY product.created_at, product.created_order, param_addr, tm`,
		qProducts, qParams, timeFrom.UnixNano(), timeTo.UnixNano())

	var chartPoints []chartPoint

	if err := data.DB.Select(&chartPoints, sQ); err != nil {
		return nil, err
	}

	return chartPoints, nil
}

func exportChartCsv(filename string, chartPoints []chartPoint) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}

	if _, err := f.WriteString("Дата,Время,ID,Адресс,Сер.№,Регистр,Параметр,Значение\n"); err != nil {
		return err
	}

	party, err := data.GetCurrentParty()
	if err != nil {
		return err
	}

	d, _ := appcfg.Cfg.Hardware.GetDevice(party.DeviceType)

	for _, x := range chartPoints {
		t := time.Unix(0, x.Tm)

		varName, _ := d.VarsNames[x.ParamAddr]

		_, err := fmt.Fprintf(f, "%s,%d,%d,%d,%d,%d,%s,%v\n",
			t.Format("2006-01-02-15:04:05.000"),
			t.Unix(),
			x.ProductID,
			x.Addr,
			x.Serial, x.ParamAddr, varName, x.Value)
		if err != nil {
			return err
		}
	}
	log.ErrIfFail(func() error {
		return f.Close()
	})

	return nil
}

func exportChartXls(filename string, chartPoints []chartPoint) error {

	party, err := data.GetCurrentParty()
	if err != nil {
		return err
	}
	d, _ := appcfg.Cfg.Hardware.GetDevice(party.DeviceType)

	wb := xlsx.NewFile()

	sh, err := wb.AddSheet("график")
	if err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}

	row := sh.AddRow()
	for _, s := range []string{"Дата", "Время", "ID", "Адресс", "Сер.№", "Регистр", "Параметр", "Значение"} {
		row.AddCell().SetValue(s)
	}

	for _, x := range chartPoints {
		t := time.Unix(0, x.Tm)
		varName, _ := d.VarsNames[x.ParamAddr]

		r := sh.AddRow()
		r.AddCell().SetValue(t.Format("2006-01-02-15:04:05.000"))

		appendInt64 := func(v int64) {
			c := r.AddCell()
			c.NumFmt = "0"
			c.SetValue(float64(v))
		}

		appendStr := func(v string) {
			r.AddCell().SetValue(v)
		}

		appendInt64(t.Unix())
		appendInt64(x.ProductID)
		appendInt64(int64(x.Addr))
		appendInt64(int64(x.Serial))
		appendInt64(int64(x.ParamAddr))
		appendStr(varName)

		c := r.AddCell()
		c.NumFmt = "#.##0"
		c.SetValue(x.Value)
	}

	if err := wb.Save(filename); err != nil {
		return err
	}

	sh.Close()

	return nil
}
