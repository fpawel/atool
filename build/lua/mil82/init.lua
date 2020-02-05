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

var16 = 16
varTemp = 2
varConcentration = 0

vars = { varConcentration, varTemp, 4, 8, 10, 12, 14, var16, }

t_low = 't_low'
t_norm = 't_norm'
t_high = 't_high'

function db_key_gas_var(gas, var)
    return 'gas'..tostring(gas) .. '_var' .. tostring(var)
end
