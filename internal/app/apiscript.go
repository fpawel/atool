package app

import (
	"context"
	"github.com/fpawel/atool/internal/gui/guiwork"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/powerman/structlog"
	lua "github.com/yuin/gopher-lua"
	"path/filepath"
	"time"
)

type scriptSvc struct{}

var _ api.ScriptService = new(scriptSvc)

func (_ *scriptSvc) RunFile(_ context.Context, filename string) error {
	return guiwork.RunWork(log, appCtx, filepath.Base(filename), func(log *structlog.Logger, ctx context.Context) error {
		L := lua.NewState()
		defer L.Close()
		L.SetContext(ctx)
		L.SetGlobal("delay_sec", L.NewFunction(func(state *lua.LState) int {
			sec := L.ToInt64(1)
			what := L.ToString(2)

			ms := new(measurements)
			defer func() {
				saveMeasurements(ms.xs)
			}()
			_ = guiwork.Delay(log, ctx, time.Second*time.Duration(sec), what, func(log *structlog.Logger, ctx context.Context) error {
				return readProductsParams(ctx, ms)
			})
			return 0
		}))

		return L.DoFile(filename)
	})
}

//type scriptWork struct {
//	w func() error
//	s string
//}
//
//func getScriptWorks(L *lua.LState) ([]scriptWork, error) {
//	var (
//		works []scriptWork
//		err   error
//	)
//	L.GetGlobal("works").(*lua.LTable).ForEach(func(luaIndex lua.LValue, x lua.LValue) {
//		if luaIndex.Type() != lua.LTNumber {
//			err = merry.Appendf(err, "ключ элемента объекта 'works': %s %+v, ожидался номер\n", luaIndex.Type(), luaIndex)
//			return
//		}
//		n := int(lua.LVAsNumber(luaIndex))
//		if n%2 != 0 {
//			if x.Type() != lua.LTString {
//				err = merry.Appendf(err, "works[%d] имеет тип %s %+v, ожидалась строка\n", n, x, x.Type())
//				return
//			}
//			works = append(works, scriptWork{s: lua.LVAsString(x)})
//			return
//		}
//
//		if x.Type() != lua.LTFunction {
//			err = merry.Appendf(err, "works[%d] имеет тип %s %+v, ожидалась функция Lua\n", x, x.Type())
//			return
//		}
//		works[len(works)-1].w = func() error {
//			return L.CallByParam(lua.P{
//				Fn:      x,
//				NRet:    0,
//				Protect: true,
//			})
//		}
//	})
//	return works, nil
//}
