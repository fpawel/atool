-- atoolgui: Оборудование: Термокамера: уставка

go:TemperatureSetup(go:ParamsDialog({
    temperature = {
        name = "Температура уставки",
        value = 20,
        type = 'float',
    },
}).temperature)