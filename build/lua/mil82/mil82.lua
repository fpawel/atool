-- atool: work: МИЛ-82: автоматическая настройка
require 'utils/help'

local var16 = 16
local var_temp = 2
local var_concentration = 0

local float_format = 'bcd'

local vars = { var_concentration, var_temp, 4, 8, 10, 12, 14, var16 }

local function work_gases_read_save(db_key_section, gases)
    go:NewWork("снятие " .. db_key_section .. ': газы: ' .. go:Stringify(gases), function()
        for _, gas in ipairs(gases) do
            go:NewWork("снятие " .. db_key_section .. ': газ: ' .. tostring(gas), function()
                go:BlowGas(gas)
                for _, var in pairs(vars) do
                    local db_key = db_key_section .. '_gas' .. tostring(gas) .. '_var' .. tostring(var)
                    go:ReadAndSaveParam(var, float_format, db_key)
                end
            end)
        end
        go:BlowGas(1)
    end)
end

local params

local function adjust()
    go:NewWork("калибровка нуля", function()
        go:BlowGas(1)
        go:Write32(1, float_format, go:GetConfig().c1)
    end)
    go:NewWork("калибровка чувствительности", function()
        go:BlowGas(4)
        go:Write32(2, float_format, go:GetConfig().c4)
        go:BlowGas(1)
    end)
end

local function mil82_product(p)

    local ret = {}

    function ret.init(prod_type)

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
        p:WriteCoefficientValues(coefficients, float_format)
    end

    local function temp_val(pt_temp, gas, var)
        return p:Value(pt_temp .. '_gas' .. tostring(gas) .. '_var' .. tostring(var))
    end

    local function temp_values(gas, var)
        local values = {}
        for _, pt_t in pairs({ 't_low', 't_norm', 't_high' }) do
            local value = temp_val(pt_t, gas, var)
            table.insert(values, value)
        end
        return values
    end

    function ret.calc_lin()
        local xy = {}
        local gases = { 1, 2, 3, 4 }
        if params.linear_degree == 3 then
            gases = { 1, 3, 4 }
        end

        local cfg = go:GetConfig()

        for _, gas in pairs(gases) do
            local x = p:Value('lin' .. tostring(gas))
            if x == nil then
                return
            end
            xy[gas] = { x, cfg['c' .. tostring(gas)] }
        end
        p:Interpolation('LIN', xy, 16, 4, float_format)
    end

    function ret.calc_T0()
        local t1 = temp_values(1, var_temp)
        local var1 = temp_values(1, var16)

        p:Info('t1', t1, 'var1', var1)

        local d1 = {}
        for i = 1, 3 do
            table.insert(d1, { t1[i], -var1[i] })
        end

        p:Interpolation('T0', d1, 23, 3, float_format)
    end

    function ret.calc_TK()
        local t4 = temp_values(4, var_temp)
        local var4 = temp_values(4, var16)
        local var1 = temp_values(1, var16)

        p:Info('t4', t4, 'var4', var4, var1, var1)

        local d4 = {}
        for i = 1, 3 do
            table.insert(d4, { t4[i], (var4[2] - var1[2]) / (var4[i] - var1[i]) })
        end

        p:Interpolation('TK', d4, 26, 3, float_format)
    end

    function ret.calc_TM()
        local C4 = go:GetConfig().c4
        local K16 = p:Kef(16)
        local K17 = p:Kef(17)
        local K18 = p:Kef(18)
        local K19 = p:Kef(19)

        local v1_norm = temp_val('t_norm', 1, var16)
        local v3_norm = temp_val('t_norm', 3, var16)
        local v4_norm = temp_val('t_norm', 4, var16)

        local v1_low = temp_val('t_low', 1, var16)
        local v3_low = temp_val('t_low', 3, var16)
        local v4_low = temp_val('t_low', 4, var16)

        local v1_high = temp_val('t_high', 1, var16)
        local v3_high = temp_val('t_high', 3, var16)
        local v4_high = temp_val('t_high', 4, var16)

        local x1 = C4 * (v1_norm - v3_norm) / (v1_norm - v4_norm)
        local x2 = C4 * (v1_low - v3_low) / (v1_low - v4_low)
        local d = K16 + K17 * x2 + K18 * x2 * x2 + K19 * x2 * x2 * x2 - x2

        local y_low = (K16 + K17 * x1 + K18 * x1 * x1 + K19 * x1 * x1 * x1 - x2) / d

        x1 = C4 * (v1_norm - v3_norm) / (v1_norm - v4_norm)
        x2 = C4 * (v1_high - v3_high) / (v1_high - v4_high)

        d = K16 + K17 * x2 + K18 * x2 * x2 + K19 * x2 * x2 * x2 - x2

        local y_hi = (K16 + K17 * x1 + K18 * x1 * x1 + K19 * x1 * x1 * x1 - x2) / d

        local t1 = temp_val('t_low', 3, var_temp)
        local t2 = temp_val('t_norm', 3, var_temp)
        local t3 = temp_val('t_high', 3, var_temp)

        local dCalc = {
            { t1, y_low },
            { t2, 1 },
            { t3, y_hi },
        }

        p:Interpolation('TM', dCalc, 37, 3, float_format)
    end

    return ret
end

