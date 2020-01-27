package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/guiwork"
	"github.com/fpawel/atool/internal/journal"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/comm/modbus"
	"github.com/lxn/win"
	"github.com/powerman/structlog"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"strconv"
	"time"
)

type luaImport struct {
	Config   *lua.LTable
	Products *lua.LTable
	l        *lua.LState
}

func (x *luaImport) init(L *lua.LState) error {
	x.l = L
	x.Products = L.NewTable()
	x.Config = L.NewTable()
	xs, err := getConfigParamsValues()
	if err != nil {
		return err
	}
	for _, p := range xs {
		if p.Type == "int" || p.Type == "float" {
			v, err := strconv.ParseFloat(p.Value, 64)
			if err != nil {
				return fmt.Errorf("%q=%v: %w", p.Name, p.Value, err)
			}
			x.Config.RawSetString(p.Key, luar.New(x.l, v))
		} else {
			x.Config.RawSetString(p.Key, luar.New(x.l, p.Value))
		}
	}

	party, err := data.GetCurrentParty(db)
	if err != nil {
		return err
	}

	for _, p := range party.Products {
		p := p
		if !p.Active {
			continue
		}
		impP := new(luaProduct)
		if err := impP.init(L, p); err != nil {
			return err
		}
		x.Products.Append(luar.New(L, impP))
	}
	return nil
}

func (x *luaImport) Temperature(destinationTemperature float64) {
	what := fmt.Sprintf("перевод термокамеры на %v⁰C", destinationTemperature)
	luaWithGuiWarn(x.l, what, setupTemperature(log, x.l.Context(), destinationTemperature))
	x.pause(config.Get().Temperature.HoldDuration, fmt.Sprintf("выдержка на температуре %v⁰C", destinationTemperature))
}

func (x *luaImport) Gas(gas byte) {
	what := fmt.Sprintf("продуть газ %d", gas)
	if gas == 0 {
		what = "отключить газ"
	}
	luaWithGuiWarn(x.l, what, switchGas(x.l.Context(), gas))
	if gas != 0 {
		x.pause(config.Get().Gas.BlowDuration, what)
	}
}

func (x *luaImport) ReadSave(reg modbus.Var, format modbus.FloatBitsFormat, dbKey string) {
	if err := format.Validate(); err != nil {
		x.l.ArgError(2, err.Error())
	}
	luaCheck(x.l, guiwork.PerformNewNamedWork(x.log(), x.l.Context(),
		fmt.Sprintf("считать из СОМ и сохранить: рег.%d,%s", reg, dbKey),
		func(log *structlog.Logger, ctx context.Context) error {
			return processEachActiveProductHardware(func(product data.Product, device config.Device) error {
				return readAndSaveProductValue(x.log(), x.l.Context(),
					product, device, reg, format, dbKey)
			})
		}))
}

func (x *luaImport) PauseSec(sec int64, what string) {
	dur := time.Second * time.Duration(sec)
	luaCheck(x.l, delay(x.log(), x.l.Context(), dur, what))
}

func (x *luaImport) ParamsDialog(arg *lua.LTable) *lua.LTable {

	luaParamValues = nil

	arg.ForEach(func(kx lua.LValue, vx lua.LValue) {
		var c apitypes.ConfigParamValue
		if err := setConfigParamFromLuaValue(kx, vx, &c); err != nil {
			x.l.RaiseError("%v:%v: %s", kx, vx, err)
		}
		luaParamValues = append(luaParamValues, &c)
	})

	gui.RequestLuaParams()

	for _, a := range luaParamValues {
		value, err := getLuaValueFromConfigParam(a)
		if err != nil {
			x.l.RaiseError("%+v: %s", a, err)
		}
		arg.RawSet(lua.LString(a.Key), value)
	}

	return arg
}

func (x *luaImport) Info(s string) {
	journal.Info(luaLog(x.l), s)
}

func (x *luaImport) Err(s string) {
	journal.Err(luaLog(x.l), errors.New(s))
}

func (x *luaImport) journalResult(s string, err error) {
	if err != nil {
		x.Err(fmt.Sprintf("%s: %s", s, err))
		return
	}
	x.Info(fmt.Sprintf("%s: успешно", s))
}

func (x *luaImport) log() logger {
	return luaLog(x.l)
}

func (x *luaImport) pause(dur time.Duration, what string) {
	luaCheck(x.l, delay(x.log(), x.l.Context(), dur, what))
}

func luaLog(L *lua.LState) logger {
	return L.Context().Value("log").(logger)
}

func luaCheckNumberOrNil(L *lua.LState, n int) {
	switch L.Get(n).(type) {
	case *lua.LNilType:
		return
	case lua.LNumber:
		return
	default:
		L.TypeError(n, lua.LTNumber)
	}
}

func luaCheck(L *lua.LState, err error) {
	if err != nil {
		L.RaiseError("%s", err)
	}
}

func luaWithGuiWarn(L *lua.LState, what string, err error) {
	if err == nil {
		return
	}
	if gui.MsgBox(what,
		formatError1(err)+"\n\nOK - продолжить выполнение\n\nОТМЕНА - прервать выполнение",
		win.MB_ICONWARNING|win.MB_OKCANCEL,
	) != win.IDOK {
		L.RaiseError("%s", err)
	}
	journal.Err(luaLog(L), merry.New("проигнорирована ошибка сценария").WithCause(err))
}
