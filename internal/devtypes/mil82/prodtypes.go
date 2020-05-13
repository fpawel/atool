package mil82

type prodT struct {
	gas   string
	scale float64
	index int
}

var prodTypes = map[string]prodT{
	"00.00": {
		gas:   "CO2",
		scale: 4,
		index: 1,
	},
	"00.01": {
		gas:   "CO2",
		scale: 10,
		index: 2,
	},
	"00.02": {
		gas:   "CO2",
		scale: 20,
		index: 3,
	},
	"01.00": {
		gas:   "CH4",
		scale: 100,
		index: 4,
	},
	"01.01": {
		gas:   "CH4",
		scale: 100,
		index: 5,
	},
	"02.00": {
		gas:   "C3H8",
		scale: 50,
		index: 6,
	},
	"02.01": {
		gas:   "C3H8",
		scale: 50,
		index: 7,
	},
	"03.00": {
		gas:   "C3H8",
		scale: 100,
		index: 8,
	},
	"03.01": {
		gas:   "C3H8",
		scale: 100,
		index: 9,
	},
	"04.00": {
		gas:   "CH4",
		scale: 100,
		index: 10,
	},
	"05.00": {
		gas:   "C6H14",
		scale: 50,
		index: 11,
	},
	"10.00": {
		gas:   "CO2",
		scale: 4,
		index: 12,
	},
	"10.01": {
		gas:   "CO2",
		scale: 10,
		index: 13,
	},
	"10.02": {
		gas:   "CO2",
		scale: 20,
		index: 14,
	},
	"10.03": {
		gas:   "CO2",
		scale: 4,
		index: 15,
	},
	"10.04": {
		gas:   "CO2",
		scale: 10,
		index: 16,
	},
	"10.05": {
		gas:   "CO2",
		scale: 20,
		index: 17,
	},
	"11.00": {
		gas:   "CH4",
		scale: 100,
		index: 18,
	},
	"11.01": {
		gas:   "CH4",
		scale: 100,
		index: 19,
	},
	"13.00": {
		gas:   "C3H8",
		scale: 100,
		index: 20,
	},
	"13.01": {
		gas:   "C3H8",
		scale: 100,
		index: 21,
	},
	"14.00": {
		gas:   "CH4",
		scale: 100,
		index: 22,
	},
	"16.00": {
		gas:   "C3H8",
		scale: 100,
		index: 23,
	},
}
