package ikds4

type productType struct {
	Name  string
	Gas   string
	Scale float64
	Index int
	limD  float64
}

var (
	prodTypesList = []productType{
		{
			Name:  "CO2-2",
			Gas:   "CO2",
			Scale: 2,
			Index: 1,
			limD:  0.1,
		},
		{
			Name:  "CO2-4",
			Gas:   "CO2",
			Scale: 4,
			Index: 2,
			limD:  0.25,
		},
		{
			Name:  "CO2-10",
			Gas:   "CO2",
			Scale: 10,
			Index: 3,
			limD:  0.5,
		},
		{
			Name:  "CH4-100",
			Gas:   "CH4",
			Scale: 100,
			Index: 4,
		},
		{
			Name:  "CH4-100НКПР",
			Gas:   "CH4",
			Scale: 100,
			Index: 5,
		},
		{
			Name:  "C3H8-100",
			Gas:   "C3H8",
			Scale: 100,
			Index: 6,
		},
	}
	prodTypes, prodTypeNames = initProductTypes(prodTypesList)
)

func initProductTypes(prodTypesList []productType) (m map[string]productType, xs []string) {
	m = map[string]productType{}
	for i := range prodTypesList {
		t := &prodTypesList[i]
		t.Index = i + 1
		m[t.Name] = *t
		xs = append(xs, t.Name)
	}
	return
}
