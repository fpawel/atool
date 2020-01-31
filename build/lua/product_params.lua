require 'mil82/init'
local lin = {}

for i = 1,4 do
    table.insert(lin, { 'lin'..tostring(i), 'ПГС'..tostring(i)} )
end

local vars_names = {
    [varConcentration] = 'концентрация',
    [varTemp] = 'сигнал датчика температуры',
    [var16] = 'измерительный канал',
    [4] = 'ток излучателя',
    [12] = 'рабочий канал',
    [14] = 'опорный канал',
}

local function formatVar(var)
    local x = vars_names[var]
    if x == nil then
        return 'регистр '..tostring(var)
    end
    return x
end

local temp = {}

local pt_temps = {
    [pt_temp_low] = 'низкая температура',
    [pt_temp_norm] = 'нормальная температура',
    [pt_temp_high] = 'высокая температура',
}

product_params = { { 'Линеаризация', lin} }

local function formatGas(gas)
    return 'ПГС'..tostring(gas)
end

for pt_temp, pt_temp_name in pairs(pt_temps) do
    for _,var in pairs(vars) do
        local x = {}
        for _,gas in pairs({1,3,4}) do
            table.insert(x, { pt_temp ..'_'.. db_key_gas_var( gas, var), formatGas(gas)})
        end
        table.insert(product_params, { pt_temp_name..': '..formatVar(var), x})
    end
end
