function value_or_nan(v)
    if v~=nil and v == v then
        return v
    end
    return 0 / 0
end

function array_n(xs, n)
    local ret = {}
    for i, v in pairs(xs) do
        ret[n + i - 1] = v
    end
    return ret
end

