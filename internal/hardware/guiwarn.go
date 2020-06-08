package hardware

import (
	"context"
	"fmt"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/atool/internal/workparty"
	"github.com/fpawel/comm"
)

type GuiWarn struct{}

func (_ GuiWarn) HoldTemperature(log comm.Logger, ctx context.Context, destinationTemperature float64) error {
	err := TemperatureSetup(log, ctx, destinationTemperature)
	if err = workgui.WithWarn(log, ctx, err); err != nil {
		return err
	}
	return workparty.Delay(log, ctx, config.Get().Temperature.HoldDuration,
		fmt.Sprintf("выдержка на температуре %v⁰C", destinationTemperature))
}

func (_ GuiWarn) BlowGas(log comm.Logger, ctx context.Context, gas byte) error {
	return workgui.Perform(log, ctx, fmt.Sprintf("продуть газ %d", gas), func(log comm.Logger, ctx context.Context) error {
		if state.blowGas == gas {
			workgui.NotifyInfo(log, "продувка выполнена ранее")
			return nil
		}
		if err := workgui.WithWarn(log, ctx, SwitchGas(log, ctx, gas)); err != nil {
			return err
		}
		c := config.Get().Gas
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
