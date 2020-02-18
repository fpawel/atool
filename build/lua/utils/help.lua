json = require("dkjson")

function stringify(v)
    return json.encode(v, { indent = true })
end

function format_temperature(temperature)
    return tostring(temperature) .. '⁰C'
end

function format_product_number(p)
    return string.format('№%d.id%d', p.Serial, p.ID)
end

function array_n(xs, n)
    local ret = {}
    for i, v in pairs(xs) do
        ret[n + i - 1] = v
    end
    return ret
end

function set_coefficients_product (values, product)
    for k, value in pairs(values) do
        product:SetKef(k, value)
    end
end

function write_coefficients_product (values, product)
    for k, value in pairs(values) do
        product:WriteKef(k, 'bcd', value)
    end
end

function write_coefficients(coefficients)

    table.sort(coefficients, function(a, b)
        return a < b
    end)

    go:NewWork('запись коэффициентов ' .. stringify(coefficients), function()
        for _, product in pairs(Products) do
            for _, k in pairs(coefficients) do
                product:WriteKef(k, 'bcd', product:Kef(k))
            end
        end
    end)
end

