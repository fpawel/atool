package hardware

import (
	"context"
	"fmt"
	"github.com/fpawel/atool/internal/config/appcfg"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/atool/internal/workparty"
	"github.com/fpawel/comm"
)

type WithWarn struct{}

func (_ WithWarn) HoldTemperature(destinationTemperature float64) workgui.WorkFunc {
	return workgui.NewWorkFuncList(
		TemperatureSetup(destinationTemperature).DoWarn,
		workparty.Delay(appcfg.Cfg.Temperature.HoldDuration,
			fmt.Sprintf("выдержка на температуре %v⁰C", destinationTemperature)),
	).Do
}

func (_ WithWarn) BlowGas(gas byte) workgui.WorkFunc {
	return func(log comm.Logger, ctx context.Context) error {
		if state.blowGas == gas {
			workgui.NotifyInfo(log, "продувка выполнена ранее")
			return nil
		}
		c := appcfg.Cfg.Gas
		blowDuration := c.BlowDuration[0]
		if gas-1 < 6 {
			blowDuration = c.BlowDuration[gas-1]
		}
		return workgui.NewWorkFuncList(
			SwitchGas(gas).DoWarn,
			workparty.Delay(blowDuration, fmt.Sprintf("продуть газ %d", gas)),
			func(comm.Logger, context.Context) error {
				state.blowGas = gas
				return nil
			},
		).Do(log, ctx)
	}
}
