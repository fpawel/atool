-- atoolgui: МИЛ82: Автоматическая настройка

require 'mil82/init'
json = require 'json'

prod_type = prod_types[go.Config.product_type]

go:Info("конфигурация: " .. json.encode(go.Config))

local params = go:ParamsDialog({
    linear_degree = {
        name = "Степень линеаризации",
        value = 4,
        format = 'int',
        list = { '3', '4' },
    },
    temp_middle_scale = {
        name = "Термокомпенсация: середина шкалы",
        value = true,
        format = 'bool',
    },
    temp_norm = {
        name = "Уставка температуры НКУ,⁰C",
        value = 20,
        format = 'float',
    },
    temp_low = {
        name = "Уставка низкой температуры,⁰C",
        value = prod_type.temp_low,
        format = 'float',
    },
    temp_high = {
        name = "Уставка высокой температуры,⁰C",
        value = prod_type.temp_high,
        format = 'float',
    },

    duration_tex = {
        name = "Длительность технологического прогона",
        value = '16h',
        format = 'duration',
    },
})

local function write_common_coefficients()
    for k, v in pairs({
        [2] = os.date("*t").year,
        [10] = go.Config.c1,
        [11] = go.Config.c3,
        [7] = scale_code[prod_type.scale],
        [8] = prod_type.scale_begin or 0,
        [9] = prod_type.scale,
        [5] = units_code[prod_type.gas],
        [6] = prod_type.gas,

        [16] = 0,
        [17] = 1,
        [18] = 0,
        [19] = 0,
        [23] = 0,
        [24] = 0,
        [25] = 0,
        [26] = 0,
        [27] = 0,
        [28] = 0,
    }) do
        go:WriteKef(k, 'bcd', v)
    end
end

function lin_calc()
    for _, p in pairs(go.Products) do
        local ct = {}
        for i = 1, 4 do
            if not (params.linear_degree == 3 and i == 2) then
                local x = p:Value('lin' .. tostring(i))
                if x == nil then
                    p:Err('расёт линеаризатора не выполнен: нет значения lin' .. tostring(i))
                    return
                end
                ct[i] = { x, go.Config['c' .. tostring(i)] }
            end
        end

        local cf = go:InterpolationCoefficients(ct)
        if cf == nil then
            p:Err('расёт линеаризатора не выполнен: ' .. json.encode(ct))
            return
        end
        p:Info('расчёт линеаризатора: ' .. json.encode(ct) .. ': ' .. json.encode(cf))
        if params.linear_degree == 3 then
            cf[4] = 0
        end
        for i = 1, 4 do
            p:WriteKef(15 + i, 'bcd', cf[i])
        end
    end
end

go:Info("параметры: " .. json.encode(params))

local current_temperature

local function format_temperature(temperature)
    return tostring(temperature)..'⁰C'
end

local function setupTemperature(temperature)
    local what = 'перевод термокамеры: '
    if current_temperature == nil then
        what = what .. format_temperature(temperature)
    else
        what = what .. format_temperature(current_temperature) .. ' -> ' .. format_temperature(temperature)
    end
    if current_temperature == temperature then
        go:Info(what .. ': температура уже установлена')
        return
    end
    go:NewWork(what, function ()
        go:SwitchGas(0, true)
        go:Temperature(temperature)
        current_temperature = temperature
    end)
end

local function gases_read_save( db_key_section, gases )
    go:NewWork("снятие " .. db_key_section .. ': газы: '..json.encode(gases), function ()
        for _, gas in ipairs(gases) do
            go:NewWork("снятие " .. db_key_section .. ': газ: '..tostring(gas), function ()
                go:BlowGas(gas)
                for _, var in pairs(vars) do
                    go:ReadSave(var, 'bcd', db_key_section..'_' .. db_key_gas_var(gas, var))
                end
            end)
        end
        go:BlowGas(1)
    end)

end

local function temperature_compensation(pt_temp )
    local gases = { 1, 4 }
    if params.temp_middle_scale then
        gases = { 1, 3, 4 }
    end
    local temperatures = {
        [pt_temp_norm] = params.temp_norm,
        [pt_temp_low] = params.temp_low,
        [pt_temp_high] = params.temp_high,
    }
    local temperature = temperatures[pt_temp]
    return function ()
        setupTemperature(temperature)
        gases_read_save(pt_temp, gases)
    end
end

local function adjust()
    go:NewWork("калибровка нуля", function ()
        go:BlowGas(1)
        go:Write32(1, 'bcd', go.Config.c1)
    end)
    go:NewWork("калибровка чувствительности", function ()
        go:BlowGas(4)
        go:Write32(2, 'bcd', go.Config.c4)
        go:BlowGas(1)
    end)
end

