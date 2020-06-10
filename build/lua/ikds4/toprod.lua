-- atool: work: ИКД-С4: выпуск в эксплуатацию

require 'utils/help'
require 'ikds4/ikds4def'

local Products = go:GetProducts()

local function encode2(a,b)
    return a * 10000 + b
end

local t = os.date("*t")
local prod_type_index = prod_types[go:GetConfig().product_type].index

print("year:", t.year, "month:", t.month)

for _, p in pairs(Products) do
    local coefficients = {
        [40] = encode2(t.year-2000, p.Serial),
        [47] = encode2(t.month, prod_type_index)
    }
    set_coefficients_product(coefficients, p)
    read_coefficients_product('bcd', {20,21,43,44}, p)

end
