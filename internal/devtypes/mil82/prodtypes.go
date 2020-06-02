package mil82

type prodT struct {
	Gas   string  `json:"gas"`
	Scale float64 `json:"scale"`
	Index int     `json:"index"`
}

var prodTypes = map[string]prodT{
	"00.00": {
		Gas:   "CO2",
		Scale: 4,
		Index: 1,
	},
	"00.01": {
		Gas:   "CO2",
		Scale: 10,
		Index: 2,
	},
	"00.02": {
		Gas:   "CO2",
		Scale: 20,
		Index: 3,
	},
	"01.00": {
		Gas:   "CH4",
		Scale: 100,
		Index: 4,
	},
	"01.01": {
		Gas:   "CH4",
		Scale: 100,
		Index: 5,
	},
	"02.00": {
		Gas:   "C3H8",
		Scale: 50,
		Index: 6,
	},
	"02.01": {
		Gas:   "C3H8",
		Scale: 50,
		Index: 7,
	},
	"03.00": {
		Gas:   "C3H8",
		Scale: 100,
		Index: 8,
	},
	"03.01": {
		Gas:   "C3H8",
		Scale: 100,
		Index: 9,
	},
	"04.00": {
		Gas:   "CH4",
		Scale: 100,
		Index: 10,
	},
	"05.00": {
		Gas:   "C6H14",
		Scale: 50,
		Index: 11,
	},
	"10.00": {
		Gas:   "CO2",
		Scale: 4,
		Index: 12,
	},
	"10.01": {
		Gas:   "CO2",
		Scale: 10,
		Index: 13,
	},
	"10.02": {
		Gas:   "CO2",
		Scale: 20,
		Index: 14,
	},
	"10.03": {
		Gas:   "CO2",
		Scale: 4,
		Index: 15,
	},
	"10.04": {
		Gas:   "CO2",
		Scale: 10,
		Index: 16,
	},
	"10.05": {
		Gas:   "CO2",
		Scale: 20,
		Index: 17,
	},
	"11.00": {
		Gas:   "CH4",
		Scale: 100,
		Index: 18,
	},
	"11.01": {
		Gas:   "CH4",
		Scale: 100,
		Index: 19,
	},
	"13.00": {
		Gas:   "C3H8",
		Scale: 100,
		Index: 20,
	},
	"13.01": {
		Gas:   "C3H8",
		Scale: 100,
		Index: 21,
	},
	"14.00": {
		Gas:   "CH4",
		Scale: 100,
		Index: 22,
	},
	"16.00": {
		Gas:   "C3H8",
		Scale: 100,
		Index: 23,
	},
}
