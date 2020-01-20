package app

import (
	"context"
	"github.com/fpawel/atool/internal/guiwork"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/powerman/structlog"
	lua "github.com/yuin/gopher-lua"
	"path/filepath"
)

type scriptSvc struct{}

var _ api.ScriptService = new(scriptSvc)

func (_ *scriptSvc) RunFile(_ context.Context, filename string) error {
	return guiwork.RunWork(log, appCtx, filepath.Base(filename), func(log *structlog.Logger, ctx context.Context) error {
		L := lua.NewState()
		defer L.Close()
		L.SetContext(ctx)
		L.SetGlobal("pause", L.NewFunction(luaPause(log)))
		L.SetGlobal("gas", L.NewFunction(luaGas(log)))
		L.SetGlobal("temperature", L.NewFunction(luaTemperature(log)))
		L.SetGlobal("read_save", L.NewFunction(luaReadSave(log)))
		if err := luaImportParty(L); err != nil {
			return err
		}
		return L.DoFile(filename)
	})
}
