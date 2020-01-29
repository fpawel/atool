package app

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/guiwork"
	"github.com/fpawel/atool/internal/journal"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/powerman/structlog"
	lua "github.com/yuin/gopher-lua"
	luajson "layeh.com/gopher-json"
	luar "layeh.com/gopher-luar"
	"path/filepath"
)

type scriptSvc struct{}

var _ api.ScriptService = new(scriptSvc)

func (_ *scriptSvc) IgnoreError(_ context.Context) error {
	luaIgnoreError()
	journal.Err(log, merry.New("Ошибка проигнорирована. Выполнение продолжено."))
	return nil
}

func (_ *scriptSvc) RunFile(_ context.Context, filename string) error {

	L := lua.NewState()
	luajson.Preload(L)
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

var (
	luaParamValues []*apitypes.ConfigParamValue
	luaIgnoreError = func() {}
)
