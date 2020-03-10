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

function mil82_db_key_gas_var(gas, var)
    return 'gas'..tostring(gas) .. '_var' .. tostring(var)
end

prod_types = {
    ['00.00'] = {
        gas = CO2,
        scale = 4,
        temp_low = -40,
        temp_high = 80,
        coefficient = {
            [4] = 5,
            [14] = 0.1,
            [35] = 5,
            [45] = 60,
            [50] = 0,
        },
    },
    ['00.01'] = {
        gas = CO2,
        scale = 10,
        temp_low = -40,
        temp_high = 80,
        coefficient = {
            [4] = 5,
            [14] = 0.1,
            [35] = 5,
            [45] = 60,
            [50] = 0,
        },
    },
    ['00.02'] = {
        gas = CO2,
        scale = 20,
        temp_low = -40,
        temp_high = 80,
        coefficient = {
            [4] = 5,
            [14] = 0.1,
            [35] = 5,
            [45] = 60,
            [50] = 0,
        },
    },
    ['01.00'] = {
        gas = CH4,
        scale = 100,
        temp_low = -40,
        temp_high = 80,
        coefficient = {
            [4] = 7.5,
            [14] = 0.5,
            [35] = 5,
            [45] = 60,
            [50] = 0,
        },
    },
    ['01.01'] = {
        gas = CH4,
        scale = 100,
        temp_low = -60,
        temp_high = 60,
        coefficient = {
            [4] = 7.5,
            [14] = 0.5,
            [35] = 5,
            [45] = 60,
            [50] = 0,
        },
    },
    ['02.00'] = {
        gas = C3H8,
        scale = 50,
        temp_low = -40,
        temp_high = 60,
        coefficient = {
            [4] = 12.5,
            [14] = 0.5,
            [35] = 5,
            [45] = 30,
            [50] = 0,
        },
    },
    ['02.01'] = {
        gas = C3H8,
        scale = 50,
        temp_low = -60,
        temp_high = 60,
        coefficient = {
            [4] = 12.5,
            [14] = 0.5,
            [35] = 5,
            [45] = 30,
            [50] = 0,
        },
    },
    ['03.00'] = {
        gas = C3H8,
        scale = 100,
        temp_low = -40,
        temp_high = 60,
        coefficient = {
            [4] = 12.5,
            [14] = 0.5,
            [35] = 5,
            [45] = 30,
            [50] = 0,
        },
    },
    ['03.01'] = {
        gas = C3H8,
        scale = 100,
        temp_low = -60,
        temp_high = 60,
        coefficient = {
            [4] = 12.5,
            [14] = 0.5,
            [35] = 5,
            [45] = 30,
            [50] = 0,
        },
    },
    ['04.00'] = {
        gas = CH4,
        scale = 100,
        temp_low = -60,
        temp_high = 80,
        coefficient = {
            [4] = 7.5,
            [14] = 0.5,
            [35] = 5,
            [45] = 60,
            [50] = 0,
        },
    },
    ['05.00'] = {
        gas = C6H14,
        scale = 50,
        temp_low = 15,
        temp_high = 80,
        coefficient = {
            [4] = 1,
            [14] = 30,
            [35] = 5,
            [45] = 30,
            [50] = 0,
        },
    },
    ['10.00'] = {
        gas = CO2,
        scale = 4,
        temp_low = -40,
        temp_high = 80,
        coefficient = {
            [4] = 1,
            [14] = 30,
            [35] = 1,
            [45] = 30,
            [50] = 1,
        },
    },
    ['10.01'] = {
        gas = CO2,
        scale = 10,
        temp_low = -40,
        temp_high = 80,
        coefficient = {
            [4] = 1,
            [14] = 30,
            [35] = 1,
            [45] = 30,
            [50] = 1,
        },
    },
    ['10.02'] = {
        gas = CO2,
        scale = 20,
        temp_low = -40,
        temp_high = 80,
        coefficient = {
            [4] = 1,
            [14] = 30,
            [35] = 1,
            [45] = 30,
            [50] = 1,
        },
    },
    ['10.03'] = {
        gas = CO2,
        scale = 4,
        temp_low = -60,
        temp_high = 80,
        coefficient = {
            [4] = 1,
            [14] = 30,
            [35] = 1,
            [45] = 30,
            [50] = 1,
        },
    },
    ['10.04'] = {
        gas = CO2,
        scale = 10,
        temp_low = -60,
        temp_high = 80,
        coefficient = {
            [4] = 1,
            [14] = 30,
            [35] = 1,
            [45] = 30,
            [50] = 1,
        },
    },
    ['10.05'] = {
        gas = CO2,
        scale = 20,
        temp_low = -60,
        temp_high = 80,
        coefficient = {
            [4] = 1,
            [14] = 30,
            [35] = 1,
            [45] = 30,
            [50] = 1,
        },
    },
    ['11.00'] = {
        gas = CH4,
        scale = 100,
        temp_low = -40,
        temp_high = 80,
        coefficient = {
            [4] = 1,
            [14] = 30,
            [35] = 1,
            [45] = 30,
            [50] = 1,
        },
    },
    ['11.01'] = {
        gas = CH4,
        scale = 100,
        temp_low = -60,
        temp_high = 80,
        coefficient = {
            [4] = 1,
            [14] = 30,
            [35] = 1,
            [45] = 30,
            [50] = 1,
        },
    },
    ['13.00'] = {
        gas = C3H8,
        scale = 100,
        temp_low = -40,
        temp_high = 80,
        coefficient = {
            [4] = 1,
            [14] = 30,
            [35] = 1,
            [45] = 30,
            [50] = 1,
        },
    },
    ['13.01'] = {
        gas = C3H8,
        scale = 100,
        temp_low = -60,
        temp_high = 80,
        coefficient = {
            [4] = 1,
            [14] = 30,
            [35] = 1,
            [45] = 30,
            [50] = 1,
        },
    },
    ['14.00'] = {
        gas = CH4,
        scale = 100,
        temp_low = -60,
        temp_high = 80,
        coefficient = {
            [4] = 1,
            [14] = 30,
            [35] = 1,
            [45] = 30,
            [50] = 1,
        },
    },
    ['16.00'] = {
        gas = C3H8,
        scale = 100,
        temp_low = -60,
        temp_high = 80,
        coefficient = {
            [4] = 1,
            [14] = 30,
            [35] = 1,
            [45] = 30,
            [50] = 1,
        },
    },
}