local function main_works()
    local prod_types =  (require('mil82/types'))
    local prod_type = prod_types[go:GetConfig().product_type]
    if prod_type == nil then
        error('МИЛ82: не определено исполнение ' .. product_type_name)
    end

    go:Info("исполнение", prod_type)

    params = go:ParamsDialog({
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

    local function read_lin()
        go:ReadAndSaveParam(var_concentration, float_format, 'lin1')
        local gases = { 3, 4 }
        if params.linear_degree == 4 then
            gases = { 2, 3, 4 }
        end
        for _, gas in pairs(gases) do
            go:BlowGas(gas)
            go:ReadAndSaveParam(var_concentration, float_format, 'lin' .. tostring(gas))
        end
        go:BlowGas(1)
    end

    local function temp_comp(pt_temp)
        local temperatures = {
            t_norm = params.temp_norm,
            t_low = params.temp_low,
            t_high = params.temp_high,
        }
        local temperature = temperatures[pt_temp]
        return function()
            go:Temperature(temperature)
            work_gases_read_save(pt_temp, { 1, 3, 4 })
        end
    end

    return {
        { "запись коэффициентов", function()
            go:ForEachSelectedProduct(function(p)
                mil82_product(p).init(prod_type)
            end )
        end },

        { "запуск термокамеры", function()
            go:TemperatureStop()
            go:TemperatureStart()
        end },


        { "установка НКУ", function()
            go:Temperature(params.temp_norm)
        end },

        { "продувка воздухом перед нормировкой", function()
            go:BlowGas(1)
        end },

        { "нормировка", function()
            go:Write32(8, float_format, 1000)
        end },

        { "калибровка", adjust },

        { "снятие линеаризации", read_lin },

        { "расчёт линеаризации", function()
            go:ForEachSelectedProduct(function(p)
                p:NewWork("расчёт линеаризации", function ()
                    mil82_product(p).calc_lin()
                end)
            end)
        end },

        { "запись линеаризации", function()
            go:WriteCoefficients({ 16, 17, 18, 19 }, float_format )
        end },

        {  string.format("компенсация Т-: %g⁰C", params.temp_low), temp_comp('t_low') },

        { string.format("компенсация Т+: %g⁰C", params.temp_high), temp_comp('t_high') },

        { string.format("компенсация НКУ: %g⁰C", params.temp_norm), temp_comp('t_norm') },

        { "расчёт и запись термокомпенсации", function()
            go:ForEachSelectedProduct(function (product)
                local p = mil82_product(product)
                product:NewWork("T0 начало шкалы", p.calc_T0)
                product:NewWork("TK конец шкалы", p.calc_TK)
                product:NewWork("TM середина шкалы", p.calc_TM)
            end )
        end },

        { "снятие сигналов каналов", function()
            go:ReadCoefficients({ 20, 21, 43, 44 }, float_format)
        end },

        { "НКУ: снятие для проверки погрешности", function()
            go:Temperature(params.temp_norm)
            adjust()
            work_gases_read_save('test_t_norm', { 1, 2, 3, 4 })
        end },

        { string.format("Т-: снятие для проверки погрешности: %g⁰C" , params.temp_low), function()
            go:Temperature(params.temp_low)
            work_gases_read_save('test_t_low', { 1, 3, 4 })
        end },

        { string.format("Т+: снятие для проверки погрешности: %g⁰C", params.temp_high), function()
            go:Temperature(params.temp_high)
            work_gases_read_save('test_t_high', { 1, 3, 4 })
        end },

        { string.format("90⁰C: снятие для проверки погрешности: %g⁰C", params.temp90), function()
            go:Temperature(params.temp90)
            work_gases_read_save('test_t80', { 1, 3, 4 })
        end },

        { "НКУ: повторное снятие для проверки погрешности", function()
            go:Temperature(params.temp_norm)
            work_gases_read_save('test2', { 1, 3, 4 })
        end },
    }
end

local function main()
    local cfg = go:GetConfig()
    go:Info("конфигурация", cfg)

    go:SelectWorksDialog(main_works())
end

-- перевод климатики
local function rework_temp()
    for _, var in pairs({ var_temp, var16 }) do
        for _, gas in pairs({ 1, 3, 4 }) do
            local k = '_gas' .. tostring(gas) .. '_var' .. tostring(var)
            for _, p in pairs(go:GetProducts()) do
                p:SetValue('t_norm_' .. k, p:Value('test2_' .. k))
            end
        end
    end
end

-- технологический прогон
local function technological_test()
    params = go:ParamsDialog({
        duration_tex = {
            name = "Длительность технологического прогона",
            value = '16h',
            format = 'duration',
        },
    })
    adjust()
    go:Info('снятие перед технологическим прогоном')
    work_gases_read_save('tex1', { 1, 3, 4 })
    go:Delay(params.duration_tex, 'технологический прогон')
    go:Info('снятие после технологического прогона')
    work_gases_read_save('tex2', { 1, 3, 4 })
end

-- выпуск в эксплуатацию
local function to_prod()
    local products = {}
    local xs = {}
    go:ForEachProduct(function (p)
        local i = #products+1
        products[i]=p
        xs[i] = {
            name = string.format('%d. №%d', i, p.Serial),
            value = false,
            format = 'bool',
        }
    end)

    local user_input = go:ParamsDialog(xs)

    for i,p in ipairs(products) do
        local v = user_input[i]
        if not v then
            v = 0 / 0
        end
        p:SetValue('production', v)
    end
end

local works = {
    ['Автоматическая настройка'] = main,
    ['Перевод климатики'] = rework_temp,
    ['Технологический прогон'] = technological_test,
    ['Выпуск в эксплуатацию'] = to_prod
}
local x = {
    work = {
        name = 'Выбор',
        value = 'Автоматическая настройка',
        list = {},
    },
}
for name, _ in pairs(works) do
    table.insert(x.work.list, name)
end
x = go:ParamsDialog(x)
works[x.work]()


