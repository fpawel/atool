package worklua

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config/appcfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/hardware"
	"github.com/fpawel/atool/internal/pkg/numeth"
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
	l   *lua.LState
	log comm.Logger
}

func NewImport(log comm.Logger, luaState *lua.LState) *Import {
	return &Import{
		l:   luaState,
		log: log,
	}
}

func (x *Import) Work(name string, Func func()) workgui.Work {
	return workgui.Work{
		Name: name,
		Func: func(*structlog.Logger, context.Context) error {
			Func()
			return nil
		},
	}
}

func (x *Import) WorkEachSelectedProduct(name string, Func func(*luaProduct)) workgui.Work {
	return workgui.Work{
		Name: name,
		Func: func(*structlog.Logger, context.Context) error {
			x.ForEachSelectedProduct(Func)
			return nil
		},
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
	x.check(hardware.WithWarn{}.HoldTemperature(destinationTemperature)(x.log, x.l.Context()))
}

func (x *Import) TemperatureStart() {
	x.do(hardware.TemperatureStart)
}

func (x *Import) TemperatureStop() {
	x.do(hardware.TemperatureStop)
}

func (x *Import) TemperatureSetup(temperature float64) {
	x.do(hardware.TemperatureSetup(temperature))
}

func (x *Import) SwitchGas(gas byte) {
	x.check(hardware.SwitchGas(gas).DoWarn(x.log, x.l.Context()))
}

func (x *Import) BlowGas(gas byte) {
	x.check(hardware.WithWarn{}.BlowGas(gas)(x.log, x.l.Context()))
}

func (x *Import) ReadAndSaveProductParam(reg modbus.Var, format modbus.FloatBitsFormat, dbKey string) {
	if err := format.Validate(); err != nil {
		x.l.ArgError(2, err.Error())
	}
	x.perform(fmt.Sprintf("📤 считать регистр %d 💾 сохранить %s %v", reg, dbKey, format),
		workparty.ReadAndSaveProductParam(reg, format, dbKey))
}

func (x *Import) Write32(cmd modbus.DevCmd, format modbus.FloatBitsFormat, value float64) {
	if err := format.Validate(); err != nil {
		x.l.ArgError(2, err.Error())
	}
	x.check(workparty.Write32(cmd, format, value)(x.log, x.l.Context()))
}

func (x *Import) Pause(strDuration string, what string) {
	duration, err := time.ParseDuration(strDuration)
	x.check(err)
	x.check(workgui.Delay(duration, what, nil)(x.log, x.l.Context()))
}

func (x *Import) Delay(strDuration string, what string) {
	duration, err := time.ParseDuration(strDuration)
	x.check(err)
	x.delay(duration, what)
}

func (x *Import) SelectWorksDialog(args workgui.Works) (selectedWorks workgui.Works) {
	return args.ExecuteSelectWorksDialog(x.l.Context().Done())
}

func (x *Import) ParamsDialog(arg *lua.LTable) *lua.LTable {

	workgui.ConfigParamValues = nil

	arg.ForEach(func(kx lua.LValue, vx lua.LValue) {
		c, err := newConfigParamValue(kx, vx)
		if err != nil {
			x.l.RaiseError("%v:%v: %s", kx, vx, err)
		}
		workgui.ConfigParamValues = append(workgui.ConfigParamValues, c)
	})
	sort.Slice(workgui.ConfigParamValues, func(i, j int) bool {
		return workgui.ConfigParamValues[i].Name < workgui.ConfigParamValues[j].Name
	})

	gui.RequestConfigParamValues()

	for _, a := range workgui.ConfigParamValues {
		value, err := configParamValue{a}.getLuaValue()
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

func (x *Import) PerformWorks(works []workgui.Work) {
	for _, work := range works {
		x.performContext(work.Name, func() error {
			return work.Func(x.log, x.l.Context())
		})
	}
}

func (x *Import) PerformEachSelectedProduct(name string, Func func(p *luaProduct)) {
	x.perform(name, func(comm.Logger, context.Context) error {
		x.ForEachSelectedProduct(func(product *luaProduct) {
			product.Perform(fmt.Sprintf("%s: %s", product.p, name), func() {
				Func(product)
			})
		})
		return nil
	})
}

func (x *Import) Perform(name string, Func func()) {
	x.perform(name, func(comm.Logger, context.Context) error {
		Func()
		return nil
	})
}

func (x *Import) WriteCoefficients(ks map[int]int, format modbus.FloatBitsFormat) {
	x.check(workparty.NewWorkWriteCfs(x.log, x.l.Context(), coefficientsList(ks), format))
}

func (x *Import) ReadCoefficients(ks map[int]int, format modbus.FloatBitsFormat) {
	x.check(workparty.ReadCfs(x.log, x.l.Context(), coefficientsList(ks), format))
}

func (x *Import) ReadAndSaveParam(param modbus.Var, format modbus.FloatBitsFormat, dbKey string) {
	x.check(workparty.ReadAndSaveProductParam(x.log, x.l.Context(), param, format, dbKey))
}

func (x *Import) perform(name string, Func workgui.WorkFunc) {
	x.check(workgui.Perform(x.log, x.l.Context(), name, Func))
}

func (x *Import) do(Func workgui.WorkFunc) {
	x.check(Func(x.log, x.l.Context()))
}

func (x *Import) performContext(name string, Func func() error) {
	x.perform(name, func(comm.Logger, context.Context) error {
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

//func (x *Import) withGuiWarn(err error) {
//	x.check( workgui.WithWarn(x.log, x.l.Context(), err) )
//}

func (x *Import) getProducts(selectedOnly bool) (Products []*luaProduct) {

	party, err := data.GetCurrentParty()
	x.check(err)

	device, err := appcfg.Cfg.Hardware.GetDevice(party.DeviceType)
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

func coefficientsList(xs map[int]int) (r workparty.CfsList) {
	for _, k := range xs {
		r = append(r, modbus.Var(k))
	}
	return
}

var (
	luaNaN = lua.LNumber(math.NaN())
)
