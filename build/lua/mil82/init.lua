require 'mil82/prod_types'
json = require 'json'

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

pt_temp_low = 't_low'
pt_temp_norm = 't_norm'
pt_temp_high = 't_high'

function db_key_gas_var(gas, var)
    return 'gas'..tostring(gas) .. '_var' .. tostring(var)
end

