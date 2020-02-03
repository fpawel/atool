-- atool-data: МИЛ82: снятие

require 'mil82/init'
--require 'print_table'
local lin = {}

for i = 1,4 do
    table.insert(lin, { 'lin'..tostring(i), 'ПГС'..tostring(i)} )
end

local product_params = {
    { 'Линеаризация', lin},
    { 'Термоконмпенсация начала шкалы', {
        {'t_low_gas1_var2',  '1: Var02'},
        {'t_norm_gas1_var2', '2: Var02'},
        {'t_high_gas1_var2', '3: Var02'},
        {'t_low_gas1_var16', '1: Var16'},
        {'t_norm_gas1_var16','2: Var16'},
        {'t_high_gas1_var16','3: Var16'},
    }},
    { 'Термоконмпенсация конца шкалы', {
        {'t_low_gas4_var2',  '1: Var02'},
        {'t_norm_gas4_var2', '2: Var02'},
        {'t_high_gas4_var2', '3: Var02'},
        {'t_low_gas4_var16', '1: Var16'},
        {'t_norm_gas4_var16','2: Var16'},
        {'t_high_gas4_var16','3: Var16'},
    }},
    { 'Термоконмпенсация середины шкалы', {
        {'t_low_gas3_var2',  '1: Var02'},
        {'t_norm_gas3_var2', '2: Var02'},
        {'t_high_gas3_var2', '3: Var02'},
        {'t_low_gas3_var16', '1: Var16'},
        {'t_norm_gas3_var16','2: Var16'},
        {'t_high_gas3_var16','3: Var16'},
    }},
}

local vars_names = {
    [varConcentration] = 'C',
    [varTemp] = 'T',
    [var16] = 'Var16',
    [4] = 'I',
    [12] = 'Work',
    [14] = 'Ref',
}

local function formatVar(var)
    local x = vars_names[var]
    if x == nil then
        return tostring(var)
    end
    return x
end

local function formatGas(gas)
    return 'ПГС'..tostring(gas)
end

for pt_key, pt_name in pairs({
    [pt_temp_low] = 'низкая температура',
    [pt_temp_norm] = 'нормальная температура',
    [pt_temp_high] = 'высокая температура',
}) do
    local x = {}
    for _,var in pairs(vars) do
        for _,gas in pairs({1,3,4}) do
            table.insert(x, { pt_key ..'_'.. db_key_gas_var( gas, var), formatVar(var)..': '..formatGas(gas)})
        end
    end
    table.insert(product_params, { pt_name, x})
end

for pt_key, pt_name in pairs({ tex1 = 'перед техпрогоном', tex2 = 'после техпрогона',}) do
    local x = {}
    for _,var in pairs(vars) do
        for _,gas in pairs({1,4}) do
            table.insert(x, { pt_key ..'_'.. db_key_gas_var( gas, var), formatVar(var)..': '..formatGas(gas)})
        end
    end
    table.insert(product_params, { pt_name, x})
end

return product_params

--print_table(product_params)
