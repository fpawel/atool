require "header"
require "print_table"
u.Name = "Вася"
u.Age = 12
print("Hello from Lua, " .. u.Name .. "!")
u:SetToken("12345",556)
print("u:Func2():", u:Func2())
local x1,x2 = gas(55)
print(x1,x2)
print("hx=",hx)
print('u', u)
print('Int',mp.Int)
print('Float',mp.Float)
print('String',mp.String)
print("sleeping...")
sleep(120)