package app

import (
	"bufio"
	"context"
	"errors"
	"github.com/fpawel/atool/internal/guiwork"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"os"
	"path/filepath"
	"strings"
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
	luaState := lua.NewState()
	luaState.SetGlobal("go", luar.New(luaState, &luaImport{luaState: luaState}))
	return guiwork.RunWork(log, appCtx, filepath.Base(filename), func(log logger, ctx context.Context) error {
		defer luaState.Close()
		luaState.SetContext(ctx)
		return luaState.DoFile(filename)
	})
}
func (_ *scriptSvc) Report(_ context.Context, filename string) error {
	luaState := lua.NewState()
	luaState.SetGlobal("go", luar.New(luaState, &luaImport{luaState: luaState}))
	return guiwork.RunWork(log, appCtx, filepath.Base(filename), func(log logger, ctx context.Context) error {
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

func getScriptType(filename string) (scriptType, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer log.ErrIfFail(file.Close)
	sc := bufio.NewScanner(file)
	if !sc.Scan() {
		return 0, errors.New("no lines")
	}
	if sc.Err() != nil {
		return 0, err
	}
	words := strings.Split(sc.Text(), ":")
	if len(words) < 3 {
		return scriptTypeUnknown, nil
	}
	switch strings.ToLower(strings.TrimSpace(words[1])) {
	case "work":
		return scriptTypeWork, nil
	case "report":
		return scriptTypeReport, nil
	default:
		return scriptTypeUnknown, nil
	}

}

type scriptType int

const (
	scriptTypeUnknown scriptType = iota
	scriptTypeWork
	scriptTypeReport
)
