-- atool: work: МИЛ-82: перевод климатики

require 'help'
require 'temp_setup'
require 'mil82/def'

local Products = go:GetProducts()

for _,var in pairs({varTemp, var16}) do
    for _,gas in pairs({1, 3, 4}) do
        local k = mil82_db_key_gas_var(gas, var)
        for _, p in pairs(Products) do
            p:SetValue('t_norm_'..k, p:Value('test2_'..k))
        end
    end
end

