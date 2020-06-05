package ikds4

type prodT struct {
	Gas   string  `json:"gas"`
	Scale float64 `json:"scale"`
	Index int     `json:"index"`
	limD  float64
}

var prodTypes = map[string]prodT{
	"CO2-2": {
		Gas:   "CO2",
		Scale: 2,
		Index: 1,
		limD:  0.1,
	},
	"CO2-4": {
		Gas:   "CO2",
		Scale: 4,
		Index: 2,
		limD:  0.25,
	},
	"CO2-10": {
		Gas:   "CO2",
		Scale: 10,
		Index: 3,
		limD:  0.5,
	},
	"CH4-100": {
		Gas:   "CH4",
		Scale: 100,
		Index: 4,
	},
	"CH4-100НКПР": {
		Gas:   "CH4",
		Scale: 100,
		Index: 5,
	},
	"C3H8-100": {
		Gas:   "C3H8",
		Scale: 100,
		Index: 6,
	},
}
