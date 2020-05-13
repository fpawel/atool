-- atool: work: Термокамера: термоциклирование

require 'utils/help'


local params = go:ParamsDialog({
    cycle_duration = {
        name = "Длительность выдержки на температуре",
        value = '1h',
        format = 'duration',
    },
    cycle_count = {
        name = "Количество циклов",
        value = 4,
        format = 'int',
    },
    temp_low = {
        name = "Низкая температура",
        value = -50,
        format = 'float',
    },
    temp_high = {
        name = "Высокая температура",
        value = 50,
        format = 'float',
    },
})

go:Info("параметры: " .. stringify(params))

for n = 1,params.cycle_count do
    go:TemperatureSetup(params.temp_low)
    go:Pause('Термоцикл ' .. tostring(n)..': '..format_temperature(params.temp_low))
    go:TemperatureSetup(params.temp_high)
    go:Pause('Термоцикл ' .. tostring(n)..': '..format_temperature(params.temp_high))
end
go:TemperatureStop()
