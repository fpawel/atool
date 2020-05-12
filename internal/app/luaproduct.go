package app

import (
	"database/sql"
	"fmt"
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/guiwork"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	lua "github.com/yuin/gopher-lua"
	"math"
)

type luaProduct struct {
	p        data.Product
	Serial   int
	ID       int64
	Addr     modbus.Addr
	device   devicecfg.Device
	luaState *lua.LState
}

func newLuaProduct(p data.Product, device devicecfg.Device, luaState *lua.LState) *luaProduct {
	return &luaProduct{
		p:        p,
		Serial:   p.Serial,
		ID:       p.ProductID,
		Addr:     p.Addr,
		device:   device,
		luaState: luaState,
	}
}

func (x *luaProduct) WriteKef(k int, format modbus.FloatBitsFormat, LValue lua.LNumber) {
	if err := format.Validate(); err != nil {
		x.luaState.ArgError(2, err.Error())
	}
	luaCheckNumberOrNil(x.luaState, 4)
	value := float64(LValue)
	if math.IsNaN(value) {
		x.Err(fmt.Sprintf("запись К%d: нет значения", k))
		return
	}
	_ = writeKefProduct(log, x.luaState.Context(), x.p, x.device, k, format, value)
}

func (x *luaProduct) Write32(cmd modbus.DevCmd, format modbus.FloatBitsFormat, LValue lua.LNumber) {
	if err := format.Validate(); err != nil {
		x.luaState.ArgError(2, err.Error())
	}
	luaCheckNumberOrNil(x.luaState, 4)
	value := float64(LValue)
	if math.IsNaN(value) {
		x.Err(fmt.Sprintf("write32: cmd=%d: нет значения", cmd))
		return
	}
	_ = write32Product(log, x.luaState.Context(), x.p, x.device, cmd, format, value)
}

func (x *luaProduct) ReadReg(reg modbus.Var, format modbus.FloatBitsFormat) lua.LNumber {
	if err := format.Validate(); err != nil {
		x.luaState.ArgError(2, err.Error())
	}
	v, err := modbus.Read3Value(log, x.luaState.Context(), x.comm(), x.p.Addr, reg, format)
	if err != nil {
		x.Err(fmt.Sprintf("считывание рег%d: %v", reg, err))
		return luaNaN
	}
	x.Info(fmt.Sprintf("считатно рег%d=%v", reg, v))
	return lua.LNumber(v)
}

func (x *luaProduct) ReadKef(k modbus.Var, format modbus.FloatBitsFormat) lua.LNumber {
	if err := format.Validate(); err != nil {
		x.luaState.ArgError(2, err.Error())
	}
	v, err := modbus.Read3Value(log, x.luaState.Context(), x.comm(), x.p.Addr, 224+2*k, format)
	if err != nil {
		x.Err(fmt.Sprintf("считывание K%d: %v", k, err))
		return luaNaN
	}
	x.Info(fmt.Sprintf("считатно K%d=%v", k, v))
	return lua.LNumber(v)
}

func (x *luaProduct) DeleteKey(key string) {
	x.Info(fmt.Sprintf("удалить ключ %q", key))
	x.luaCheck(deleteProductKey(x.p.ProductID, key))
}

func (x *luaProduct) SetKef(k int, LValue lua.LNumber) {
	x.SetValue(dbKeyCoefficient(k), LValue)
}

func (x *luaProduct) SetValue(key string, LValue lua.LNumber) {
	value := float64(LValue)
	if math.IsNaN(value) {
		x.Err(fmt.Sprintf("%q: нет значения", key))
		_, err := data.DB.Exec(`DELETE FROM product_value WHERE product_id=? AND key=?`, x.p.ProductID, key)
		x.luaCheck(err)
		return
	}
	x.setValue(key, value)
}

func (x *luaProduct) Kef(k int) lua.LNumber {
	return x.Value(dbKeyCoefficient(k))
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
		log.Panic(err)
	}
	return lua.LNumber(v)
}

func (x *luaProduct) Info(s string) {
	guiwork.JournalInfo(log, fmt.Sprintf("№%d.id%d: %s", x.p.Serial, x.p.ProductID, s))
}

func (x *luaProduct) Err(s string) {
	guiwork.JournalErr(log, fmt.Errorf("№%d.id%d: %s", x.p.Serial, x.p.ProductID, s))
}

func (x *luaProduct) luaCheck(err error) {
	luaCheck(x.luaState, err)
}

func (x *luaProduct) journalResult(s string, err error) {
	if err != nil {
		x.Err(fmt.Sprintf("%s: %s", s, err))
		return
	}
	x.Info(fmt.Sprintf("%s: успешно", s))
}

func (x *luaProduct) comm() comm.T {
	return getCommProduct(x.p.Comport, x.device)
}

func (x *luaProduct) setValue(key string, value float64) {
	x.Info(fmt.Sprintf("сохранение %q=%v", key, value))
	x.luaCheck(saveProductValue(x.p.ProductID, key, value))
}
