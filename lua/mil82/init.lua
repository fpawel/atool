json = require 'json'
require 'mil82/prod_types'

CO2 = 4
CH4 = 5
C3H8 = 7
C6H14 = 7

scale_code = {
    [4] = 57,
    [10] = 7,
    [20] = 9,
    [50] = 0,
    [100] = 21,
}

units_code = {
    [CO2] = 7,
    [CH4] = 14,
    [C3H8] = 14,
    [C6H14] = 14,
}

prod_type = prod_types[go.Config.product_type]

var16 = 16
varTemp = 2
varConcentration = 0

vars = { varConcentration, varTemp, 4, 8, 10, 12, 14, var16, }

function write_common_coefficients()
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

function adjust()
    go:Info("калибровка нуля")
    go:BlowGas(1)
    go:Write32(1, 'bcd', go.Config.c1)

    go:Info("калибровка чувствительности")
    go:BlowGas(4)
    go:Write32(2, 'bcd', go.Config.c4)
end

function lin_read_save(linear_degree)
    go:BlowGas(1)
    go:ReadSave(0, 'bcd', 'lin1')
    if linear_degree == 4 then
        go:BlowGas(2)
        go:ReadSave(varConcentration, 'bcd', 'lin2')
    end
    go:BlowGas(3)
    go:ReadSave(varConcentration, 'bcd', 'lin3')
    go:BlowGas(4)
    go:ReadSave(varConcentration, 'bcd', 'lin4')
    go:BlowGas(1)
end

function lin_calc(linear_degree)
    for _, p in pairs(go.Products) do
        local ct = {}
        for i = 1, 4 do
            if not (linear_degree == 3 and i == 2) then
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
        if linear_degree == 3 then
            cf[4] = 0
        end
        for i = 1, 4 do
            p:WriteKef(15 + i, 'bcd', cf[i])
        end
    end
end

local function check_value_in_array (array, val)
    for _, value in pairs(array) do
        if value == val then
            return
        end
    end
    error('value: '..tostring(val)..': must be one of: '..json.encode(array))
    return false
end

pt_temp_low = 't_low'
pt_temp_norm = 't_norm'
pt_temp_high = 't_high'

local function check_pt_temp (pt_temp)
    check_value_in_array ({pt_temp_low, pt_temp_norm, pt_temp_high}, pt_temp)
end

function temp_db_key(pt_temp, gas, var)
    check_pt_temp(pt_temp)
    return pt_temp .. '_gas' .. tostring(gas) .. '_var' .. tostring(var)
end

function temp_read_save(temperature, pt_temp, middle_scale)
    check_pt_temp(pt_temp)
    go:Info("перевод термокамеры на " .. tostring(temperature) .. "⁰C")
    go:Temperature(temperature)
    go:Info("снятие " .. pt_temp .. ": " .. tostring(temperature) .. "⁰C")
    local gases = {}
    if middle_scale then
        gases = { 1, 3, 4 }
    else
        gases = { 1, 4 }
    end
    for _, gas in ipairs(gases) do
        go:BlowGas(gas)
        for _, var in pairs(vars) do
            go:ReadSave(var, 'bcd', temp_db_key(pt_temp, gas, var))
        end
    end
    go:BlowGas(1)
end

function temp_calc(middle_scale)
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