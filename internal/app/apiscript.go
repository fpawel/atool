package app

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/thriftgen/api"
	lua "github.com/yuin/gopher-lua"
)

type scriptSvc struct{}

var _ api.ScriptService = new(scriptSvc)

func (_ *scriptSvc) ListWorksNames(ctx context.Context) ([]string, error) {
	L := lua.NewState()
	defer L.Close()

	if err := L.DoFile("works.lua"); err != nil {
		return nil, err
	}

	var (
		works []string
		err   error
	)
	L.GetGlobal("works").(*lua.LTable).ForEach(func(index lua.LValue, x lua.LValue) {
		if index.Type() != lua.LTNumber {
			err = merry.Appendf(err, "ключ элемента объекта 'works' имеет тип %s, ожидался номер\n", index.Type())
			return
		}
		if x.Type() != lua.LTTable {
			err = merry.Appendf(err, "значение ключа объекта 'works' имеет тип %s, ожидалась таблица Lua\n", x.Type())
			return
		}

		funName := L.GetTable(x, lua.LNumber(1))
		if funName.Type() != lua.LTString {
			err = merry.Appendf(err, "ключ элемента [1] объекта 'works' имеет тип %s, ожидалась строка\n", funName.Type())
			return
		}

		funBody := L.GetTable(x, lua.LNumber(2))
		if funBody.Type() != lua.LTFunction {
			err = merry.Appendf(err, "ключ элемента [2] объекта 'works' имеет тип %s, ожидалась функция Lua\n", funBody.Type())
			return
		}

		if err == nil {
			works = append(works, lua.LVAsString(funName))
		}
	})
	return works, err
}
