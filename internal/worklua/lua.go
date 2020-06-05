package worklua

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/config/appcfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/hardware"
	"github.com/fpawel/atool/internal/pkg/intrng"
	"github.com/fpawel/atool/internal/pkg/numeth"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/atool/internal/workparty"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/powerman/structlog"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Import struct {
	ParamValues       []*apitypes.ConfigParamValue
	SelectedWorksChan chan []bool
	IgnoreError       func()
	l                 *lua.LState
	log               comm.Logger
}

func NewImport(log comm.Logger, luaState *lua.LState) *Import {
	return &Import{
		SelectedWorksChan: make(chan []bool),
		IgnoreError:       func() {},
		l:                 luaState,
		log:               log,
	}
}

func (x *Import) GetConfig() *lua.LTable {
	Config := x.l.NewTable()
	xs, err := appcfg.GetParamsValues()
	x.check(err)
	for _, p := range xs {
		switch p.Type {
		case "int", "float":
			v, err := strconv.ParseFloat(p.Value, 64)
			if err != nil {
				x.l.RaiseError("%q=%v: %v", p.Name, p.Value, err)
			}
			Config.RawSetString(p.Key, luar.New(x.l, v))
		default:
			Config.RawSetString(p.Key, luar.New(x.l, p.Value))
		}
	}
	return Config
}

func (x *Import) ForEachSelectedProduct(f func(*luaProduct)) {
	for _, p := range x.getProducts(true) {
		f(p)
	}
}

func (x *Import) ForEachProduct(f func(*luaProduct)) {
	for _, p := range x.getProducts(false) {
		f(p)
	}
}

func (x *Import) InterpolationCoefficients(a *lua.LTable) lua.LValue {
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
		r = make([]float64, len(dt))
		for i := range r {
			r[i] = math.NaN()
		}
		workgui.NotifyErr(x.log, merry.Errorf("расчёт не выполнен: %+v", dt))
	}
	a = x.l.NewTable()
	for i, v := range r {
		a.RawSetInt(i+1, lua.LNumber(v))
	}
	return a
}

func (x *Import) Temperature(destinationTemperature float64) {
	x.NewWork(fmt.Sprintf("перевод термокамеры на %v⁰C", destinationTemperature), func() {
		x.withGuiWarn(hardware.TemperatureSetup(x.log, x.l.Context(), destinationTemperature))
	})
	x.delay(config.Get().Temperature.HoldDuration,
		fmt.Sprintf("выдержка на температуре %v⁰C", destinationTemperature))
}

func (x *Import) TemperatureStart() {
	x.withGuiWarn(hardware.TemperatureStart(x.log, x.l.Context()))
}

func (x *Import) TemperatureStop() {
	x.withGuiWarn(hardware.TemperatureStop(x.log, x.l.Context()))
}

func (x *Import) TemperatureSetup(temperature float64) {
	x.newNestedWork1(fmt.Sprintf("перевод термокамеры на %v⁰C", temperature),
		func() error {
			return hardware.TemperatureSetup(x.log, x.l.Context(), temperature)
		})
}

func (x *Import) SwitchGas(gas byte, warn bool) {

	err := hardware.SwitchGas(x.log, x.l.Context(), gas)
	if err == nil {
		return
	}
	what := fmt.Sprintf("подать газ %d", gas)
	if gas == 0 {
		what = "отключить газ"
	}
	err = merry.New(what).WithCause(err)
	if warn {
		x.withGuiWarn(err)
		return
	}
	x.check(err)
}

func (x *Import) BlowGas(gas byte) {
	x.newNestedWork1(fmt.Sprintf("продуть газ %d", gas),
		func() error {
			x.SwitchGas(gas, true)
			x.delay(config.Get().Gas.BlowDuration, fmt.Sprintf("продуть газ %d", gas))
			return nil
		})
}

