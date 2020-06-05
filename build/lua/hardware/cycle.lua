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

go:Info( { ["параметры"] =  params})

for n = 1,params.cycle_count do
    go:TemperatureSetup(params.temp_low)
    --tostring(temperature) .. '⁰C'
    go:Pause( string.format('Термоцикл %d: выдержка на низкой температуре: %g⁰C', n, params.temp_low))
    go:TemperatureSetup(params.temp_high)
    go:Pause( string.format('Термоцикл %d: выдержка на высокой температуре: %g⁰C', n, params.temp_high))
end
go:TemperatureStop()
