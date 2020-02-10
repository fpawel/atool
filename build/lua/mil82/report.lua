-- atool: report: МИЛ-82: погрешность
require 'mil82/init'

local Products = go:GetProducts()

report:AddHeader('МИЛ-82')

for _, p in pairs(Products) do
    report:AddHeader(tostring(p.Serial))
end

local tab = report:AddTable("")

local row = tab:AddRow("ID")
for _, p in pairs(Products) do
    row:AddCell(tostring(p.ID))
end

row = tab:AddRow("Адрес")
for _, p in pairs(Products) do
    row:AddCell(tostring(p.Addr))
end

local config = go:GetConfig()
local prod_type = prod_types[config.product_type]

local function concentrationErrorLimit20(concentration)
    if prod_type.gas ~= CO2 then
        return 2.5 + 0.05 * concentration
    end
    if prod_type.scale == 4 then
        return 0.2 + 0.05 * concentration
    elseif prod_type.scale == 10 then
        return 0.5
    elseif prod_type.scale == 20 then
        return 1
    else
        return 0 / 0
    end
end

for pt_key, pt_info in pairs({
    ['test_'..t_norm] = {'нормальная температура (НКУ)'},
    ['test_'..t_low] = {'низкая температура', 20},
    ['test_'..t_high] = {'высокая температура', 20},
    ['test2'] = {'возврат НКУ'},
}) do

    tab = report:AddTable(pt_info[1])

    for _,gas in pairs({1,3,4}) do
        row = tab:AddRow("ПГС"..tostring(gas))

        for _, p in pairs(Products) do
            local value = p:Value(pt_key..'_'..db_key_gas_var( gas, varConcentration))
            local nominal = config['c'..tostring(gas)]
            local absErr = value - nominal

            local absErrLimit = concentrationErrorLimit20(value)
            if pt_info[2] ~= nil then
                local t = p:Value(pt_key..'_'..db_key_gas_var( gas, varTemp))
                local t_norm = pt_info[2]
                absErrLimit = 0.5 * math.abs( absErrLimit * (t - t_norm) ) / 10
            end

            local absErrOk = math.abs(absErr) < math.abs(absErrLimit)
            local relErrPercent =
                100 * math.abs(absErr) / absErrLimit

            if relErrPercent == relErrPercent then
                local s = string.format("%.1f", relErrPercent)
                if absErrOk then
                    row:AddCellOk(s)
                else
                    row:AddCellErr(s)
                end
            else
                row:AddCell("")
            end

        end
    end
end







