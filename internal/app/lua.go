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
	"github.com/fpawel/atool/internal/pkg/numeth"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/comm/modbus"
	"github.com/lxn/win"
	"github.com/powerman/structlog"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"sort"
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

func (x *luaImport) InterpolationCoefficients(a *lua.LTable) lua.LValue {
	var dt []numeth.Coordinate
	a.ForEach(func(_ lua.LValue, a lua.LValue) {
		par, f := a.(*lua.LTable)
		if !f || par.Len() != 2 {
			x.l.RaiseError("type error: %+v: table with two elements expected", a)
		}
		vx, xOk := par.RawGetInt(1).(lua.LNumber)
		vy, yOk := par.RawGetInt(2).(lua.LNumber)
		if xOk && yOk {
			dt = append(dt, numeth.Coordinate{
				X: float64(vx),
				Y: float64(vy),
			})
		}
	})
	sort.Slice(dt, func(i, j int) bool {
		return dt[i].X < dt[i].Y
	})
	r, ok := numeth.InterpolationCoefficients(dt)

	if !ok {
		journal.Err(x.log(), fmt.Errorf("интерполяция: %+v: расчёт не может быть выполнен", dt))
		return lua.LNil
	}
	journal.Info(x.log(), fmt.Sprintf("интерполяция: %+v: %+v", dt, r))
	a = x.l.NewTable()
	for i, v := range r {
		a.RawSetInt(i+1, lua.LNumber(v))
	}
	return a
}

func (x *luaImport) Temperature(destinationTemperature float64) {
	what := fmt.Sprintf("перевод термокамеры на %v⁰C", destinationTemperature)
	luaWithGuiWarn(x.l, what, setupTemperature(log, x.l.Context(), destinationTemperature))
	x.pause(config.Get().Temperature.HoldDuration, fmt.Sprintf("выдержка на температуре %v⁰C", destinationTemperature))
}

func (x *luaImport) SwitchGas(gas byte) {
	what := fmt.Sprintf("подать газ %d", gas)
	if gas == 0 {
		what = "отключить газ"
	}
	luaWithGuiWarn(x.l, what, switchGas(x.l.Context(), gas))
	if gas != 0 {
		x.pause(config.Get().Gas.BlowGasDuration, what)
	}
}

func (x *luaImport) BlowGas(gas byte) {
	what := fmt.Sprintf("продуть газ %d", gas)
	luaWithGuiWarn(x.l, what, switchGas(x.l.Context(), gas))
	x.pause(config.Get().Gas.BlowGasDuration, what)
}

func (x *luaImport) ReadSave(reg modbus.Var, format modbus.FloatBitsFormat, dbKey string) {
	if err := format.Validate(); err != nil {
		x.l.ArgError(2, err.Error())
	}
	luaCheck(x.l, guiwork.PerformNewNamedWork(x.log(), x.l.Context(),
		fmt.Sprintf("считать из СОМ и сохранить: рег.%d,%s", reg, dbKey),
		func(log *structlog.Logger, ctx context.Context) error {
			return processEachActiveProduct(nil, func(product data.Product, device config.Device) error {
				return readAndSaveProductValue(x.log(), x.l.Context(),
					product, device, reg, format, dbKey)
			})
		}))
}

func (x *luaImport) Write32(cmd modbus.DevCmd, format modbus.FloatBitsFormat, value float64) {
	if err := format.Validate(); err != nil {
		x.l.ArgError(2, err.Error())
	}
	luaCheck(x.l, guiwork.PerformNewNamedWork(x.log(), x.l.Context(),
		fmt.Sprintf("команда %d(%v)", cmd, value),
		func(log *structlog.Logger, ctx context.Context) error {
			return processEachActiveProduct(nil, func(product data.Product, device config.Device) error {
				_ = write32Product(log, x.l.Context(), product, device, cmd, format, value)
				return nil
			})
		}))
}

func (x *luaImport) WriteKef(kef int, format modbus.FloatBitsFormat, value float64) {
	if err := format.Validate(); err != nil {
		x.l.ArgError(2, err.Error())
	}
	luaCheck(x.l, guiwork.PerformNewNamedWork(x.log(), x.l.Context(),
		fmt.Sprintf("запись K%d=%v", kef, value),
		func(log *structlog.Logger, ctx context.Context) error {
			return processEachActiveProduct(nil, func(product data.Product, device config.Device) error {
				_ = writeKefProduct(log, x.l.Context(), product, device, kef, format, value)
				return nil
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
	sort.Slice(luaParamValues, func(i, j int) bool {
		return luaParamValues[i].Name < luaParamValues[j].Name
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
