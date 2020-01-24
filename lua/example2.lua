require '/print_table'

params = {
    FBool = {
        name = "the bool",
        value = true,
    },
    FFloat = {
        name = "the float",
        value = 12.33,
    },
    FInt = {
        name = "the int",
        value = 12,
        type = 'int',
    },
    FStr = {
        name = "a str",
        value = "str",
    },
    FStrList = {
        name = "a str list",
        value = "some",
        list = {
            "00.12", "00.13", "00.14",
        }
    },
}

print('before')
print_table(params)

params = go:ParamsDialog(params)

print('after')
print('theBool', params.FBool)
print('theFloat', params.FFloat)
print('theInt', params.FInt)
print('theBool', params.FBool)
print('theStrList', params.FStrList)

print_table(params)