func (x *Import) ReadAndSaveProductParam(reg modbus.Var, format modbus.FloatBitsFormat, dbKey string) {
	if err := format.Validate(); err != nil {
		x.l.ArgError(2, err.Error())
	}
	x.newWork(fmt.Sprintf("📤 считать регистр %d 💾 сохранить %s %v", reg, dbKey, format),
		func(log *structlog.Logger, ctx context.Context) error {
			return workparty.ReadAndSaveProductParam(x.log, ctx, reg, format, dbKey)
		})
}

func (x *Import) Write32(cmd modbus.DevCmd, format modbus.FloatBitsFormat, value float64) {
	if err := format.Validate(); err != nil {
		x.l.ArgError(2, err.Error())
	}
	err := workparty.Write32(x.log, x.l.Context(), cmd, format, value)
	x.check(err)
}

func (x *Import) Pause(strDuration string, what string) {
	duration, err := time.ParseDuration(strDuration)
	x.check(err)
	x.check(workgui.Delay(x.log, x.l.Context(), duration, what, nil))
}

func (x *Import) Delay(strDuration string, what string) {
	duration, err := time.ParseDuration(strDuration)
	x.check(err)
	x.delay(duration, what)
}

func (x *Import) ParamsDialog(arg *lua.LTable) *lua.LTable {

	x.ParamValues = nil

	arg.ForEach(func(kx lua.LValue, vx lua.LValue) {
		var c apitypes.ConfigParamValue
		if err := setConfigParamFromLuaValue(kx, vx, &c); err != nil {
			x.l.RaiseError("%v:%v: %s", kx, vx, err)
		}
		x.ParamValues = append(x.ParamValues, &c)
	})
	sort.Slice(x.ParamValues, func(i, j int) bool {
		return x.ParamValues[i].Name < x.ParamValues[j].Name
	})

	gui.RequestLuaParams()

	for _, a := range x.ParamValues {
		value, err := getLuaValueFromConfigParam(a)
		if err != nil {
			x.l.RaiseError("%+v: %s", a, err)
		}
		arg.RawSet(lua.LString(a.Key), value)
	}

	return arg
}

func (x *Import) Stringify(v lua.LValue) string {
	return stringify(v)
}

func (x *Import) Info(args ...lua.LValue) {
	xs := make([]string, len(args))
	for i := range args {
		xs[i] = stringify(args[i])
	}
	workgui.NotifyInfo(x.log, strings.Join(xs, " "))
}

func (x *Import) Err(s lua.LValue) {
	workgui.NotifyErr(x.log, merry.New(stringify(s)))
}

func (x *Import) SelectWorksDialog(arg *lua.LTable) {

	var (
		functions []func() error
		names     []string
	)
	arg.ForEach(func(n lua.LValue, arg lua.LValue) {
		if _, ok := n.(lua.LNumber); !ok {
			x.l.RaiseError("type error: %v: %v: key must be an integer index", n, arg)
		}
		w, ok := arg.(*lua.LTable)
		if !ok {
			x.l.RaiseError("type error: %v: %v: value must be a tuple of string and function", n, arg)
		}
		if w.Len() != 2 {
			x.l.RaiseError("type error: %v: %v: w.Len() != 2", n, arg)
		}
		what, ok := w.RawGetInt(1).(lua.LString)
		if !ok {
			x.l.RaiseError("type error: %v: %v: w.RawGetInt(1).(lua.LString)", n, arg)
		}
		Func, ok := w.RawGetInt(2).(*lua.LFunction)
		if !ok {
			x.l.RaiseError("type error: %v: %v: w.RawGetInt(2).(lua.LFunction)", n, arg)
		}
		names = append(names, string(what))
		functions = append(functions, func() error {
			return x.l.CallByParam(lua.P{
				Fn:      Func,
				Protect: true,
			})
		})
	})

	go gui.NotifyLuaSelectWorks(names)

	select {
	case <-x.l.Context().Done():
		return
	case luaSelectedWorks := <-x.SelectedWorksChan:
		for i, f := range luaSelectedWorks {
			if !f {
				continue
			}
			x.newNestedWork1(names[i], functions[i])
		}
	}
}

