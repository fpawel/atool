package ikds4

type prodT struct {
	gas   string
	scale float64
	index int
	limD  float64
}

var prodTypes = map[string]prodT{
	"СО2-2": {
		gas:   "CO2",
		scale: 2,
		index: 1,
		limD:  0.1,
	},
	"СО2-4": {
		gas:   "CO2",
		scale: 4,
		index: 2,
		limD:  0.25,
	},
	"СО2-10": {
		gas:   "CO2",
		scale: 10,
		index: 3,
		limD:  0.5,
	},
	"CH4-100": {
		gas:   "CH4",
		scale: 100,
		index: 4,
	},
	"CH4-100НКПР": {
		gas:   "CH4",
		scale: 100,
		index: 5,
	},
	"C3H8-100": {
		gas:   "C3H8",
		scale: 100,
		index: 6,
	},
}
