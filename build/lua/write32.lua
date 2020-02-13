-- atool: work: MODBUS: Отправка команды

params =  go:ParamsDialog({
    cmd = {
        name = "Код команды",
        value = 1,
        type = 'int',
    },
    format = {
        name = "Формат",
        value = 'bcd',
        list = {
            "bcd", "float_big_endian", "float_little_endian",
        }
    },
    value = {
        name = "Аргумент",
        type = 'float',
        value = 1,
    },
})

for _,p in pairs(go:GetProducts()) do
    p:Write32(params.cmd, params.format, params.value)
end