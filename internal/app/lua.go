package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/guiwork"
	"github.com/fpawel/atool/internal/journal"
	"github.com/fpawel/comm/modbus"
	"github.com/lxn/win"
	"github.com/powerman/structlog"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"time"
)

func luaImportParty(L *lua.LState) error {
	m, err := getCurrentPartyValues()
	if err != nil {
		return err
	}

	impParty := dynamic{}

	party, err := data.GetCurrentParty(db)
	if err != nil {
		return err
	}
	impParty["product_type"] = party.ProductType

	for k, name := range config.Get().PartyParams {
		if v, f := m[k]; !f {
			return fmt.Errorf("не задано значение параметра файла: %q", name)
		} else {
			impParty[k] = v
		}
	}

	impProducts := L.NewTable()

	for _, p := range party.Products {
		p := p
		if !p.Active {
			continue
		}
		impP, err := newLuaProduct(L, p)
		if err != nil {
			return err
		}
		impProducts.Append(luar.New(L, impP))
	}

	impParty["products"] = impProducts

	L.SetGlobal("party", luar.New(L, impParty))
	return nil
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

func luaWithGuiWarn(L *lua.LState, err error) {
	if err == nil {
		return
	}
	if gui.MsgBox("Ошибка сценария",
		formatError1(err)+"\n\nOK - продолжить выполнение\n\nОТМЕНА - прервать выполнение",
		win.MB_ICONWARNING|win.MB_OKCANCEL,
	) != win.IDOK {
		L.RaiseError("%s", err)
	}
	journal.Err(log, merry.New("проигнорирована ошибка сценария").WithCause(err))
}

func luaReadSave(log logger) lua.LGFunction {
	return func(L *lua.LState) int {
		param := L.CheckInt(1)
		format := modbus.FloatBitsFormat(L.CheckString(2))
		dbKey := L.CheckString(3)
		if err := format.Validate(); err != nil {
			L.ArgError(2, err.Error())
		}
		luaCheck(L, guiwork.PerformNewNamedWork(log, L.Context(),
			fmt.Sprintf("считать из СОМ и сохранить: рег.%d,%s", param, dbKey),
			func(log *structlog.Logger, ctx context.Context) error {
				return processEachActiveProduct(func(product data.Product, device config.Device) error {
					return readAndSaveProductValue(log, L.Context(),
						product, device, modbus.Var(param), format, dbKey)
				})
			}))
		return 0
	}
}

func luaPause(log logger) lua.LGFunction {
	return func(L *lua.LState) int {
		sec := L.ToInt64(1)
		what := L.ToString(2)
		dur := time.Second * time.Duration(sec)
		luaCheck(L, delay(log, L.Context(), dur, what))
		return 0
	}
}

func luaGas(log logger) lua.LGFunction {
	return func(L *lua.LState) int {
		gas := L.CheckInt(1)
		dur := luaGetDuration(L, 2, config.Get().BlowGas)
		luaWithGuiWarn(L, switchGas(L.Context(), byte(gas)))
		if gas != 0 {
			luaCheck(L, delay(log, L.Context(), dur, fmt.Sprintf("продувка газа %d", gas)))
		}
		return 0
	}
}

func luaTemperature(log logger) lua.LGFunction {
	return func(L *lua.LState) int {
		destinationTemperature := float64(L.CheckNumber(1))
		dur := luaGetDuration(L, 2, config.Get().HoldTemperature)
		luaWithGuiWarn(L, guiwork.PerformNewNamedWork(log, L.Context(),
			fmt.Sprintf("перевод термокамеры на %v⁰C", destinationTemperature),
			func(log logger, ctx context.Context) error {
				return setupTemperature(log, ctx, destinationTemperature)
			}))
		luaCheck(L, delay(log, L.Context(), dur,
			fmt.Sprintf("выдержка на температуре %v⁰C", destinationTemperature)))
		return 0
	}
}

func luaGetDuration(L *lua.LState, nArg int, def time.Duration) time.Duration {
	arg := L.Get(nArg)
	if arg == lua.LNil {
		return def
	}

	if minutes, ok := arg.(lua.LNumber); ok {
		return time.Minute * time.Duration(minutes)
	}
	strDur := L.CheckString(nArg)
	dur, err := time.ParseDuration(strDur)
	if err != nil {
		L.ArgError(nArg, err.Error())
	}
	return dur
}
