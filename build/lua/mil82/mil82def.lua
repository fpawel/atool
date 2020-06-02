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
        index = 1
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
        index = 2
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
        index = 3
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
        index = 4
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
        index = 5
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
        index = 6
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
        index = 7
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
        index = 8
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
        index = 9
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
        index = 10
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
        index = 11
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
        index = 12
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
        index = 13
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
        index = 14
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
        index = 15
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
        index = 16
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
        index = 17
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
        index = 18
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
        index = 19
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
        index = 20
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
        index = 21
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
        index = 22
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
        index = 23
    },
}

