require './sleep' -- включить файл sleep.lua

print(#go.Config)
for k,v in ipairs(go.Config) do
    print(k .. "=",v)
end


go:PauseSec(1, "пауза 1c") -- пауза 1 c
print("пауза 1с")
sleep_sec(1) -- пауза 1 c

for _,p in pairs(go.Products) do -- для каждого прибора партии
    k99 = p:ReadKef(99,'bcd') -- считать коэффициент 99
    print('K99=', k99)

    p:Info("K99=" .. tostring(k99) ) -- информационное сообщение в журнал
    p:Err("error occurred") -- сообщение об ошибке в журнал

    invalidKeyValue = p:Value('invalid_key')
    print('invalid_key=', invalidKeyValue)
    p:SetValue('invalid_key', invalidKeyValue)

    p:DeleteKey('concentration1')
    p:SetValue('concentration1', 11.12)
    print('concentration1=', p:Value('concentration1'))

    p:WriteKef(100, 'bcd', p:Value('concentration1')) -- записать значение к-та 100

    p:SetValue('lin1', p:ReadReg(2, 'bcd')) -- считать из регистра 2 и сохранить в lin1
    p:SetValue('lin2', p:ReadReg(4, 'bcd'))
    p:SetValue('lin3', p:ReadReg(5, 'bcd'))

    p:SetValue('k12', p:ReadKef(12,'bcd')) -- считать к-т 12 и сохранить в k12

    p:WriteKef(12,'bcd', 100.1) -- отправка команды 12 с аргументом 100.1
end

go:ReadSave(8,'bcd', 'lin4') -- для каждого прибора партии считать из регистра 8 и сохранить в lin4
go:Gas(0) -- пневмоблок 0
go:Temperature(20) -- температура 20
