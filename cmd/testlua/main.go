package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yuin/gopher-lua"
	"layeh.com/gopher-luar"
)

type User struct {
	Name  string
	Age   int
	token string
}

func (u *User) SetToken(t string) {
	u.token = t
}

func (u *User) Token() string {
	return u.token
}

func main() {
	L := lua.NewState()
	defer L.Close()

	gasFunc := func(L *lua.LState) int {
		gas := L.ToInt(1)
		fmt.Println("set gas", gas)
		L.Push(lua.LNumber(11))
		L.Push(lua.LNumber(22))
		return 2
	}

	u := &User{
		Name: "Tim",
	}
	L.SetGlobal("u", luar.New(L, u))

	var mp map[string]interface{}
	err := json.Unmarshal([]byte(`{"Int":12, "Float":34.45, "String":"str"}`), &mp)
	if err != nil {
		panic(err)
	}

	L.SetGlobal("mp", luar.New(L, mp))
	L.SetGlobal("gas", L.NewFunction(gasFunc))

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	L.SetContext(ctx)

	if err := L.DoFile("main.lua"); err != nil {
		fmt.Printf("%+v\n", err)
	}

	fmt.Println("Lua set your token to:", u.Token())
	fmt.Printf("%+v\n", u)

	L.SetContext(context.Background())

	L.GetGlobal("works").(*lua.LTable).ForEach(func(value lua.LValue, value2 lua.LValue) {
		fmt.Println("call: ", value.String())
		if err := L.CallByParam(lua.P{
			Fn:      value2, // name of Lua function
			NRet:    0,      // number of returned values
			Protect: true,   // return err or panic
		}, lua.LNumber(12)); err != nil {
			panic(err)
		}
	})

}
