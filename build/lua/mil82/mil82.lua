-- atool: work: МИЛ-82: автоматическая настройка

require 'utils/help'
require 'utils/temp_setup'
require 'utils/print_table'
require 'mil82/def'

print(stringify({y=12, x = 0/0}))

local Products = go:GetProducts()

local Config = go:GetConfig()
go:Info("конфигурация: "..stringify(Config))

local prod_type = prod_types[Config.product_type]
--print_table(prod_type)
go:Info("исполнение: "..stringify(prod_type))

if prod_type == nil then
    error('не определено исполнение: '..Config.product_type)
end

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

    local scale_code = scale_code[prod_type.scale]

    if scale_code == nil then
        go:Err('не определён код шкалы: '..tostring(prod_type.scale))
    end

    local units_code = units_code[prod_type.gas]

    if units_code == nil then
        error('не определён код единиц измерения: '..tostring(prod_type.gas))
    end

    local coefficients = {
        [2] = os.date("*t").year,
        [10] = Config.c1,
        [11] = Config.c3,
        [7] = scale_code,
        [8] = prod_type.scale_begin or 0,
        [9] = prod_type.scale,
        [5] = units_code,
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
        [37] = 0,
        [38] = 0,
        [39] = 0,
    }

    for _, p in pairs(Products) do
        set_coefficients_product(coefficients, p)
        write_coefficients_product(coefficients, p)
    end
end

go:Info("параметры: " .. stringify(params))

local function gases_read_save(db_key_section, gases)
    go:NewWork("снятие " .. db_key_section .. ': газы: ' .. json.encode(gases, { indent = true }), function()
        for _, gas in ipairs(gases) do
            go:NewWork("снятие " .. db_key_section .. ': газ: ' .. tostring(gas), function()
                go:BlowGas(gas)
                for _, var in pairs(vars) do
                    go:ReadSave(var, 'bcd', db_key_section .. '_' .. mil82_db_key_gas_var(gas, var))
                end
            end)
        end
        go:BlowGas(1)
    end)
end

local function temp_comp(pt_temp)
    local gases = { 1, 4 }
    if params.temp_middle_scale then
        gases = { 1, 3, 4 }
    end
    local temperatures = {
        [t_norm] = params.temp_norm,
        [t_low] = params.temp_low,
        [t_high] = params.temp_high,
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
        go:Write32(1, 'bcd', Config.c1)
    end)
    go:NewWork("калибровка чувствительности", function()
        go:BlowGas(4)
        go:Write32(2, 'bcd', Config.c4)
        go:BlowGas(1)
    end)
end


local function calc_lin()
    for _, p in pairs(Products) do
        go:NewWork(string.format('%s: расчёт линеаризации', format_product_number(p)), function()
            local xy = {}
            local gases = { 1, 2, 3, 4 }
            if params.linear_degree == 3 then
                gases = { 1, 3, 4 }
            end
            for _, gas in pairs(gases) do
                local x = p:Value('lin' .. tostring(gas))
                if x == nil then
                    return
                end
                xy[gas] = { x, Config['c' .. tostring(gas)] }
            end

            local LIN = go:InterpolationCoefficients(xy)
            if params.linear_degree == 3 then
                LIN[4] = 0
            end
            p:Info(stringify(xy) .. ': ' .. stringify(LIN))
            set_coefficients_product(array_n(LIN, 16), p)
        end)
    end
end

local function temp_db_key(pt_temp, gas, var)
    return pt_temp .. '_gas' .. tostring(gas) .. '_var' .. tostring(var)
end

local function get_temp_values_product(product, gas, var)
    local values = {}
    for _, pt_t in pairs({ t_low, t_norm, t_high }) do
        local key = temp_db_key(pt_t, gas, var)
        local value = product:Value(key)
        table.insert(values, value)
    end
    return values
end


local function calc_T0_product(p)
    go:NewWork(string.format('%s: расчёт термокомпенсации начала шкалы', format_product_number(p)), function()
        local t1 = get_temp_values_product(p, 1, varTemp)
        local var1 = get_temp_values_product(p, 1, var16)

        p:Info('t1=' .. stringify(t1))
        p:Info('var1=' .. stringify(var1))

        local d1 = {}
        for i = 1, 3 do
            table.insert(d1, { t1[i], -var1[i] })
        end

        local T0 = go:InterpolationCoefficients(d1)
        p:Info('T0=' .. stringify(T0))
        set_coefficients_product(array_n(T0, 23), p)
    end)
end

