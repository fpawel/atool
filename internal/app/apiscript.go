package app

import (
	"context"
	"github.com/fpawel/atool/internal/guiwork"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/powerman/structlog"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"path/filepath"
)

type scriptSvc struct {
	//paramsRequested []*apitypes.ConfigParamValue
}

var _ api.ScriptService = new(scriptSvc)

func (_ *scriptSvc) RunFile(ctx context.Context, filename string) error {

	L := lua.NewState()
	imp := new(luaImport)
	if err := imp.init(L); err != nil {
		return err
	}
	L.SetGlobal("go", luar.New(L, imp))
	return guiwork.RunWork(log, appCtx, filepath.Base(filename), func(log *structlog.Logger, ctx context.Context) error {
		defer L.Close()
		if err := imp.init(L); err != nil {
			return err
		}
		L.SetContext(context.WithValue(ctx, "log", log))
		return L.DoFile(filename)
	})
}

func (_ *scriptSvc) SetConfigParamValues(_ context.Context, configParamValues []*apitypes.ConfigParamValue) error {
	luaParamValues = configParamValues
	return nil
}

func (_ *scriptSvc) GetConfigParamValues(_ context.Context) ([]*apitypes.ConfigParamValue, error) {
	return luaParamValues, nil
}

var luaParamValues []*apitypes.ConfigParamValue
