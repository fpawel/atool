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
	luaState *lua.LState
}

var (
	luaParamValues       []*apitypes.ConfigParamValue
	luaSelectedWorksChan = make(chan []bool)
	luaIgnoreError       = func() {}

	luaNaN = lua.LNumber(math.NaN())
)

func (x *luaImport) GetConfig() *lua.LTable {
	Config := x.luaState.NewTable()
	xs, err := getConfigParamsValues()
	x.luaCheck(err)
	for _, p := range xs {
		switch p.Type {
		case "int", "float":
			v, err := strconv.ParseFloat(p.Value, 64)
			if err != nil {
				x.luaState.RaiseError("%q=%v: %v", p.Name, p.Value, err)
			}
			Config.RawSetString(p.Key, luar.New(x.luaState, v))
		default:
			Config.RawSetString(p.Key, luar.New(x.luaState, p.Value))
		}
	}
	return Config
}

func (x *luaImport) GetProducts() *lua.LTable {
	Products := x.luaState.NewTable()

	party, err := data.GetCurrentParty()
	x.luaCheck(err)

	device, err := config.Get().Hardware.GetDevice(party.DeviceType)
	x.luaCheck(err)

	for _, p := range party.Products {
		p := p
		if !p.Active {
			continue
		}
		impP := newLuaProduct(p, device, x.luaState)
		Products.Append(luar.New(x.luaState, impP))
	}
	x.luaState.SetGlobal("Products", Products)

	return Products
}

