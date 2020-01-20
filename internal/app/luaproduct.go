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
	luar "layeh.com/gopher-luar"
)

type luaProduct struct {
	p data.Product
	l *lua.LState

	Serial int
	Device config.Device
}

func newLuaProduct(l *lua.LState, p data.Product) (lua.LValue, error) {
	device, okDevice := config.Get().Hardware[p.Device]
	if !okDevice {
		return nil, fmt.Errorf("%s: не заданы параметры устройства", p)
	}
	return luar.New(l, luaProduct{
		p:      p,
		l:      l,
		Serial: p.Serial,
		Device: device,
	}), nil
}

func (x luaProduct) WriteKef(k modbus.DevCmd, format modbus.FloatBitsFormat, pv lua.LValue) {
	if err := format.Validate(); err != nil {
		x.l.ArgError(2, err.Error())
	}
	luaCheckNumberOrNil(x.l, 4)
	if pv == lua.LNil {
		x.Err(fmt.Sprintf("запись К%d: нет значения", k))
		return
	}
	v := pv.(lua.LNumber)
	x.journalResult(
		fmt.Sprintf("запись К%d=%v format=%s", k, v, format),
		modbus.RequestWrite32{
			Addr:      x.p.Addr,
			ProtoCmd:  0x10,
			DeviceCmd: (0x80 << 8) + k,
			Format:    format,
			Value:     float64(v),
		}.GetResponse(log, x.l.Context(), x.comm()))
}

func (x luaProduct) Write32(cmd modbus.DevCmd, format modbus.FloatBitsFormat, pv lua.LValue) {
	if err := format.Validate(); err != nil {
		x.l.ArgError(2, err.Error())
	}
	luaCheckNumberOrNil(x.l, 4)
	if pv == lua.LNil {
		x.Err(fmt.Sprintf("write32: cmd=%d: нет значения", cmd))
		return
	}
	v := pv.(lua.LNumber)
	x.journalResult(
		fmt.Sprintf("write32: cmd=%d arg=%v format=%s", cmd, v, format),
		modbus.RequestWrite32{
			Addr:      x.p.Addr,
			ProtoCmd:  0x10,
			DeviceCmd: cmd,
			Format:    format,
			Value:     float64(v),
		}.GetResponse(log, x.l.Context(), x.comm()))
}

func (x luaProduct) ReadReg(reg modbus.Var, format modbus.FloatBitsFormat) lua.LValue {
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

func (x luaProduct) ReadKef(k modbus.Var, format modbus.FloatBitsFormat) lua.LValue {
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

func (x luaProduct) SetValue(key string, pv lua.LValue) {
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

func (x luaProduct) Value(key string) lua.LValue {
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

func (x luaProduct) Info(s string) {
	journal.Info(log, fmt.Sprintf("прибор %d.%d: %s", x.p.Serial, x.p.ProductID, s))
}

func (x luaProduct) Err(s string) {
	journal.Err(log, fmt.Errorf("прибор %d.%d: %s", x.p.Serial, x.p.ProductID, s))
}

func (x luaProduct) journalResult(s string, err error) {
	if err != nil {
		x.Err(fmt.Sprintf("%s: %s", s, err))
		return
	}
	x.Info(fmt.Sprintf("%s: успешно", s))
}

func (x luaProduct) comm() comm.T {
	return getCommProduct(x.p.Comport, x.Device)
}
