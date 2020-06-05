-- atool: work: МИЛ-82: перевод климатики

local var16 = 16
local varTemp = 2

for _,var in pairs({varTemp, var16}) do
    for _,gas in pairs({1, 3, 4}) do
        local k = '_gas'..tostring(gas) .. '_var' .. tostring(var)
        for _, p in pairs(go:GetProducts()) do
            p:SetValue('t_norm_'..k, p:Value('test2_'..k))
        end
    end
end

