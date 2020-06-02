function value_or_nan(v)
    if v~=nil and v == v then
        return v
    end
    return 0 / 0
end

function format_temperature(temperature)
    return tostring(temperature) .. '⁰C'
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

function write_coefficients_product (float_format, values, product)
    for k, value in pairs(values) do
        product:WriteKef(k, float_format, value)
    end
end

function read_coefficients_product (float_format, coefficients, product)
    for k, _ in pairs(coefficients) do
        product:ReadKef(k, float_format)
    end
end

function write_coefficients(float_format, coefficients)
    table.sort(coefficients, function(a, b)
        return a < b
    end)

    go:NewWork('запись коэффициентов ' .. go:Stringify(coefficients), function()
        for _, product in pairs(Products) do
            for _, k in pairs(coefficients) do
                product:WriteKef(k, float_format, product:Kef(k))
            end
        end
    end)
end

