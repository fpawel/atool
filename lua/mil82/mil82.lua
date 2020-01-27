-- atoolgui: МИЛ82: Автоматическая настройка

require 'mil82/init'
require 'print_table'

print_table(common_coefficients_values)

params = go:ParamsDialog({
    float_format = {
        name = "Формат вещественного числа",
        value = 'bcd',
        list = { 'bcd', 'float_big_endian', 'float_little_endian' },
    },
})

for _, p in pairs(go.Products) do
    for k, v in pairs (common_coefficients_values) do
        p:WriteKef(k, params.float_format, v)
    end
end