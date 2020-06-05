-- atool: work: МИЛ-82: выпуск в эксплуатацию

require 'utils/help'
require 'types'

local products = {}
local xs = {}
go:ForEachProduct(function (p)
    local i = #products+1
    products[i]=p
    xs[i] = {
        name = string.format('%d. №%d', i, p.Serial),
        value = false,
        format = 'bool',
    }
end)

local user_input = go:ParamsDialog(xs)

for i,p in ipairs(products) do
    local v = user_input[i]
    if ~v then
        v = 0 / 0
    end
    p:SetValue('production', v)
end