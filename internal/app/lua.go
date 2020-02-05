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
	"github.com/fpawel/atool/internal/pkg/numeth"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/comm/modbus"
	"github.com/powerman/structlog"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"math"
	"sort"
	"strconv"
	"time"
)

type luaImport struct {
	Config   *lua.LTable
	Products *lua.LTable
}

var (
	luaParamValues       []*apitypes.ConfigParamValue
	luaSelectedWorksChan = make(chan []bool)
	luaIgnoreError       = func() {}
	luaState             *lua.LState
	luaNaN               = lua.LNumber(math.NaN())
)

func (x *luaImport) init() error {
	x.Products = luaState.NewTable()
	x.Config = luaState.NewTable()
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
			x.Config.RawSetString(p.Key, luar.New(luaState, v))
		} else {
			x.Config.RawSetString(p.Key, luar.New(luaState, p.Value))
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
		if err := impP.init(p); err != nil {
			return err
		}

		x.Products.Append(luar.New(luaState, impP))
	}
	return nil
}

func (x *luaImport) InterpolationCoefficients(a *lua.LTable) lua.LValue {
	var dt []numeth.Coordinate
	a.ForEach(func(_ lua.LValue, a lua.LValue) {
		par, f := a.(*lua.LTable)
		if !f || par.Len() != 2 {
			luaState.RaiseError("type error: %+v: table with two elements expected", a)
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
		return lua.LNil
	}
	a = luaState.NewTable()
	for i, v := range r {
		a.RawSetInt(i+1, lua.LNumber(v))
	}
	return a
}

func (x *luaImport) Temperature(destinationTemperature float64) {
	luaWithGuiWarn(setupTemperature(luaLog(), luaState.Context(), destinationTemperature))
	luaDelay(config.Get().Temperature.HoldDuration, fmt.Sprintf("выдержка на температуре %v⁰C", destinationTemperature))
}

func (x *luaImport) TemperatureStart() {
	tempDevice, err := getTemperatureDevice()
	luaCheck(err)
	luaCheck(tempDevice.Start(luaLog(), luaState.Context()))
}

func (x *luaImport) TemperatureStop() {
	tempDevice, err := getTemperatureDevice()
	luaCheck(err)
	luaCheck(tempDevice.Stop(luaLog(), luaState.Context()))
}

func (x *luaImport) TemperatureSetup(temperature float64) {
	x.newNestedWork1(fmt.Sprintf("перевод термокамеры на %v⁰C", temperature),
		func() error {
			tempDevice, err := getTemperatureDevice()
			luaCheck(err)
			return tempDevice.Setup(luaLog(), luaState.Context(), temperature)
		})
}

func (x *luaImport) SwitchGas(gas byte, warn bool) {

	err := switchGas(luaState.Context(), gas)
	if err == nil {
		return
	}
	what := fmt.Sprintf("подать газ %d", gas)
	if gas == 0 {
		what = "отключить газ"
	}
	err = merry.New(what).WithCause(err)
	if warn {
		luaWithGuiWarn(err)
		return
	}
	luaCheck(err)
}

func (x *luaImport) BlowGas(gas byte) {
	x.newNestedWork1(fmt.Sprintf("продуть газ %d", gas),
		func() error {
			x.SwitchGas(gas, true)
			luaDelay(config.Get().Gas.BlowDuration, fmt.Sprintf("продуть газ %d", gas))
			return nil
		})
}

func (x *luaImport) ReadSave(reg modbus.Var, format modbus.FloatBitsFormat, dbKey string) {
	if err := format.Validate(); err != nil {
		luaState.ArgError(2, err.Error())
	}
	x.newWork(fmt.Sprintf("считать из СОМ и сохранить: рег.%d,%s", reg, dbKey),
		func(s *structlog.Logger, ctx context.Context) error {
			return processEachActiveProduct(nil, func(product data.Product, device config.Device) error {
				return readAndSaveProductValue(log, ctx,
					product, device, reg, format, dbKey)
			})
		})
}

func (x *luaImport) Write32(cmd modbus.DevCmd, format modbus.FloatBitsFormat, value float64) {
	if err := format.Validate(); err != nil {
		luaState.ArgError(2, err.Error())
	}
	x.newWork(fmt.Sprintf("команда %d(%v)", cmd, value), func(s *structlog.Logger, ctx context.Context) error {
		return processEachActiveProduct(nil, func(product data.Product, device config.Device) error {
			_ = write32Product(log, ctx, product, device, cmd, format, value)
			return nil
		})
	})
}

func (x *luaImport) WriteKef(kef int, format modbus.FloatBitsFormat, value float64) {
	if err := format.Validate(); err != nil {
		luaState.ArgError(2, err.Error())
	}
	luaCheck(guiwork.PerformNewNamedWork(luaLog(), luaState.Context(),
		fmt.Sprintf("запись K%d=%v", kef, value),
		func(log *structlog.Logger, ctx context.Context) error {
			return processEachActiveProduct(nil, func(product data.Product, device config.Device) error {
				_ = writeKefProduct(log, ctx, product, device, kef, format, value)
				return nil
			})
		}))
}

func (x *luaImport) Delay(strDuration string, what string) {
	duration, err := time.ParseDuration(strDuration)
	luaCheck(err)
	luaDelay(duration, what)
}

func (x *luaImport) ParamsDialog(arg *lua.LTable) *lua.LTable {

	luaParamValues = nil

	arg.ForEach(func(kx lua.LValue, vx lua.LValue) {
		var c apitypes.ConfigParamValue
		if err := setConfigParamFromLuaValue(kx, vx, &c); err != nil {
			luaState.RaiseError("%v:%v: %s", kx, vx, err)
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
			luaState.RaiseError("%+v: %s", a, err)
		}
		arg.RawSet(lua.LString(a.Key), value)
	}

	return arg
}

func (x *luaImport) Info(s string) {
	guiwork.JournalInfo(luaLog(), s)
}

func (x *luaImport) Err(s string) {
	guiwork.JournalErr(luaLog(), errors.New(s))
}

func (x *luaImport) SelectWorksDialog(arg *lua.LTable) {

	var (
		functions []func() error
		names     []string
	)
	arg.ForEach(func(n lua.LValue, arg lua.LValue) {
		if _, ok := n.(lua.LNumber); !ok {
			luaState.RaiseError("type error: %v: %v: key must be an integer index", n, arg)
		}
		w, ok := arg.(*lua.LTable)
		if !ok {
			luaState.RaiseError("type error: %v: %v: value must be a tuple of string and function", n, arg)
		}
		if w.Len() != 2 {
			luaState.RaiseError("type error: %v: %v: w.Len() != 2", n, arg)
		}
		what, ok := w.RawGetInt(1).(lua.LString)
		if !ok {
			luaState.RaiseError("type error: %v: %v: w.RawGetInt(1).(lua.LString)", n, arg)
		}
		Func, ok := w.RawGetInt(2).(*lua.LFunction)
		if !ok {
			luaState.RaiseError("type error: %v: %v: w.RawGetInt(2).(lua.LFunction)", n, arg)
		}
		names = append(names, string(what))
		functions = append(functions, func() error {
			return luaState.CallByParam(lua.P{
				Fn:      Func,
				Protect: true,
			})
		})
	})

	go gui.NotifyLuaSelectWorks(names)

	select {
	case <-luaState.Context().Done():
		return
	case luaSelectedWorks := <-luaSelectedWorksChan:
		for i, f := range luaSelectedWorks {
			if !f {
				continue
			}
			x.newNestedWork1(names[i], functions[i])
		}
	}
}

func (x *luaImport) NewWork(name string, Func func()) {
	x.newWork(name, func(logger, context.Context) error {
		Func()
		return nil
	})
}

func (x *luaImport) newWork(name string, Func guiwork.WorkFunc) {
	luaCheck(guiwork.PerformNewNamedWork(luaLog(), luaState.Context(), name, Func))
}

func (x *luaImport) newNestedWork1(name string, Func func() error) {
	x.newWork(name, func(logger, context.Context) error {
		return Func()
	})
}

func (x *luaImport) journalResult(s string, err error) {
	if err != nil {
		x.Err(fmt.Sprintf("%s: %s", s, err))
		return
	}
	x.Info(fmt.Sprintf("%s: успешно", s))
}

func luaLog() logger {
	return luaState.Context().Value("log").(logger)
}

func luaCheckNumberOrNil(n int) {
	switch luaState.Get(n).(type) {
	case *lua.LNilType:
		return
	case lua.LNumber:
		return
	default:
		luaState.TypeError(n, lua.LTNumber)
	}
}

func luaCheck(err error) {
	if merry.Is(err, context.Canceled) {
		return
	}
	if err != nil {
		luaState.RaiseError("%s", err)
	}
}

func luaDelay(dur time.Duration, what string) {
	luaCheck(delay(luaLog(), luaState.Context(), dur, what))
}

func luaWithGuiWarn(err error) {
	if err == nil {
		return
	}

	var ctxIgnoreError context.Context
	ctxIgnoreError, luaIgnoreError = context.WithCancel(luaState.Context())
	guiwork.NotifyLuaSuspended(luaLog(), err)
	<-ctxIgnoreError.Done()
	luaIgnoreError()
	if luaState.Context().Err() == nil {
		guiwork.JournalErr(log, merry.New("ошибка проигнорирована: выполнение продолжено").WithCause(err))
	}
}
