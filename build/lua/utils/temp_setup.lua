require 'utils/help'

local current_temperature

function setupTemperature(temperature)
    local what = 'перевод термокамеры: '
    if current_temperature == nil then
        what = what .. format_temperature(temperature)
    else
        what = what .. format_temperature(current_temperature) .. ' -> ' .. format_temperature(temperature)
    end
    if current_temperature == temperature then
        go:Info(what .. ': температура уже установлена')
        return
    end
    go:NewWork(what, function()
        go:SwitchGas(0, true)
        go:Temperature(temperature)
        current_temperature = temperature
    end)
end