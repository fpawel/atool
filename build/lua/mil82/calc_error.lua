require 'mil82/def'
require 'utils/help'

local sections = {
    ['test_' .. t_norm] = { 'нормальная температура' },
    ['test_' .. t_low] = { 'низкая температура', 20 },
    ['test_' .. t_high] = { 'высокая температура', 20 },
    ['test2'] = { 'возврат НКУ' },
    ['test_t80'] = { '80⁰C', 80 },
    ['tex1'] = { 'перед технологическим прогоном' },
    ['tex2'] = { 'после технологического прогона' },
}

local function keyGasVar(gas, var)
    return ikds4_db_key_gas_var(gas, var)
end

local function calc_error(party, calc, prod_type)
    for pt_key, pt in pairs(sections) do
        local section = calc:AddCalcSection(pt[1])
        for _, gas in pairs({ 1, 3, 4 }) do
            local param = section:AddParam('газ ' .. tostring(gas))
            local valueKey = pt_key .. '_' .. keyGasVar(gas, varConcentration)
            for i = 1, #party.Products do
                local product = party.Products[i]
                local function val(k)
                    return value_or_nan(product.Values[k])
                end

                local nominal = value_or_nan(party.Values['c' .. tostring(pt.gas)])

                local absErrLimit = 0 / 0

                if prod_type.gas ~= CO2 then
                    absErrLimit = 2.5 + 0.05 * nominal
                elseif prod_type.scale == 4 then
                    absErrLimit = 0.2 + 0.05 * nominal
                elseif prod_type.scale == 10 then
                    absErrLimit = 0.5
                elseif prod_type.scale == 20 then
                    absErrLimit = 1
                end

                local tNorm = pt[2]
                if tNorm ~= nil then
                    local var2 = 0 / 0
                    if tNorm == 80 then
                        nominal = val('test_t80_' .. keyGasVar(gas, varConcentration))
                        var2 = val('test_t80_' .. keyGasVar(gas, varTemp))
                    elseif tNorm == 20 then
                        nominal = val('test_t_norm_' .. keyGasVar(gas, varConcentration))
                        var2 = val('test_t_norm_' .. keyGasVar(gas, varTemp))
                    end
                    if prod_type.gas == CO2 then
                        absErrLimit = 0.5 * math.abs(absErrLimit * (var2 - tNorm)) / 10
                    elseif pt.gas == 1 then
                        absErrLimit = 5
                    else
                        absErrLimit = math.abs(nominal * 0.15)
                    end
                end
                local value = val(valueKey)
                local absErr = value - nominal
                local relErr = 100 * absErr / absErrLimit

                local v = param:AddValue()
                v.Detail = stringify({
                    ['газ'] = gas,
                    ['концентрация'] = value,
                    ['номинал'] = nominal,
                    ['погрешность'] = absErr,
                    ['предел погрешности'] = absErrLimit,
                    ['db_key'] = valueKey,
                    ['t_norm'] = tNorm,
                    ['product_type'] = party.ProductType,
                    ['gas'] = prod_type.gas,
                    ['ПГС'] = party['c' .. tostring(pt.gas)],
                })
                if relErr == relErr then
                    v.Validated = true
                    v.Valid = math.abs(absErr) < math.abs(absErrLimit)
                    v.Value = string.format("%.1f", relErr)
                end
            end
        end
    end
end

function mil82_calc(party, calc)
    local prod_type = prod_types[party.ProductType]
    if prod_type == nil then
        return 'не правильное исполнение: ' .. product_type_name
    end
    calc_error(party, calc, prod_type)
    return ""
end