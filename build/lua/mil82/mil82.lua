-- atool: work: МИЛ-82: автоматическая настройка
require 'utils/help'

local var16 = 16
local var_temp = 2
local var_concentration = 0

local FLOAT_FORMAT = 'bcd'

local vars = { var_concentration, var_temp, 4, 8, 10, 12, 14, var16 }

local function read_save_section_gases(db_key_section, gases)
    go:Perform("снятие " .. db_key_section .. ': газы: ' .. go:Stringify(gases), function()
        for _, gas in ipairs(gases) do
            go:Perform("снятие " .. db_key_section .. ': газ: ' .. tostring(gas), function()
                go:BlowGas(gas)
                for _, var in pairs(vars) do
                    local db_key = db_key_section .. '_gas' .. tostring(gas) .. '_var' .. tostring(var)
                    go:ReadAndSaveParam(var, FLOAT_FORMAT, db_key)
                end
            end)
        end
        go:BlowGas(1)
    end)
end

local function adjust()
    go:Perform("калибровка нуля", function()
        go:BlowGas(1)
        go:Write32(1, FLOAT_FORMAT, go:GetConfig().c1)
    end)
    go:Perform("калибровка чувствительности", function()
        go:BlowGas(4)
        go:Write32(2, FLOAT_FORMAT, go:GetConfig().c4)
        go:BlowGas(1)
    end)
end



local function new_product(p)

    local ret = {}

    function ret.get_init_coefficients(prod_type)

        local scale_codes = {
            [4] = 57,
            [10] = 7,
            [20] = 9,
            [50] = 0,
            [100] = 21,
        }
        local scale_code = scale_codes[prod_type.scale]

        if scale_code == nil then
            error('не определён код шкалы')
        end

        local unit_codes = {
            ['CO2'] = 7,
            ['CH4'] = 14,
            ['C3H8'] = 14,
            ['C6H14'] = 14,
        }

        local units_code = unit_codes[prod_type.gas]

        if units_code == nil then
            error('не определён код единиц измерения')
        end

        local gas_codes = {
            CO2 = 4,
            CH4 = 5,
            C3H8 = 7,
            C6H14 = 7,
        }
        local gas_code = gas_codes[prod_type.gas]

        if gas_code == nil then
            error('не определён код газа')
        end

        local function encode2(a, b)
            return a * 10000 + b
        end

        local date = os.date("*t")

        local cfg = go:GetConfig()

        local coefficients = {
            [2] = date.year,
            [10] = cfg.c1,
            [11] = cfg.c4,
            [7] = scale_code,
            [8] = prod_type.scale_begin or 0,
            [9] = prod_type.scale,
            [5] = units_code,
            [6] = gas_code,

            [16] = 0,
            [17] = 1,
            [18] = 0,
            [19] = 0,
            [23] = 0,
            [24] = 0,
            [25] = 0,
            [26] = 1,
            [27] = 0,
            [28] = 0,
            [37] = 1,
            [38] = 0,
            [39] = 0,

            [40] = encode2(date.year - 2000, p.Serial),
            [47] = encode2(date.month, prod_type.index)
        }
        for k, v in pairs(prod_type.coefficient) do
            coefficients[k] = v
        end
    end

    return ret
end

local function lin_calc_write(linear_degree)
    go:Interpolation('LIN', 16, 4, FLOAT_FORMAT, function(getValue)
        local xy = {}
        local gases = { 1, 2, 3, 4 }
        if linear_degree == 3 then
            gases = { 1, 3, 4 }
        end
        local cfg = go:GetConfig()
        for _, gas in pairs(gases) do
            local x = getValue('lin' .. tostring(gas))
            local y = cfg['c' .. tostring(gas)]
            xy[gas] = { x, y }
        end
        return xy
    end )
end

local function t_val(getValue, pt_temp, gas, var)
    return getValue(pt_temp .. '_gas' .. tostring(gas) .. '_var' .. tostring(var))
end

local function t_values(getValue, gas, var)
    local values = {}
    for _, pt_t in pairs({ 't_low', 't_norm', 't_high' }) do
        local value = t_val(getValue, pt_t, gas, var)
        table.insert(values, value)
    end
    return values
end

local function T0_calc_write()
    go:Interpolation('T0', 23, 3, FLOAT_FORMAT, function(getValue)
        local t1 = t_values(getValue, 1, var_temp)
        local var1 = t_values(getValue, 1, var16)
        local d1 = {}
        for i = 1, 3 do
            table.insert(d1, { t1[i], -var1[i] })
        end
        return d1
    end )
end

local function TK_calc_write()
    go:Interpolation('TK', 26, 3, FLOAT_FORMAT, function(getValue)
        local t4 = t_values(getValue, 4, var_temp)
        local var4 = t_values(getValue, 4, var16)
        local var1 = t_values(getValue, 1, var16)
        local d4 = {}
        for i = 1, 3 do
            table.insert(d4, { t4[i], (var4[2] - var1[2]) / (var4[i] - var1[i]) })
        end
        return d4
    end )
end

