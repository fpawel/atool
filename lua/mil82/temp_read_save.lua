-- atoolgui: МИЛ82: Термокомпенсация: снятие

require 'mil82/init'

local params = go:ParamsDialog({
    middle_scale = {
        name = "середина шкалы",
        value = true,
        format = 'bool',
    },
    minus = {
        name = "Минус",
        value = true,
        format = 'bool',
    },
    plus = {
        name = "Плюс",
        value = true,
        format = 'bool',
    },
    nku = {
        name = "НКУ",
        value = true,
        format = 'bool',
    },
    temp_norm = {
        name = "Уставка температуры НКУ,⁰C",
        value = 20,
        format = 'float',
    },
    temp_low = {
        name = "Уставка низкой температуры,⁰C",
        value = prod_type.temp_low,
        format = 'float',
    },
    temp_high = {
        name = "Уставка высокой температуры,⁰C",
        value = prod_type.temp_high,
        format = 'float',
    },
})

if params.minus then
    temp_read_save(params.temp_low,pt_temp_low, params.middle_scale)
end
if params.plus then
    temp_read_save(params.temp_high, pt_temp_high, params.middle_scale)
end
if params.nku then
    temp_read_save(params.temp_norm, pt_temp_norm, params.middle_scale)
end