local function calc_TK_product(p)
    go:NewWork(string.format('%s: расчёт термокомпенсации конца шкалы', format_product_number(p)), function()
        local t4 = get_temp_values_product(p, 4, varTemp)
        local var4 = get_temp_values_product(p, 4, var16)
        local var1 = get_temp_values_product(p, 1, var16)

        p:Info('t1=' .. stringify(t4))
        p:Info('var1=' .. stringify(var4))

        local d4 = {}
        for i = 1, 3 do
            table.insert(d4, { t4[i], (var4[2] - var1[2]) / (var4[i] - var1[i]) })
        end

        local TK = go:InterpolationCoefficients(d4)
        p:Info('TK: ' .. stringify(TK))
        set_coefficients_product(array_n(TK, 26), p)
    end)
end

local function calc_TM_product(p)
    go:NewWork(string.format('%s: расчёт термокомпенсации середины шкалы', format_product_number(p)), function()

        local C4 = Config.c4;

        local K16 = p:Kef(16)
        local K17 = p:Kef(17)
        local K18 = p:Kef(18)
        local K19 = p:Kef(19)

        local v1_norm = p:Value(temp_db_key(t_norm, 1, var16))
        local v3_norm = p:Value(temp_db_key(t_norm, 3, var16))
        local v4_norm = p:Value(temp_db_key(t_norm, 4, var16))

        local v1_low = p:Value(temp_db_key(t_low, 1, var16))
        local v3_low = p:Value(temp_db_key(t_low, 3, var16))
        local v4_low = p:Value(temp_db_key(t_low, 4, var16))

        local v1_high = p:Value(temp_db_key(t_high, 1, var16))
        local v3_high = p:Value(temp_db_key(t_high, 3, var16))
        local v4_high = p:Value(temp_db_key(t_high, 4, var16))

        local x1 = C4 * (v1_norm - v3_norm) / (v1_norm - v4_norm)
        local x2 = C4 * (v1_low - v3_low) / (v1_low - v4_low)
        local d = K16 + K17 * x2 + K18 * x2 * x2 + K19 * x2 * x2 * x2 - x2

        local y_low = (K16 + K17 * x1 + K18 * x1 * x1 + K19 * x1 * x1 * x1 - x2) / d

        x1 = C4 * (v1_norm - v3_norm) / (v1_norm - v4_norm)
        x2 = C4 * (v1_high - v3_high) / (v1_high - v4_high)

        d = K16 + K17 * x2 + K18 * x2 * x2 + K19 * x2 * x2 * x2 - x2

        local y_hi = (K16 + K17 * x1 + K18 * x1 * x1 + K19 * x1 * x1 * x1 - x2) / d

        local t1 = p:Value(temp_db_key(t_low, 3, varTemp))
        local t2 = p:Value(temp_db_key(t_norm, 3, varTemp))
        local t3 = p:Value(temp_db_key(t_high, 3, varTemp))

        local TM = go:InterpolationCoefficients({
            { t1, y_low },
            { t2, 1 },
            { t3, y_hi },
        })

        p:Info('TM: ' .. stringify(TM))
        set_coefficients_product(array_n(TM, 37), p)

    end)
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

    { "запись линеаризации", function()
        write_coefficients({ 16, 17, 18, 19 })
    end },

    { format_temperature(params.temp_low) .. ": снятие термокомпенсации", temp_comp(t_low) },

    { format_temperature(params.temp_high) .. ": снятие термокомпенсации", temp_comp(t_high) },

    { format_temperature(params.temp_norm) .. ": снятие термокомпенсации", temp_comp(t_norm) },

    { "расчёт термокомпенсации", function()
        for _, p in pairs(Products) do
            calc_T0_product(p)
            calc_TK_product(p)
            if params.temp_middle_scale then
                calc_TM_product(p)
            end
        end
    end },

    { "запись термокомпенсации", function ()
        write_coefficients({ 23, 24, 25, 26, 27, 28, 37, 38, 39 })
    end },

    { "снятие сигналов каналов", function()
        for _, k in pairs({ 21, 22, 43, 44 }) do
            go:ReadSave(224 + k * 2, 'bcd', 'K' .. tostring(k))
        end
    end },

    { format_temperature(params.temp_norm) .. ": снятие для проверки погрешности", function()
        setupTemperature(params.temp_norm)
        adjust()
        gases_read_save('test_' .. t_norm, { 1, 4 })
    end },

    { format_temperature(params.temp_low) .. ": снятие для проверки погрешности", function()
        setupTemperature(params.temp_low)
        gases_read_save('test_' .. t_low, { 1, 4 })
    end },

    { format_temperature(params.temp_high) .. ": снятие для проверки погрешности", function()
        setupTemperature(params.temp_high)
        gases_read_save('test_' .. t_high, { 1, 4 })
    end },

    { format_temperature(80) .. ": снятие для проверки погрешности", function()
        setupTemperature(80)
        gases_read_save('test_t80', { 1, 4 })
    end },

    { format_temperature(params.temp_norm) .. ": повторное снятие для проверки погрешности", function()
        setupTemperature(params.temp_norm)
        gases_read_save('test2_', { 1, 4 })
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
