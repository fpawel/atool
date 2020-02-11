const productTypes = {
    ["00.00"] : {
        gas : "CO2",
        scale : 4,
    },
    ["00.01"] : {
        gas : "CO2",
        scale : 10,
    },
    ["00.02"] : {
        gas : "CO2",
        scale : 20,
    },
    ["01.00"] : {
        gas : "CH4",
        scale : 100,
    },
    ["01.01"] : {
        gas : "CH4",
        scale : 100,
    },
    ["02.00"] : {
        gas : "C3H8",
        scale : 50,
    },
    ["02.01"] : {
        gas : "C3H8",
        scale : 50,
    },
    ["03.00"] : {
        gas : "C3H8",
        scale : 100,
    },
    ["03.01"] : {
        gas : "C3H8",
        scale : 100,
    },
    ["04.00"] : {
        gas : "CH4",
        scale : 100,
    },
    ["05.00"] : {
        gas : "C6H14",
        scale : 50,
    },
    ["10.00"] : {
        gas : "CO2",
        scale : 4,
    },
    ["10.01"] : {
        gas : "CO2",
        scale : 10,
    },
    ["10.02"] : {
        gas : "CO2",
        scale : 20,
    },
    ["10.03"] : {
        gas : "CO2",
        scale : 4,
    },
    ["10.04"] : {
        gas : "CO2",
        scale : 10,
    },
    ["10.05"] : {
        gas : "CO2",
        scale : 20,
    },
    ["11.00"] : {
        gas : "CH4",
        scale : 100,
    },
    ["11.01"] : {
        gas : "CH4",
        scale : 100,
    },
    ["13.00"] : {
        gas : "C3H8",
        scale : 100,
    },
    ["13.01"] : {
        gas : "C3H8",
        scale : 100,
    },
    ["14.00"] : {
        gas : "CH4",
        scale : 100,
    },
    ["16.00"] : {
        gas : "C3H8",
        scale : 100,
    },
};


function concentrationErrorLimit20(concentration, prodType) {
    if (prodType.gas !== "CO2"){
        return 2.5 + 0.05 * concentration;
    }
    switch (prodType.scale) {
        case 4:
            return 0.2 + 0.05 * concentration;
        case 10:
            return 0.5;
        case 20:
            return 1;
        default:
            return NaN;
    }
}