package app

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/guiwork"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"path/filepath"
)

type scriptSvc struct{}

var _ api.ScriptService = new(scriptSvc)

func (_ *scriptSvc) SelectWorks(_ context.Context, works []bool) (err error) {
	luaSelectedWorksChan <- works
	return nil
}

func (_ *scriptSvc) IgnoreError(context.Context) error {
	luaIgnoreError()
	return nil
}

func (_ *scriptSvc) RunFile(_ context.Context, filename string) error {
	luaState := lua.NewState()
	luaState.SetGlobal("go", luar.New(luaState, &luaImport{luaState: luaState}))
	return guiwork.RunWork(log, appCtx, filepath.Base(filename), func(log logger, ctx context.Context) error {
		defer closeHardware()
		defer luaState.Close()
		luaState.SetContext(ctx)
		return luaState.DoFile(filename)
	})
}

func (_ *scriptSvc) SetConfigParamValues(_ context.Context, configParamValues []*apitypes.ConfigParamValue) error {
	luaParamValues = configParamValues
	return nil
}

func (_ *scriptSvc) GetConfigParamValues(_ context.Context) ([]*apitypes.ConfigParamValue, error) {
	return luaParamValues, nil
}

func closeHardware() {
	if err := switchGas(appCtx, 0); err != nil {
		guiwork.JournalErr(log, merry.Prepend(err, "отключить газ по окончании настройки"))
	} else {
		guiwork.JournalInfo(log, "отключен газ по окончании настройки")
	}

	if err := func() error {
		tempDev, err := getTemperatureDevice()
		if err != nil {
			return err
		}
		if tempDev == nil {
			panic("unexpected")
		}
		return tempDev.Stop(log, appCtx)
	}(); err != nil {
		guiwork.JournalErr(log, merry.Prepend(err, "остановить термокамеру по окончании настройки"))
	} else {
		guiwork.JournalInfo(log, "термокамера остановлена по окончании настройки")
	}
}
