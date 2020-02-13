-- atool: work: Газовый блок: клапан

go:SwitchGas(go:ParamsDialog({
    x = {
        name = "Клапан",
        value = 0,
        type = 'int',
    },
}).x, false)