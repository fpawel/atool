package worklua

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/pkg/numeth"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/atool/internal/workparty"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	lua "github.com/yuin/gopher-lua"
	"math"
	"sort"
	"strings"
)

type luaProduct struct {
	Serial int
	ID     int64
	Addr   modbus.Addr
	String string

	p   workparty.Product
	l   *lua.LState
	log comm.Logger
}

func newLuaProduct(p workparty.Product, i *Import) *luaProduct {
	return &luaProduct{
		Serial: p.Serial,
		ID:     p.ProductID,
		Addr:   p.Addr,
		String: p.String(),
		p:      p,
		l:      i.l,
		log:    i.log,
	}
}

func (x *luaProduct) Perform(name string, Func func()) {
	x.do(workgui.NewFunc(fmt.Sprintf("%s: %s", x.p, name), func(log comm.Logger, ctx context.Context) error {
		Func()
		return nil
	}))
}

func (x *luaProduct) ReadKef(k modbus.Coefficient, format modbus.FloatBitsFormat) lua.LNumber {
	if err := format.Validate(); err != nil {
		x.l.ArgError(2, err.Error())
	}
	v, err := x.p.ReadKef(x.log, x.l.Context(), k, format)
	if err != nil {
		x.Err(fmt.Sprintf("считывание K%d: %v", k, err))
		return luaNaN
	}
	x.info(fmt.Sprintf("считатно K%d=%v", k, v))
	return lua.LNumber(v)
}

func (x *luaProduct) SetKef(k modbus.Coefficient, LValue lua.LNumber) {
	x.SetValue(data.KeyCoefficient(k), LValue)
}

func (x *luaProduct) WriteCoefficients(ks map[modbus.Coefficient]float64, format modbus.FloatBitsFormat) {
	for k, value := range ks {
		_ = x.p.WriteKef(k, format, value)(x.log, x.l.Context())
	}
}

func (x *luaProduct) SetValue(key string, LValue lua.LNumber) {
	value := float64(LValue)
	if math.IsNaN(value) {
		x.Err(fmt.Sprintf("%q: нет значения", key))
		_, err := data.DB.Exec(`DELETE FROM product_value WHERE product_id=? AND key=?`, x.p.ProductID, key)
		x.check(err)
		return
	}
	x.setValue(key, value)
	x.info(fmt.Sprintf("💾 %s = %v", key, value))
}

func (x *luaProduct) Kef(k modbus.Coefficient) lua.LNumber {
	return x.Value(data.KeyCoefficient(k))
}

func (x *luaProduct) Value(key string) lua.LNumber {
	var v float64
	err := data.DB.Get(&v,
		`SELECT value FROM product_value WHERE product_id = ? AND key = ?`,
		x.p.ProductID, key)
	if err == sql.ErrNoRows {
		x.Err(fmt.Sprintf("%q: нет значения", key))
		return luaNaN
	}
	if err != nil {
		x.log.Panic(err)
	}
	return lua.LNumber(v)
}

func (x *luaProduct) Info(args ...lua.LValue) {
	xs := make([]string, len(args))
	for i := range args {
		xs[i] = stringify(args[i])
	}
	x.info(strings.Join(xs, " "))
}

func (x *luaProduct) Err(err string) {
	workgui.NotifyErr(x.log, merry.Errorf("%s: %s", x.p, err))
}

func (x *luaProduct) Interpolation(name string, xy [][2]float64, k0, kCount int, format modbus.FloatBitsFormat) {

	what := fmt.Sprintf("📈 расчёт %s K%d...K%d 📝 %v", name, k0, k0+kCount, xy)

	var dt []numeth.Coordinate
	for _, pt := range xy {
		dt = append(dt, numeth.Coordinate{
			pt[0],
			pt[1],
		})
	}
	sort.Slice(dt, func(i, j int) bool {
		return dt[i][0] < dt[i][1]
	})

	r, ok := numeth.InterpolationCoefficients(dt)

	if !ok {
		r = make([]float64, len(dt))
		for i := range r {
			r[i] = math.NaN()
		}
		x.Err(fmt.Sprintf("%s - расчёт не выполнен", what))
		return
	}

	for len(r) < kCount {
		r = append(r, 0)
	}
	for i, value := range r {
		_ = x.p.WriteKef(modbus.Coefficient(k0+i), format, value)(x.log, x.l.Context())
	}
}

func (x *luaProduct) info(s string) {
	workgui.NotifyInfo(x.log, fmt.Sprintf("%s: %s", x.p, s))
}

func (x *luaProduct) check(err error) {
	check(x.l, err)
}

func (x *luaProduct) do(Func workgui.WorkFunc) {
	x.check(Func(x.log, x.l.Context()))
}

func (x *luaProduct) journalResult(s string, err error) {
	if err != nil {
		x.Err(fmt.Sprintf("%s: %s", s, err))
		return
	}
	x.info(fmt.Sprintf("%s: успешно", s))
}

func (x *luaProduct) setValue(key string, value float64) {
	x.check(data.SaveProductValue(x.p.ProductID, key, value))
}
