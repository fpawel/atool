package app

import (
	"database/sql"
	"fmt"
	"github.com/fpawel/atool/internal/config"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/journal"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	lua "github.com/yuin/gopher-lua"
)

type luaProduct struct {
	p data.Product
	l *lua.LState

	Serial int
	Device config.Device
}

func (x *luaProduct) init(l *lua.LState, p data.Product) error {
	device, okDevice := config.Get().Hardware[p.Device]
	if !okDevice {
		return fmt.Errorf("%s: не заданы параметры устройства", p)
	}
	x.p = p
	x.l = l
	x.Serial = p.Serial
	x.Device = device
	return nil
}

func (x *luaProduct) WriteKef(k int, format modbus.FloatBitsFormat, pv lua.LValue) {
	if err := format.Validate(); err != nil {
		x.l.ArgError(2, err.Error())
	}
	luaCheckNumberOrNil(x.l, 4)
	if pv == lua.LNil {
		x.Err(fmt.Sprintf("запись К%d: нет значения", k))
		return
	}
	_ = writeKefProduct(log, x.l.Context(), x.p, x.Device, k, format, float64(pv.(lua.LNumber)))
}

func (x *luaProduct) Write32(cmd modbus.DevCmd, format modbus.FloatBitsFormat, pv lua.LValue) {
	if err := format.Validate(); err != nil {
		x.l.ArgError(2, err.Error())
	}
	luaCheckNumberOrNil(x.l, 4)
	if pv == lua.LNil {
		x.Err(fmt.Sprintf("write32: cmd=%d: нет значения", cmd))
		return
	}
	_ = write32Product(log, x.l.Context(), x.p, x.Device, cmd, format, float64(pv.(lua.LNumber)))
}

func (x *luaProduct) ReadReg(reg modbus.Var, format modbus.FloatBitsFormat) lua.LValue {
	if err := format.Validate(); err != nil {
		x.l.ArgError(2, err.Error())
	}
	v, err := modbus.Read3Value(log, x.l.Context(), x.comm(), x.p.Addr, reg, format)
	if err != nil {
		x.Err(fmt.Sprintf("считывание рег%d", reg))
		return lua.LNil
	}
	x.Info(fmt.Sprintf("считатно рег%d=%v", reg, v))
	return lua.LNumber(v)
}

func (x *luaProduct) ReadKef(k modbus.Var, format modbus.FloatBitsFormat) lua.LValue {
	if err := format.Validate(); err != nil {
		x.l.ArgError(2, err.Error())
	}
	v, err := modbus.Read3Value(log, x.l.Context(), x.comm(), x.p.Addr, 224+2*k, format)
	if err != nil {
		x.Err(fmt.Sprintf("считывание K%d", k))
		return lua.LNil
	}
	x.Info(fmt.Sprintf("считатно K%d=%v", k, v))
	return lua.LNumber(v)
}

func (x *luaProduct) DeleteKey(key string) {
	x.Info(fmt.Sprintf("удалить ключ %q", key))
	luaCheck(x.l, deleteProductKey(x.p.ProductID, key))
}

func (x *luaProduct) SetValue(key string, pv lua.LValue) {
	if !config.Get().IsProductParamKeyExists(key) {
		x.Err(fmt.Sprintf("%q: параметр приборов не задан в настройках", key))
	}
	if pv == lua.LNil {
		x.Err(fmt.Sprintf("%q: нет значения", key))
		return
	}
	v := float64(pv.(lua.LNumber))
	x.Info(fmt.Sprintf("сохранение %q=%v", key, v))
	luaCheck(x.l, saveProductValue(x.p.ProductID, key, v))
}

func (x *luaProduct) Value(key string) lua.LValue {
	var v float64
	err := db.Get(&v,
		`SELECT value FROM product_value WHERE product_id = ? AND key = ?`,
		x.p.ProductID, key)
	if err == sql.ErrNoRows {
		x.Err(fmt.Sprintf("%q: нет данных", key))
		return lua.LNil
	}
	if err != nil {
		log.Panic(err)
	}
	return lua.LNumber(v)
}

func (x *luaProduct) Info(s string) {
	journal.Info(log, fmt.Sprintf("прибор %d.%d: %s", x.p.Serial, x.p.ProductID, s))
}

func (x *luaProduct) Err(s string) {
	journal.Err(log, fmt.Errorf("прибор %d.%d: %s", x.p.Serial, x.p.ProductID, s))
}

func (x *luaProduct) journalResult(s string, err error) {
	if err != nil {
		x.Err(fmt.Sprintf("%s: %s", s, err))
		return
	}
	x.Info(fmt.Sprintf("%s: успешно", s))
}

func (x *luaProduct) comm() comm.T {
	return getCommProduct(x.p.Comport, x.Device)
}
