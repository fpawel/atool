package worklua

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
)

func stringify(v lua.LValue) string {
	return fmt.Sprintf("%+v", convertLuaValue(v, nil))
}

func convertLuaValue(value lua.LValue, visited map[*lua.LTable]bool) interface{} {

	if visited == nil {
		visited = make(map[*lua.LTable]bool)
	}
	if value == lua.LNil {
		return nil
	}

	switch converted := value.(type) {
	case lua.LBool:
		return bool(converted)
	case lua.LNumber:
		return float64(converted)
	case *lua.LNilType:
		return nil
	case lua.LString:
		return string(converted)
	case *lua.LTable:
		if visited[converted] {
			return "?nested"
		}
		visited[converted] = true

		isArray := true

		converted.ForEach(func(key, value lua.LValue) {
			if key.Type() != lua.LTNumber {
				isArray = false
			}
		})
		if isArray {
			ret := map[float64]interface{}{}
			converted.ForEach(func(i, value lua.LValue) {
				ret[float64(i.(lua.LNumber))] = convertLuaValue(value, visited)
			})
			return ret
		} else {
			ret := make(map[interface{}]interface{})
			converted.ForEach(func(key, value lua.LValue) {
				ret[convertLuaValue(key, visited)] = convertLuaValue(value, visited)
			})
			return ret
		}
	default:
		return value.Type().String()
	}
}