func (x *Import) NewWorkEachProduct(name string, Func func(p *luaProduct)) {
	x.newWork(name, func(comm.Logger, context.Context) error {
		x.ForEachProduct(func(product *luaProduct) {
			product.NewWork(fmt.Sprintf("%s: %s", product.p, name), func() {
				Func(product)
			})
		})
		return nil
	})
}

func (x *Import) NewWork(name string, Func func()) {
	x.newWork(name, func(comm.Logger, context.Context) error {
		Func()
		return nil
	})
}

func formatCoefficients(ks map[int]int) string {
	var coefficients []int
	for _, k := range ks {
		coefficients = append(coefficients, k)
	}
	return fmt.Sprintf("запись коэффициентов %v", intrng.IntRanges(coefficients))
}

func (x *Import) WriteCoefficients(ks map[int]int, format modbus.FloatBitsFormat) {
	x.check(workparty.WriteCoefficients(x.log, x.l.Context(), coefficientsList(ks), format))
}

func (x *Import) ReadCoefficients(ks map[int]int, format modbus.FloatBitsFormat) {
	x.check(workparty.ReadCoefficients(x.log, x.l.Context(), coefficientsList(ks), format))
}

func (x *Import) ReadAndSaveParam(param modbus.Var, format modbus.FloatBitsFormat, dbKey string) {
	x.check(workparty.ReadAndSaveProductParam(x.log, x.l.Context(), param, format, dbKey))
}

func (x *Import) newWork(name string, Func workgui.WorkFunc) {
	x.check(workgui.PerformNewNamedWork(x.log, x.l.Context(), name, Func))
}

func (x *Import) newNestedWork1(name string, Func func() error) {
	x.newWork(name, func(comm.Logger, context.Context) error {
		return Func()
	})
}

func (x *Import) journalResult(s string, err error) {
	if err != nil {
		workgui.NotifyErr(x.log, fmt.Errorf("%s: %s", s, err))
		return
	}
	workgui.NotifyInfo(x.log, fmt.Sprintf("%s: успешно", s))
}

func (x *Import) delay(dur time.Duration, what string) {
	x.check(workparty.Delay(x.log, x.l.Context(), dur, what))
}

func (x *Import) check(err error) {
	check(x.l, err)
}

func (x *Import) withGuiWarn(err error) {
	if err == nil {
		return
	}
	var ctxIgnoreError context.Context
	ctxIgnoreError, x.IgnoreError = context.WithCancel(x.l.Context())
	workgui.NotifyWorkSuspended(err)
	<-ctxIgnoreError.Done()
	x.IgnoreError()
	if x.l.Context().Err() == nil {
		workgui.NotifyWarn(x.log, "ошибка проигнорирована")
	}
}

func (x *Import) getProducts(selectedOnly bool) (Products []*luaProduct) {

	party, err := data.GetCurrentParty()
	x.check(err)

	device, err := config.Get().Hardware.GetDevice(party.DeviceType)
	x.check(err)

	for _, p := range party.Products {
		p := p
		if selectedOnly && !p.Active {
			continue
		}
		Products = append(Products, newLuaProduct(workparty.Product{
			Product: p,
			Device:  device,
			Party:   party,
		}, x))
	}
	return
}

func check(l *lua.LState, err error) {
	if merry.Is(err, context.Canceled) {
		return
	}
	if err != nil {
		l.RaiseError("%s", err)
	}
}

func coefficientsList(xs map[int]int) (r workparty.CoefficientsList) {
	for _, k := range xs {
		r = append(r, modbus.Var(k))
	}
	return
}

var (
	luaNaN = lua.LNumber(math.NaN())
)
