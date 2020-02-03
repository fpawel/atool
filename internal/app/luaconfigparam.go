package app

import (
	"errors"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
	"strconv"
	"time"
)

type luaConfigParam struct {
	Name  string
	Type  string
	List  []string
	Value interface{}
}

func getLuaValueFromConfigParam(a *apitypes.ConfigParamValue) (lua.LValue, error) {
	switch a.Type {
	case "float":
		v, err := parseFloat(a.Value)
		if err != nil {
			return nil, err
		}
		return lua.LNumber(v), nil
	case "int":
		v, err := strconv.ParseInt(a.Value, 10, 64)
		if err != nil {
			return nil, err
		}
		return lua.LNumber(v), nil
	case "bool":
		v, err := strconv.ParseBool(a.Value)
		if err != nil {
			return nil, err
		}
		return lua.LBool(v), nil
	case "duration":
		v, err := time.ParseDuration(a.Value)
		if err != nil {
			return nil, err
		}
		return lua.LString(v), nil
	default:
		return lua.LString(a.Value), nil
	}
}

func setConfigParamFromLuaValue(kx, vx lua.LValue, a *apitypes.ConfigParamValue) error {

	v, ok := vx.(*lua.LTable)
	if !ok {
		return errors.New("type error: value must be table")
	}
	k, ok := kx.(lua.LString)
	if !ok {
		return errors.New("type error: key must be string")
	}

	var c luaConfigParam
	if err := gluamapper.Map(v, &c); err != nil {
		return err
	}
	*a = apitypes.ConfigParamValue{
		Name:       c.Name,
		Type:       c.Type,
		Key:        string(k),
		ValuesList: c.List,
	}
	if len(a.ValuesList) == 0 {
		a.ValuesList = []string{}
	}

	switch v := c.Value.(type) {
	case float64:
		a.Value = formatFloat(v)
		if len(a.Type) == 0 {
			a.Type = "float"
		}
	case bool:
		a.Type = "bool"
		a.Value = strconv.FormatBool(v)
	case string:
		a.Type = "string"
		a.Value = v
	default:
		return errors.New("type error: value")
	}
	return nil
}