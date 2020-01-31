package configlua

import (
	"github.com/fpawel/atool/internal/pkg/must"
	lua "github.com/yuin/gopher-lua"
	luajson "layeh.com/gopher-json"
	"os"
	"path/filepath"
)

type Section struct {
	Name   string
	Params []Param
}

type Param struct {
	Key  string
	Name string
}

func GetProductParamsSections() []Section {
	L := lua.NewState()
	luajson.Preload(L)
	filename := filepath.Join(filepath.Dir(os.Args[0]), "lua", "product_params.lua")
	must.PanicIf(L.DoFile(filename))

	getStrAt := func(luaValue lua.LValue, n int) string {
		return string(luaValue.(*lua.LTable).RawGetInt(n).(lua.LString))
	}
	getTblAt := func(luaValue lua.LValue, n int) *lua.LTable {
		return luaValue.(*lua.LTable).RawGetInt(n).(*lua.LTable)
	}

	var sections []Section

	L.GetGlobal("product_params").(*lua.LTable).ForEach(func(_ lua.LValue, aSect lua.LValue) {
		var sect Section

		sect.Name = getStrAt(aSect, 1)
		getTblAt(aSect, 2).ForEach(func(_ lua.LValue, aParam lua.LValue) {
			key := getStrAt(aParam, 1)
			name := getStrAt(aParam, 2)
			sect.Params = append(sect.Params, Param{
				Key:  key,
				Name: name,
			})
		})
		sections = append(sections, sect)
	})
	return sections
}
