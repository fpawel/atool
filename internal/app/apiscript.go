package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/gui/guiwork"
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
		return L.DoFile(filename)
	})
}

func (_ *scriptSvc) Run(_ context.Context, nWorks []int32) error {
	L := lua.NewState()

	if err := L.DoFile("works.lua"); err != nil {
		return err
	}
	xs, err := getScriptWorks(L)
	if err != nil {
		return err
	}

	return guiwork.RunWork(log, appCtx, "сценарий", func(log *structlog.Logger, ctx context.Context) error {
		L.SetContext(ctx)

		defer L.Close()

		for _, i := range nWorks {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if i < 0 || int(i) >= len(xs) {
				return fmt.Errorf("index out of range %d", i)
			}
			x := xs[int(i)]
			err := guiwork.PerformNewNamedWork(log, ctx, x.s, func(log *structlog.Logger, ctx context.Context) error {
				return x.w()
			})
			if err != nil {
				return merry.Appendf(err, "%d:%s", i, x.s)
			}
		}
		return nil
	})

}

func (_ *scriptSvc) ListWorksNames(_ context.Context) ([]string, error) {
	L := lua.NewState()
	defer L.Close()

	if err := L.DoFile("works.lua"); err != nil {
		return nil, err
	}
	xs, err := getScriptWorks(L)
	if err != nil {
		return nil, err
	}
	var works []string
	for _, x := range xs {
		works = append(works, x.s)
	}
	return works, nil
}

type scriptWork struct {
	w func() error
	s string
}

func getScriptWorks(L *lua.LState) ([]scriptWork, error) {
	var (
		works []scriptWork
		err   error
	)
	L.GetGlobal("works").(*lua.LTable).ForEach(func(luaIndex lua.LValue, x lua.LValue) {
		if luaIndex.Type() != lua.LTNumber {
			err = merry.Appendf(err, "ключ элемента объекта 'works': %s %+v, ожидался номер\n", luaIndex.Type(), luaIndex)
			return
		}
		n := int(lua.LVAsNumber(luaIndex))
		if n%2 != 0 {
			if x.Type() != lua.LTString {
				err = merry.Appendf(err, "works[%d] имеет тип %s %+v, ожидалась строка\n", n, x, x.Type())
				return
			}
			works = append(works, scriptWork{s: lua.LVAsString(x)})
			return
		}

		if x.Type() != lua.LTFunction {
			err = merry.Appendf(err, "works[%d] имеет тип %s %+v, ожидалась функция Lua\n", x, x.Type())
			return
		}
		works[len(works)-1].w = func() error {
			return L.CallByParam(lua.P{
				Fn:      x,
				NRet:    0,
				Protect: true,
			})
		}
	})
	return works, nil
}
