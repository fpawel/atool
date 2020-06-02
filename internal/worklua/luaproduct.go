package worklua

import (
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/atool/internal/workparty"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	lua "github.com/yuin/gopher-lua"
	"math"
)

type luaProduct struct {
	p      workparty.Product
	Serial int
	ID     int64
	Addr   modbus.Addr
	l      *lua.LState
	log    comm.Logger
}

func newLuaProduct(p workparty.Product, i *Import) *luaProduct {
	return &luaProduct{
		p:      p,
		Serial: p.Serial,
		ID:     p.ProductID,
		Addr:   p.Addr,
		l:      i.l,
		log:    i.log,
	}
}

func (x *luaProduct) WriteKef(k int, format modbus.FloatBitsFormat, LValue lua.LNumber) {
	if err := format.Validate(); err != nil {
		x.l.ArgError(2, err.Error())
	}
	checkNumberOrNil(x.l, 4)
	value := float64(LValue)
	if math.IsNaN(value) {
		x.Err(fmt.Sprintf("запись К%d: нет значения", k))
		return
	}
	_ = x.p.WriteKef(x.log, x.l.Context(), k, format, value)
}

func (x *luaProduct) Write32(cmd modbus.DevCmd, format modbus.FloatBitsFormat, LValue lua.LNumber) {
	if err := format.Validate(); err != nil {
		x.l.ArgError(2, err.Error())
	}
	checkNumberOrNil(x.l, 4)
	value := float64(LValue)
	if math.IsNaN(value) {
		x.Err(fmt.Sprintf("write32: cmd=%d: нет значения", cmd))
		return
	}
	_ = x.p.Write32(x.log, x.l.Context(), cmd, format, value)
}

func (x *luaProduct) ReadReg(reg modbus.Var, format modbus.FloatBitsFormat) lua.LNumber {
	if err := format.Validate(); err != nil {
		x.l.ArgError(2, err.Error())
	}
	v, err := x.p.ReadParamValue(x.log, x.l.Context(), reg, format)
	if err != nil {
		x.Err(fmt.Sprintf("считывание рег%d: %v", reg, err))
		return luaNaN
	}
	x.Info(fmt.Sprintf("считатно рег%d=%v", reg, v))
	return lua.LNumber(v)
}

func (x *luaProduct) ReadKef(k modbus.Var, format modbus.FloatBitsFormat) lua.LNumber {
	if err := format.Validate(); err != nil {
		x.l.ArgError(2, err.Error())
	}
	v, err := x.p.ReadKef(x.log, x.l.Context(), k, format)
	if err != nil {
		x.Err(fmt.Sprintf("считывание K%d: %v", k, err))
		return luaNaN
	}
	x.Info(fmt.Sprintf("считатно K%d=%v", k, v))
	return lua.LNumber(v)
}

func (x *luaProduct) DeleteKey(key string) {
	x.Info(fmt.Sprintf("удалить ключ %q", key))
	x.check(data.DeleteProductKey(x.p.ProductID, key))
}

func (x *luaProduct) SetKef(k int, LValue lua.LNumber) {
	x.SetValue(data.KeyCoefficient(k), LValue)
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
}

func (x *luaProduct) Kef(k int) lua.LNumber {
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

func (x *luaProduct) Info(s string) {
	workgui.NotifyInfo(x.log, fmt.Sprintf("№%d.id%d: %s", x.p.Serial, x.p.ProductID, s))
}

func (x *luaProduct) Err(s string) {
	workgui.NotifyErr(x.log, merry.Errorf("№%d.id%d: %s", x.p.Serial, x.p.ProductID, s))
}

func (x *luaProduct) check(err error) {
	check(x.l, err)
}

func (x *luaProduct) journalResult(s string, err error) {
	if err != nil {
		x.Err(fmt.Sprintf("%s: %s", s, err))
		return
	}
	x.Info(fmt.Sprintf("%s: успешно", s))
}

func (x *luaProduct) setValue(key string, value float64) {
	x.Info(fmt.Sprintf("сохранение %q=%v", key, value))
	x.check(data.SaveProductValue(x.p.ProductID, key, value))
}

func checkNumberOrNil(l *lua.LState, n int) {
	switch l.Get(n).(type) {
	case *lua.LNilType:
		return
	case lua.LNumber:
		return
	default:
		l.TypeError(n, lua.LTNumber)
	}
}
