package app

import (
	"context"
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

func (_ *scriptSvc) IgnoreError(_ context.Context) error {
	luaIgnoreError()

	return nil
}

func (_ *scriptSvc) RunFile(_ context.Context, filename string) error {
	luaState = lua.NewState()
	return guiwork.RunWork(log, appCtx, filepath.Base(filename), func(log logger, ctx context.Context) error {
		defer luaState.Close()
		luaState.SetContext(context.WithValue(ctx, "log", log))

		imp := new(luaImport)
		if err := imp.init(); err != nil {
			return err
		}
		luaState.SetGlobal("go", luar.New(luaState, imp))

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
