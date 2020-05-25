package luadata

import (
	"github.com/ansel1/merry"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"os"
	"path/filepath"
)

var (
	Devices  = make(map[string]*Device)
	luaState = lua.NewState()
)

func init() {
	filename := filepath.Join(filepath.Dir(os.Args[0]), "lua", "devices.lua")
	luaState.SetGlobal("Device", luar.New(luaState, func(name string) *Device {
		Devices[name] = new(Device)
		return Devices[name]
	}))

	if err := luaState.DoFile(filename); err != nil {
		panic(merry.Errorf("devices.lua: %w", err))
	}
}
