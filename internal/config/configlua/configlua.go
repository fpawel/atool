package configlua

import (
	"fmt"
	"github.com/fpawel/atool/internal/pkg/must"
	lua "github.com/yuin/gopher-lua"
	"os"
	"path/filepath"
)

type ProductParamsSection struct {
	Name   string
	Params []ProductParam
}

type ProductParam struct {
	Key  string
	Name string
}

type ProductParamsSectionsList []ProductParamsSection

func (xs ProductParamsSectionsList) Keys() map[string]struct{} {
	r := map[string]struct{}{}
	for _, x := range xs {
		for _, p := range x.Params {
			r[p.Key] = struct{}{}
		}
	}
	return r
}

func (xs ProductParamsSectionsList) HasKey(key string) bool {
	for _, x := range xs {
		for _, p := range x.Params {
			if p.Key == key {
				return true
			}
		}
	}
	return false
}

func GetProductParamsSectionsList(filename string) (ProductParamsSectionsList, error) {
	L := lua.NewState()
	if err := L.DoFile(filename); err != nil {
		return nil, err
	}

	var err error

	getStrAt := func(luaValue lua.LValue, n int) string {
		vt, ok := luaValue.(*lua.LTable)
		if !ok {
			err = fmt.Errorf("type error: %+v: table excepted", luaValue)
			return ""
		}
		v, ok := vt.RawGetInt(n).(lua.LString)
		if !ok {
			err = fmt.Errorf("type error: %+v[%d]: string excepted", luaValue, n)
			return ""
		}
		return string(v)
	}
	getTblAt := func(luaValue lua.LValue, n int) *lua.LTable {
		return luaValue.(*lua.LTable).RawGetInt(n).(*lua.LTable)
	}

	retValue := L.Get(-1)
	retTab, ok := retValue.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("type error: %+v: table expected", retValue)
	}

	var sections ProductParamsSectionsList

	retTab.ForEach(func(_ lua.LValue, aSect lua.LValue) {
		var sect ProductParamsSection
		sect.Name = getStrAt(aSect, 1)
		getTblAt(aSect, 2).ForEach(func(_ lua.LValue, aParam lua.LValue) {
			key := getStrAt(aParam, 1)
			name := getStrAt(aParam, 2)
			sect.Params = append(sect.Params, ProductParam{
				Key:  key,
				Name: name,
			})
		})
		sections = append(sections, sect)
	})
	return sections, nil
}

var ProductParamsSections = func() ProductParamsSectionsList {
	xs, err := GetProductParamsSectionsList(
		filepath.Join(filepath.Dir(os.Args[0]), "lua", "product_params.lua"))
	must.PanicIf(err)
	return xs
}()
