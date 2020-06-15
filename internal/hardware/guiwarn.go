package hardware

import (
	"context"
	"fmt"
	"github.com/fpawel/atool/internal/config/appcfg"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/atool/internal/workparty"
	"github.com/fpawel/comm"
	"github.com/powerman/structlog"
)

type GuiWarn struct{}

func (_ GuiWarn) HoldTemperature(destinationTemperature float64) workgui.WorkFunc {
	return func(log *structlog.Logger, ctx context.Context) error {
		err := TemperatureSetup(log, ctx, destinationTemperature)
		if err = workgui.WithWarn(log, ctx, err); err != nil {
			return err
		}
		return workparty.Delay(log, ctx, appcfg.Cfg.Temperature.HoldDuration,
			fmt.Sprintf("выдержка на температуре %v⁰C", destinationTemperature))
	}
}

func (_ GuiWarn) BlowGas(gas byte) workgui.WorkFunc {
	return func(log *structlog.Logger, ctx context.Context) error {
		return workgui.Perform(log, ctx, fmt.Sprintf("продуть газ %d", gas), func(log comm.Logger, ctx context.Context) error {
			if state.blowGas == gas {
				workgui.NotifyInfo(log, "продувка выполнена ранее")
				return nil
			}
			if err := workgui.WithWarn(log, ctx, SwitchGas(log, ctx, gas)); err != nil {
				return err
			}
			c := appcfg.Cfg.Gas
			blowDuration := c.BlowDuration[0]
			if gas-1 < 6 {
				blowDuration = c.BlowDuration[gas-1]
			}
			if err := workparty.Delay(log, ctx, blowDuration, fmt.Sprintf("продуть газ %d", gas)); err != nil {
				return err
			}
			state.blowGas = gas
			return nil
		})
	}
}