func (x *luaImport) InterpolationCoefficients(a *lua.LTable) lua.LValue {
	var dt []numeth.Coordinate
	a.ForEach(func(_ lua.LValue, a lua.LValue) {
		par, f := a.(*lua.LTable)
		if !f || par.Len() != 2 {
			x.luaState.RaiseError("type error: %+v: table with two elements expected", a)
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
		r = make([]float64, len(dt))
		for i := range r {
			r[i] = math.NaN()
		}
		guiwork.JournalErr(log, fmt.Errorf("расёт не выполнен: %+v", dt))
	}
	a = x.luaState.NewTable()
	for i, v := range r {
		a.RawSetInt(i+1, lua.LNumber(v))
	}
	return a
}

func (x *luaImport) Temperature(destinationTemperature float64) {
	luaWithGuiWarn(x.luaState, setupTemperature(log, x.luaState.Context(), destinationTemperature))
	luaDelay(x.luaState, config.Get().Temperature.HoldDuration,
		fmt.Sprintf("выдержка на температуре %v⁰C", destinationTemperature))
}

func (x *luaImport) TemperatureStart() {
	tempDevice, err := getTemperatureDevice()
	x.luaCheck(err)
	luaWithGuiWarn(x.luaState, tempDevice.Start(log, x.luaState.Context()))
}

func (x *luaImport) TemperatureStop() {
	tempDevice, err := getTemperatureDevice()
	x.luaCheck(err)
	luaWithGuiWarn(x.luaState, tempDevice.Stop(log, x.luaState.Context()))
}

func (x *luaImport) TemperatureSetup(temperature float64) {
	x.newNestedWork1(fmt.Sprintf("перевод термокамеры на %v⁰C", temperature),
		func() error {
			tempDevice, err := getTemperatureDevice()
			x.luaCheck(err)
			return tempDevice.Setup(log, x.luaState.Context(), temperature)
		})
}

func (x *luaImport) SwitchGas(gas byte, warn bool) {

	err := switchGas(x.luaState.Context(), gas)
	if err == nil {
		return
	}
	what := fmt.Sprintf("подать газ %d", gas)
	if gas == 0 {
		what = "отключить газ"
	}
	err = merry.New(what).WithCause(err)
	if warn {
		luaWithGuiWarn(x.luaState, err)
		return
	}
	x.luaCheck(err)
}

func (x *luaImport) BlowGas(gas byte) {
	x.newNestedWork1(fmt.Sprintf("продуть газ %d", gas),
		func() error {
			x.SwitchGas(gas, true)
			luaDelay(x.luaState, config.Get().Gas.BlowDuration, fmt.Sprintf("продуть газ %d", gas))
			return nil
		})
}

func (x *luaImport) ReadSave(reg modbus.Var, format modbus.FloatBitsFormat, dbKey string) {
	if err := format.Validate(); err != nil {
		x.luaState.ArgError(2, err.Error())
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
		x.luaState.ArgError(2, err.Error())
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
		x.luaState.ArgError(2, err.Error())
	}
	x.luaCheck(guiwork.PerformNewNamedWork(log, x.luaState.Context(),
		fmt.Sprintf("запись K%d=%v", kef, value),
		func(log *structlog.Logger, ctx context.Context) error {
			return processEachActiveProduct(nil, func(product data.Product, device config.Device) error {
				_ = writeKefProduct(log, ctx, product, device, kef, format, value)
				return nil
			})
		}))
}

func (x *luaImport) Pause(strDuration string, what string) {
	duration, err := time.ParseDuration(strDuration)
	x.luaCheck(err)
	x.luaCheck(guiwork.Delay(log, x.luaState.Context(), duration, what, nil))
}

func (x *luaImport) Delay(strDuration string, what string) {
	duration, err := time.ParseDuration(strDuration)
	x.luaCheck(err)
	luaDelay(x.luaState, duration, what)
}

func (x *luaImport) ParamsDialog(arg *lua.LTable) *lua.LTable {

	luaParamValues = nil

	arg.ForEach(func(kx lua.LValue, vx lua.LValue) {
		var c apitypes.ConfigParamValue
		if err := setConfigParamFromLuaValue(kx, vx, &c); err != nil {
			x.luaState.RaiseError("%v:%v: %s", kx, vx, err)
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
			x.luaState.RaiseError("%+v: %s", a, err)
		}
		arg.RawSet(lua.LString(a.Key), value)
	}

	return arg
}

func (x *luaImport) Info(s string) {
	guiwork.JournalInfo(log, s)
}

func (x *luaImport) Err(s string) {
	guiwork.JournalErr(log, errors.New(s))
}

func (x *luaImport) SelectWorksDialog(arg *lua.LTable) {

	var (
		functions []func() error
		names     []string
	)
	arg.ForEach(func(n lua.LValue, arg lua.LValue) {
		if _, ok := n.(lua.LNumber); !ok {
			x.luaState.RaiseError("type error: %v: %v: key must be an integer index", n, arg)
		}
		w, ok := arg.(*lua.LTable)
		if !ok {
			x.luaState.RaiseError("type error: %v: %v: value must be a tuple of string and function", n, arg)
		}
		if w.Len() != 2 {
			x.luaState.RaiseError("type error: %v: %v: w.Len() != 2", n, arg)
		}
		what, ok := w.RawGetInt(1).(lua.LString)
		if !ok {
			x.luaState.RaiseError("type error: %v: %v: w.RawGetInt(1).(lua.LString)", n, arg)
		}
		Func, ok := w.RawGetInt(2).(*lua.LFunction)
		if !ok {
			x.luaState.RaiseError("type error: %v: %v: w.RawGetInt(2).(lua.LFunction)", n, arg)
		}
		names = append(names, string(what))
		functions = append(functions, func() error {
			return x.luaState.CallByParam(lua.P{
				Fn:      Func,
				Protect: true,
			})
		})
	})

	go gui.NotifyLuaSelectWorks(names)

	select {
	case <-x.luaState.Context().Done():
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

func (x *luaImport) luaCheck(err error) {
	luaCheck(x.luaState, err)
}

func (x *luaImport) newWork(name string, Func guiwork.WorkFunc) {
	x.luaCheck(guiwork.PerformNewNamedWork(log, x.luaState.Context(), name, Func))
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

func luaCheckNumberOrNil(luaState *lua.LState, n int) {
	switch luaState.Get(n).(type) {
	case *lua.LNilType:
		return
	case lua.LNumber:
		return
	default:
		luaState.TypeError(n, lua.LTNumber)
	}
}

func luaCheck(luaState *lua.LState, err error) {
	if merry.Is(err, context.Canceled) {
		return
	}
	if err != nil {
		luaState.RaiseError("%s", err)
	}
}

func luaDelay(luaState *lua.LState, dur time.Duration, what string) {
	luaCheck(luaState, delay(log, luaState.Context(), dur, what))
}

func luaWithGuiWarn(luaState *lua.LState, err error) {
	if err == nil {
		return
	}

	var ctxIgnoreError context.Context
	ctxIgnoreError, luaIgnoreError = context.WithCancel(luaState.Context())
	guiwork.NotifyLuaSuspended(log, err)
	<-ctxIgnoreError.Done()
	luaIgnoreError()
	if luaState.Context().Err() == nil {
		guiwork.JournalErr(log, merry.New("ошибка проигнорирована: выполнение продолжено").WithCause(err))
	}
}
