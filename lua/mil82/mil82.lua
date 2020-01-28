-- atoolgui: МИЛ82: Автоматическая настройка
require 'print_table'
require 'mil82/init'
-- print_table(common_coefficients_values)

params = go:ParamsDialog({
    linear_degree = {
        name = "Степень линеаризации",
        value = 4,
        format = 'int',
        list = { '3', '4' },
    },
    temp_middle_scale = {
        name = "Термокомпенсация середины шкалы",
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
        value = prod.temp_low,
        format = 'float',
    },
    temp_high = {
        name = "Уставка высокой температуры,⁰C",
        value = prod.temp_high,
        format = 'float',
    },
})
print_table(params)

go:Info("запись коэффициентов")
for k, v in pairs (common_coefficients_values) do
    go:WriteKef(k, 'bcd', v)
end

go:Info("установка НКУ")
go:SwitchGas(0)
go:Temperature(params.temp_norm)

go:Info("нормировка")
go:Write32(8, 'bcd', 1000)

go:Info("калибровка нуля")
go:BlowGas(1)
go:Write32(1, 'bcd', go.Config.c1)

go:Info("калибровка чувствительности")
go:BlowGas(4)
go:Write32(2, 'bcd', go.Config.c4)

go:Info("снятие линиаризации")
lin_read_save(params.linear_degree)

go:Info("расчёт линиаризации")
lin_calc(params.linear_degree)