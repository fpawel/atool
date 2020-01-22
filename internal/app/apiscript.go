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

func (_ *scriptSvc) RunFile(ctx context.Context, filename string) error {
	return guiwork.RunWork(log, appCtx, filepath.Base(filename), func(log *structlog.Logger, ctx context.Context) error {
		return luaDoFile(context.WithValue(ctx, "log", log), filename)
	})
}

func luaDoFile(ctx context.Context, filename string) error {
	L := lua.NewState()
	defer L.Close()
	L.SetContext(ctx)
	if err := luaImportGlobals(L); err != nil {
		return err
	}
	return L.DoFile(filename)
}
