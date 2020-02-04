-- atoolgui: МИЛ82: Автоматическая настройка

require 'mil82/init'
require 'print_table'
json = require("dkjson")

prod_type = prod_types[go.Config.product_type]

local function stringify(v)
    return json.encode(v, { indent = true })
end

go:Info("конфигурация: " .. stringify(go.Config))

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
        for _,p in pairs(go.Products) do
            p:WriteKef(k, 'bcd', v)
            p:SetValue(db_key_coefficient(n), value)
        end
    end
end

go:Info("параметры: " .. stringify(params))

local current_temperature

local function format_temperature(temperature)
    return tostring(temperature) .. '⁰C'
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
    go:NewWork(what, function()
        go:SwitchGas(0, true)
        go:Temperature(temperature)
        current_temperature = temperature
    end)
end

local function gases_read_save(db_key_section, gases)
    go:NewWork("снятие " .. db_key_section .. ': газы: ' .. json.encode(gases, { indent = true }), function()
        for _, gas in ipairs(gases) do
            go:NewWork("снятие " .. db_key_section .. ': газ: ' .. tostring(gas), function()
                go:BlowGas(gas)
                for _, var in pairs(vars) do
                    go:ReadSave(var, 'bcd', db_key_section .. '_' .. db_key_gas_var(gas, var))
                end
            end)
        end
        go:BlowGas(1)
    end)

end

local function temperature_compensation(pt_temp)
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
    return function()
        setupTemperature(temperature)
        gases_read_save(pt_temp, gases)
    end
end

local function adjust()
    go:NewWork("калибровка нуля", function()
        go:BlowGas(1)
        go:Write32(1, 'bcd', go.Config.c1)
    end)
    go:NewWork("калибровка чувствительности", function()
        go:BlowGas(4)
        go:Write32(2, 'bcd', go.Config.c4)
        go:BlowGas(1)
    end)
end

local function temp_db_key(pt_temp, gas, var)
    return pt_temp .. '_gas' .. tostring(gas) .. '_var' .. tostring(var)
end

local write_coefficients_product = function(product, k, values)
    for i,value in pairs(values) do
        local n = k + i - 1
        local s = string.format('K%02d', n)
        if value ~= value then
            product:Err( s..': нет значения для записи: NaN')
        elseif value == nil then
            product:Err( s..': нет значения для записи: nil')
        else
            product:WriteKef(n, 'bcd', value)
            product:SetValue(db_key_coefficient(n), value)
        end
    end
end

local function get_temp_values_product(product, gas, var)
    local values = {}
    for _, pt_t in pairs({ pt_temp_low, pt_temp_norm, pt_temp_high }) do
        local key = temp_db_key(pt_t, gas, var)
        local value = product:Value(key)
        if value == nil then
            value = 0 / 0
        end
        table.insert(values, value)
    end
    return values
end

local function format_product_number(p)
    return string.format('№%d.id%d', p.Serial, p.ID)
end

local function calc_T0_product(p)
    go:NewWork( string.format('%s: расчёт термокомпенсации начала шкалы', format_product_number(p) ), function()
        local t1 = get_temp_values_product(p, 1, varTemp)
        local var1 = get_temp_values_product(p,1, var16)

        p:Info('t1='..stringify(t1))
        p:Info('var1='.. stringify(var1))

        local d1 = {}
        for i = 1, 3 do
            table.insert(d1, { t1[i], -var1[i] })
        end

        local T0 = go:InterpolationCoefficients(d1)
        p:Info('T0=' .. stringify(T0))
        write_coefficients_product(p, 23, T0)
    end)
end

