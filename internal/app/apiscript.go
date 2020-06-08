package app

import (
	"context"
	"github.com/fpawel/atool/internal/hardware"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/atool/internal/worklua"
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
	workgui.IgnoreError()
	return nil
}

func (x *scriptSvc) RunFile(_ context.Context, filename string) error {
	luaState := lua.NewState()
	x.imp = worklua.NewImport(log, luaState)
	luaState.SetGlobal("go", luar.New(luaState, x.imp))
	return workgui.RunWork(log, appCtx, filepath.Base(filename), func(log logger, ctx context.Context) error {
		defer hardware.CloseHardware(log, appCtx)
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
