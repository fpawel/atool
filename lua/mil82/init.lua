require 'print_table'

local json = require ("dkjson")

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

prod = ({
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
})[go.Config.product_type]

common_coefficients_values = (function()
    return {
        [2] = os.date("*t").year,
        [10] = go.Config.c1,
        [11] = go.Config.c3,
        [7] = scale_code[prod.scale],
        [8] = prod.scale_begin or 0,
        [9] = prod.scale,
        [5] = units_code[prod.gas],
        [6] = prod.gas,

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
    }
end)()

function lin_read_save(linear_degree)
    go:Info("снятие линиаризации")
    go:BlowGas(1)
    go:ReadSave(0, 'bcd', 'lin1')
    if linear_degree == 4 then
        go:BlowGas(2)
        go:ReadSave(0, 'bcd', 'lin2')
    end
    go:BlowGas(3)
    go:ReadSave(0, 'bcd', 'lin3')
    go:BlowGas(4)
    go:ReadSave(0, 'bcd', 'lin4')
end

function lin_calc(linear_degree)

    for _, p in pairs (go.Products) do
        local ct = {}
        for i = 1,4 do
            if not (linear_degree == 3 and i == 2) then
                local x = p:Value('lin' .. tostring(i))
                if x == nil then
                    p:Err('расёт линеаризатора не выполнен: нет значения lin'..tostring(i))
                    return
                end
                ct[i] = {x, go.Config['c' .. tostring(i)]}
            end
        end

        p:Info('расчёт линеаризатора:'..json.encode(ct, { indent = true }))

        local cf = go:InterpolationCoefficients(ct)
        if cf == nil then
            p:Err('расёт линеаризатора не выполнен')
            return
        end
        if linear_degree == 3 then
            cf[4] = 0
        end
        p:Info('расчёт линеаризатора:'..json.encode(cf, { indent = true }))
        for i = 1,4 do
            p:WriteKef(15 + i, 'bcd', cf[i])
        end
    end
end