local function calc_TK_product(p)
    go:NewWork( string.format('%s: расчёт термокомпенсации конца шкалы', format_product_number(p) ), function()
        local t4 = get_temp_values_product(p,4, varTemp)
        local var4 = get_temp_values_product(p,4, var16)
        local var1 = get_temp_values_product(p,1, var16)

        p:Info('t1='..stringify(t4))
        p:Info('var1='.. stringify(var4))

        local d4 = {}
        for i = 1, 3 do
            table.insert(d4, { t4[i], (var4[2] - var1[2]) / (var4[i] - var1[i]) })
        end

        local TK = go:InterpolationCoefficients(d4)

        p:Info('TK: ' .. stringify(TK))
        write_coefficients_product(p, 26, TK)
    end)
end

local function calc_TM_product(p)
    go:NewWork( string.format('%s: расчёт термокомпенсации середины шкалы', format_product_number(p) ), function()
        local K = {}
        for i = 16,19 do
            K[i] = p:Value(db_key_coefficient(n))
        end
    end)
end

local function calc_lin()
    for _, p in pairs(go.Products) do
        go:NewWork( string.format('%s: расчёт линеаризации', format_product_number(p) ), function()
            local xy = {}
            local gases = { 1, 2, 3, 4 }
            if params.linear_degree == 3 then
                gases = { 1, 3, 4 }
            end
            for _, gas in pairs(gases) do
                local x = p:Value('lin' .. tostring(gas))
                if x == nil then return end
                xy[gas] = { x, go.Config['c' .. tostring(gas)] }
            end

            local LIN = go:InterpolationCoefficients(xy)
            if params.linear_degree == 3 then
                LIN[4] = 0
            end
            p:Info(stringify(xy) .. ': ' .. stringify(LIN))
            write_coefficients_product(p, 16, LIN)
        end)
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
        local gases = { 3, 4 }
        if params.linear_degree == 4 then
            gases = { 2, 3, 4 }
        end
        for _, gas in pairs(gases) do
            go:BlowGas(gas)
            go:ReadSave(varConcentration, 'bcd', 'lin' .. tostring(gas))
        end
        go:BlowGas(1)
    end },

    { "расчёт линеаризации", calc_lin },

    { format_temperature(params.temp_low) .. ": снятие термокомпенсации", temperature_compensation(pt_temp_low) },

    { format_temperature(params.temp_high) .. ": снятие термокомпенсации", temperature_compensation(pt_temp_high) },

    { format_temperature(params.temp_norm) .. ": снятие термокомпенсации", temperature_compensation(pt_temp_norm) },

    { "расчёт и ввод термокомпенсации", function()
        for _, p in pairs(go.Products) do
            calc_T0_product(p)
            calc_TK_product(p)
        end
    end },

    { "снятие сигналов каналов", function()
        for _, k in pairs({ 21, 22, 43, 44 }) do
            go:ReadSave(224 + k * 2, 'bcd', 'K' .. tostring(k))
        end
    end },

    { format_temperature(params.temp_norm) .. ": снятие для проверки погрешности", function()
        setupTemperature(params.temp_norm)
        adjust()
        gases_read_save('test_' .. pt_temp_norm, { 1, 4 })
    end },

    { format_temperature(params.temp_low) .. ": снятие для проверки погрешности", function()
        setupTemperature(params.temp_low)
        gases_read_save('test_' .. pt_temp_low, { 1, 4 })
    end },

    { format_temperature(params.temp_high) .. ": снятие для проверки погрешности", function()
        setupTemperature(params.temp_high)
        gases_read_save('test_' .. pt_temp_high, { 1, 4 })
    end },

    { format_temperature(params.temp_norm) .. ": повторное снятие для проверки погрешности", function()
        setupTemperature(params.temp_norm)
        gases_read_save('test2_' .. pt_temp_norm, { 1, 4 })
    end },

    { "технологический прогон", function()
        adjust()
        go:Info('снятие перед технологическим прогоном')
        gases_read_save('tex1', { 1, 4 })
        go:Delay(params.duration_tex, 'технологический прогон')
        go:Info('снятие после технологического прогона')
        gases_read_save('tex2', { 1, 4 })
    end },
})