local function  TM_calc_write()
    go:Interpolation('TM', 37, 3, FLOAT_FORMAT, function(getValue)
        local C4 = go:GetConfig().c4
        local K16 = getValue('K'..tostring(16))
        local K17 = getValue('K'..tostring(17))
        local K18 = getValue('K'..tostring(18))
        local K19 = getValue('K'..tostring(19))

        local v1_norm = t_val(getValue,'t_norm', 1, var16)
        local v3_norm = t_val(getValue,'t_norm', 3, var16)
        local v4_norm = t_val(getValue,'t_norm', 4, var16)

        local v1_low = t_val(getValue,'t_low', 1, var16)
        local v3_low = t_val(getValue,'t_low', 3, var16)
        local v4_low = t_val(getValue,'t_low', 4, var16)

        local v1_high = t_val(getValue,'t_high', 1, var16)
        local v3_high = t_val(getValue,'t_high', 3, var16)
        local v4_high = t_val(getValue,'t_high', 4, var16)

        local x1 = C4 * (v1_norm - v3_norm) / (v1_norm - v4_norm)
        local x2 = C4 * (v1_low - v3_low) / (v1_low - v4_low)
        local d = K16 + K17 * x2 + K18 * x2 * x2 + K19 * x2 * x2 * x2 - x2

        local y_low = (K16 + K17 * x1 + K18 * x1 * x1 + K19 * x1 * x1 * x1 - x2) / d

        x1 = C4 * (v1_norm - v3_norm) / (v1_norm - v4_norm)
        x2 = C4 * (v1_high - v3_high) / (v1_high - v4_high)

        d = K16 + K17 * x2 + K18 * x2 * x2 + K19 * x2 * x2 * x2 - x2

        local y_hi = (K16 + K17 * x1 + K18 * x1 * x1 + K19 * x1 * x1 * x1 - x2) / d

        local t1 = t_val(getValue,'t_low', 3, var_temp)
        local t2 = t_val(getValue,'t_norm', 3, var_temp)
        local t3 = t_val(getValue,'t_high', 3, var_temp)

        return {
            { t1, y_low },
            { t2, 1 },
            { t3, y_hi },
        }
    end )
end

local cfg = go:GetConfig()
go:Info("конфигурация", cfg)

local prod_types = (require('mil82/types'))
local product_type_name = go:GetConfig().product_type
local prod_type = prod_types[product_type_name]
if prod_type == nil then
    error('МИЛ82: не определено исполнение ' .. product_type_name)
end

go:Info("исполнение", prod_type)

local params = go:ParamsDialog({
    linear_degree = {
        name = "Степень линеаризации",
        value = 4,
        format = 'int',
        list = { '3', '4' },
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
    temp90 = {
        name = "Уставка 90⁰C",
        value = 90,
        format = 'float',
    },
})
go:Info("параметры", params)

local function temp_comp(pt_temp)
    local temperatures = {
        t_norm = params.temp_norm,
        t_low = params.temp_low,
        t_high = params.temp_high,
    }
    local temperature = temperatures[pt_temp]
    return function()
        go:Temperature(temperature)
        read_save_section_gases(pt_temp, { 1, 3, 4 })
    end
end

local function read_lin()
    go:ReadAndSaveParam(var_concentration, FLOAT_FORMAT, 'lin1')
    local gases = { 3, 4 }
    if params.linear_degree == 4 then
        gases = { 2, 3, 4 }
    end
    for _, gas in pairs(gases) do
        go:BlowGas(gas)
        go:ReadAndSaveParam(var_concentration, FLOAT_FORMAT, 'lin' .. tostring(gas))
    end
    go:BlowGas(1)
end

local works = {
    go:WorkEachSelectedProduct("Запись коэффициентов", function(p)
        local xs = new_product(p).get_init_coefficients(prod_type)
        p:Info(xs)
        p:WriteCoefficients(xs, FLOAT_FORMAT)
    end),

    go:Work("Установка НКУ", function()
        go:Temperature(params.temp_norm)
    end),

    go:Work("Нормировка", function()
        go:BlowGas(1)
        go:Write32(8, FLOAT_FORMAT, 1000)
    end),

    go:Work("Калибровка", adjust),

    go:Work("Снятие линеаризации", read_lin),

    go:Work("Расчёт и запись линеаризации", function()
        lin_calc_write(params.linear_degree)
    end),

    go:Work(string.format("Снятие термокомпенсации Т-: %g⁰C", params.temp_low), temp_comp('t_low')),

    go:Work(string.format("Снятие термокомпенсации Т+: %g⁰C", params.temp_high), temp_comp('t_high')),

    go:Work(string.format("Снятие термокомпенсации НКУ: %g⁰C", params.temp_norm), temp_comp('t_norm')),

    go:WorkEachSelectedProduct("Расчёт и запись термокомпенсации", function(product)
        local p = new_product(product)
        product:Perform("T0 начало шкалы", T0_calc_write)
        product:Perform("TK конец шкалы", TK_calc_write)
        product:Perform("TM середина шкалы", TM_calc_write)
    end),

    go:Work("снятие сигналов каналов", function()
        go:ReadCoefficients({ 20, 21, 43, 44 }, FLOAT_FORMAT)
    end),

    go:Work("НКУ: снятие для проверки погрешности", function()
        go:Temperature(params.temp_norm)
        adjust()
        read_save_section_gases('test_t_norm', { 1, 2, 3, 4 })
    end),

    go:Work(string.format("Т-: снятие для проверки погрешности: %g⁰C", params.temp_low), function()
        go:Temperature(params.temp_low)
        read_save_section_gases('test_t_low', { 1, 3, 4 })
    end),

    go:Work(string.format("Т+: снятие для проверки погрешности: %g⁰C", params.temp_high), function()
        go:Temperature(params.temp_high)
        read_save_section_gases('test_t_high', { 1, 3, 4 })
    end),

    go:Work(string.format("80⁰C: снятие для проверки погрешности: %g⁰C", params.temp90), function()
        go:Temperature(params.temp90)
        read_save_section_gases('test_t80', { 1, 3, 4 })
    end),

    go:Work("НКУ: повторное снятие для проверки погрешности", function()
        go:Temperature(params.temp_norm)
        read_save_section_gases('test2', { 1, 3, 4 })
    end),
}
return go:PerformWorks(go:SelectWorksDialog(works))


