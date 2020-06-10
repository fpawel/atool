package app

import (
	"context"
	"github.com/fpawel/atool/internal/hardware"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/atool/internal/worklua"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"path/filepath"
)

type scriptSvc struct {
}

var _ api.ScriptService = new(scriptSvc)

func (x *scriptSvc) RunFile(_ context.Context, filename string) error {
	luaState := lua.NewState()
	imp := worklua.NewImport(log, luaState)
	luaState.SetGlobal("go", luar.New(luaState, imp))
	return workgui.RunWork(log, appCtx, filepath.Base(filename), func(log logger, ctx context.Context) error {
		defer hardware.CloseHardware(log, appCtx)
		defer luaState.Close()
		luaState.SetContext(ctx)
		return luaState.DoFile(filename)
	})
}
