-- atoolgui: МИЛ82: Линиаризация: снятие
require 'mil82/init'

lin_read_save(go:ParamsDialog({
    linear_degree = {
        name = "Степень линеаризации",
        value = 4,
        format = 'int',
        list = { '3', '4' },
    },
}).linear_degree)