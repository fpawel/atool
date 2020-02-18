require 'mil82/def'

local product_params = {
    { name = 'Линеаризация', items = lin },
    { name = 'Термоконмпенсация начала шкалы', items = {
        { 't_low_gas1_var2', '1: Var02' },
        { 't_norm_gas1_var2', '2: Var02' },
        { 't_high_gas1_var2', '3: Var02' },
        { 't_low_gas1_var16', '1: Var16' },
        { 't_norm_gas1_var16', '2: Var16' },
        { 't_high_gas1_var16', '3: Var16' },
    } },
    { name = 'Термоконмпенсация конца шкалы', items = {
        { 't_low_gas4_var2', '1: Var02' },
        { 't_norm_gas4_var2', '2: Var02' },
        { 't_high_gas4_var2', '3: Var02' },
        { 't_low_gas4_var16', '1: Var16' },
        { 't_norm_gas4_var16', '2: Var16' },
        { 't_high_gas4_var16', '3: Var16' },
    } },
    { name = 'Термоконмпенсация середины шкалы', items = {
        { 't_low_gas3_var2', '1: Var02' },
        { 't_norm_gas3_var2', '2: Var02' },
        { 't_high_gas3_var2', '3: Var02' },
        { 't_low_gas3_var16', '1: Var16' },
        { 't_norm_gas3_var16', '2: Var16' },
        { 't_high_gas3_var16', '3: Var16' },
    } },
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
    return 'ПГС' .. tostring(gas)
end

function mil82_data(d)
    local lin = d:AddSection('Линеаризация')
    for i = 1, 4 do
        lin:AddParam('lin' .. tostring(i), 'ПГС' .. tostring(i))
    end

    for pt_key, pt_name in pairs({
        [t_low] = 'низкая температура',
        [t_norm] = 'нормальная температура',
        [t_high] = 'высокая температура',
        ['test_' .. t_norm] = 'проверка погрешности: НКУ',
        ['test_' .. t_low] = 'проверка погрешности: низкая температура',
        ['test_' .. t_high] = 'проверка погрешности: высокая температура',
        ['test2'] = 'проверка погрешности: возврат НКУ',
    }) do
        local section = d:AddSection(pt_name)
        for _, var in pairs(vars) do
            for _, gas in pairs({ 1, 3, 4 }) do
                section:AddParam(pt_key .. '_' .. mil82_db_key_gas_var(gas, var), formatVar(var) .. ': ' .. formatGas(gas))
            end
        end
    end

    for pt_key, pt_name in pairs({ tex1 = 'перед техпрогоном', tex2 = 'после техпрогона', }) do
        local section = d:AddSection(pt_name)
        for _, var in pairs(vars) do
            for _, gas in pairs({ 1, 4 }) do
                section:AddParam( pt_key .. '_' .. mil82_db_key_gas_var(gas, var), formatVar(var) .. ': ' .. formatGas(gas) )
            end
        end
    end
end



return product_params

--print_table(product_params)
