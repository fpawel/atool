-- atoolgui: МИЛ82: Автоматическая настройка

require 'mil82/init'

go:Info("конфигурация: "..json.encode(go.Config))

local params = go:ParamsDialog({
    linear_degree = {
        name = "Степень линеаризации",
        value = 4,
        format = 'int',
        list = { '3', '4' },
    },
    temp_middle_scale = {
        name = "Термокомпенсация: середина шкалы",
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

--go:Info("параметры: "..json.encode(params))
--
--go:Info("запись коэффициентов")
--write_common_coefficients()
--
--go:Info("установка НКУ")
--go:SwitchGas(0,true)
--go:Temperature(params.temp_norm)
--
--go:Info("нормировка")
--go:Write32(8, 'bcd', 1000)
--
--go:Info("калибровка нуля")
--go:BlowGas(1)
--go:Write32(1, 'bcd', go.Config.c1)
--
--go:Info("калибровка чувствительности")
--go:BlowGas(4)
--go:Write32(2, 'bcd', go.Config.c4)
--
--go:Info("снятие линиаризации")
--lin_read_save(params.linear_degree)
--
--go:Info("расчёт линиаризации")
--lin_calc(params.linear_degree)

temp_read_save(params.temp_low, pt_temp_low, params.temp_middle_scale)
temp_read_save(params.temp_high, pt_temp_high, params.temp_middle_scale)
temp_read_save(params.temp_norm, pt_temp_norm, params.temp_middle_scale)

