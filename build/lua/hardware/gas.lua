-- atool: work: Газовый блок: клапан

go:SwitchGasNoWarn(go:ParamsDialog({
    x = {
        name = "Клапан",
        value = 0,
        type = 'int',
    },
}).x, false)