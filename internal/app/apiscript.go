package app

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/atool/internal/worklua"
	"github.com/fpawel/atool/internal/workparty"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"path/filepath"
)

type scriptSvc struct {
	imp *worklua.Import
}

var _ api.ScriptService = new(scriptSvc)

func (x *scriptSvc) SelectWorks(_ context.Context, works []bool) (err error) {
	x.imp.SelectedWorksChan <- works
	return nil
}

func (x *scriptSvc) IgnoreError(context.Context) error {
	x.imp.IgnoreError()
	return nil
}

func (x *scriptSvc) RunFile(_ context.Context, filename string) error {
	luaState := lua.NewState()
	x.imp = worklua.NewImport(log, luaState)
	luaState.SetGlobal("go", luar.New(luaState, x.imp))
	return workgui.RunWork(log, appCtx, filepath.Base(filename), func(log logger, ctx context.Context) error {
		defer closeHardware()
		defer luaState.Close()
		luaState.SetContext(ctx)
		return luaState.DoFile(filename)
	})
}

func (x *scriptSvc) SetConfigParamValues(_ context.Context, configParamValues []*apitypes.ConfigParamValue) error {
	x.imp.ParamValues = configParamValues
	return nil
}

func (x *scriptSvc) GetConfigParamValues(_ context.Context) ([]*apitypes.ConfigParamValue, error) {
	return x.imp.ParamValues, nil
}

func closeHardware() {
	if err := workparty.SwitchGas(log, appCtx, 0); err != nil {
		workgui.NotifyErr(log, merry.Prepend(err, "отключить газ по окончании настройки"))
	} else {
		workgui.NotifyInfo(log, "отключен газ по окончании настройки")
	}

	if err := func() error {
		tempDev, err := workparty.GetTemperatureDevice()
		if err != nil {
			return err
		}
		if tempDev == nil {
			panic("unexpected")
		}
		return tempDev.Stop(log, appCtx)
	}(); err != nil {
		workgui.NotifyErr(log, merry.Prepend(err, "остановить термокамеру по окончании настройки"))
	} else {
		workgui.NotifyInfo(log, "термокамера остановлена по окончании настройки")
	}
}