local function calc_temp()

    local function temp_db_key( pt_temp, gas, var)
        return pt_temp .. '_gas' .. tostring(gas) .. '_var' .. tostring(var)
    end

    for _, p in pairs(go.Products) do

        local val = function(db_key, gas, var)
            return p:Value(temp_db_key(db_key, gas, var))
        end

        local write_coefficients = function (k, values)
            for i = 1, 3 do
                p:WriteKef(k + i - 1, 'bcd', values[i])
            end
        end

        local y1 = {
            val(pt_temp_low, 1, var16),
            val(pt_temp_norm, 1, var16),
            val(pt_temp_high, 1, var16),
        }

        write_coefficients(23, go:InterpolationCoefficients({
            { val(pt_temp_low, 1, varTemp), -y1[1] },
            { val(pt_temp_norm, 1, varTemp), -y1[2] },
            { val(pt_temp_high, 1, varTemp), -y1[3] },
        }))

        local xs3 = {
            {
                val(pt_temp_low, 4, varTemp),
                val(pt_temp_low, 4, var16) - y1[1],
            },
            {
                val(pt_temp_norm, 4, varTemp),
                val(pt_temp_norm, 4, var16) - y1[2],
            },
            {
                val(pt_temp_high, 4, varTemp),
                val(pt_temp_high, 4, var16) - y1[3],
            },
        }
        for i = 1, 3 do
            xs3[i][2] = xs3[2][2] / xs3[i][2]
        end

        write_coefficients(23, go:InterpolationCoefficients(xs3))
    end

end

local function calc_lin()
    for _, p in pairs(go.Products) do
        local ct = {}
        local gases = {1,2,3,4}
        if params.linear_degree == 3 then
            gases = {1,3,4}
        end
        for _,gas in pairs(gases) do
            local x = p:Value('lin' .. tostring(gas))
            if x == nil then
                p:Err('расёт линеаризатора не выполнен: нет значения lin' .. tostring(gas))
                return
            end
            ct[gas] = { x, go.Config['c' .. tostring(gas)] }
        end

        local cf = go:InterpolationCoefficients(ct)
        if cf == nil then
            p:Err('расёт линеаризатора не выполнен: ' .. json.encode(ct))
            return
        end
        p:Info('расчёт линеаризатора: ' .. json.encode(ct) .. ': ' .. json.encode(cf))
        if params.linear_degree == 3 then
            cf[4] = 0
        end
        for i = 1, 4 do
            p:WriteKef(15 + i, 'bcd', cf[i])
        end
    end
end

go:SelectWorksDialog({
    { "запись коэффициентов", write_common_coefficients },

    { "установка НКУ", function()
        setupTemperature(params.temp_norm)
    end },

    { "нормировка", function()
        go:Write32(8, 'bcd', 1000)
    end },

    { "калибровка", adjust },

    { "снятие линеаризации", function()
        go:ReadSave(varConcentration, 'bcd', 'lin1')
        local gases = {3,4}
        if params.linear_degree == 4 then
            gases = {2, 3,4}
        end
        for _,gas in pairs(gases) do
            go:BlowGas(gas)
            go:ReadSave(varConcentration, 'bcd', 'lin'..tostring(gas))
        end
        go:BlowGas(1)
    end },

    { "расчёт линеаризации", calc_lin },

    { format_temperature(params.temp_low)..": снятие термокомпенсации", temperature_compensation(pt_temp_low) },

    { format_temperature(params.temp_high)..": снятие термокомпенсации", temperature_compensation(pt_temp_high)},

    { format_temperature(params.temp_norm)..": снятие термокомпенсации", temperature_compensation(pt_temp_norm) },

    { "расчёт и ввод термокомпенсации", calc_temp },

    { "снятие сигналов каналов", function()
        for _,k in pairs({21,22,43,44}) do
            go:ReadSave(224 + k*2, 'bcd', 'K'..tostring(k))
        end
    end },

    { format_temperature(params.temp_norm)..": снятие для проверки погрешности", function()
        setupTemperature(params.temp_norm)
        adjust()
        gases_read_save('test_'..pt_temp_norm, {1,4})
    end },

    { format_temperature(params.temp_low)..": снятие для проверки погрешности", function()
        setupTemperature(params.temp_low)
        gases_read_save('test_'..pt_temp_low, {1,4})
    end },

    { format_temperature(params.temp_high)..": снятие для проверки погрешности", function()
        setupTemperature(params.temp_high)
        gases_read_save('test_'..pt_temp_high, {1,4})
    end },

    { format_temperature(params.temp_norm)..": повторное снятие для проверки погрешности", function()
        setupTemperature(params.temp_norm)
        gases_read_save('test2_'..pt_temp_norm, {1,4})
    end },

    { "технологический прогон", function()
        adjust()
        go:Info('снятие перед технологическим прогоном')
        gases_read_save('tex1', {1,4})
        go:Delay(params.duration_tex, 'технологический прогон')
        go:Info('снятие после технологического прогона')
        gases_read_save('tex2', {1,4})
    end },
})
