package mil82

type productType struct {
	Name  string
	Gas   string
	Scale float64
	Index int
}

var (
	prodTypesList = []productType{
		{
			Name:  "00.00",
			Gas:   "CO2",
			Scale: 4,
		},
		{
			Name:  "00.01",
			Gas:   "CO2",
			Scale: 10,
		},
		{
			Name:  "00.02",
			Gas:   "CO2",
			Scale: 20,
		},
		{
			Name:  "01.00",
			Gas:   "CH4",
			Scale: 100,
		},
		{
			Name:  "01.01",
			Gas:   "CH4",
			Scale: 100,
		},
		{
			Name:  "02.00",
			Gas:   "C3H8",
			Scale: 50,
		},
		{
			Name:  "02.01",
			Gas:   "C3H8",
			Scale: 50,
		},
		{
			Name:  "03.00",
			Gas:   "C3H8",
			Scale: 100,
		},
		{
			Name:  "03.01",
			Gas:   "C3H8",
			Scale: 100,
		},
		{
			Name:  "04.00",
			Gas:   "CH4",
			Scale: 100,
		},
		{
			Name:  "05.00",
			Gas:   "C6H14",
			Scale: 50,
		},
		{
			Name:  "10.00",
			Gas:   "CO2",
			Scale: 4,
		},
		{
			Name:  "10.01",
			Gas:   "CO2",
			Scale: 10,
		},
		{
			Name:  "10.02",
			Gas:   "CO2",
			Scale: 20,
		},
		{
			Name:  "10.03",
			Gas:   "CO2",
			Scale: 4,
		},
		{
			Name:  "10.04",
			Gas:   "CO2",
			Scale: 10,
		},
		{
			Name:  "10.05",
			Gas:   "CO2",
			Scale: 20,
		},
		{
			Name:  "11.00",
			Gas:   "CH4",
			Scale: 100,
		},
		{
			Name:  "11.01",
			Gas:   "CH4",
			Scale: 100,
		},
		{
			Name:  "13.00",
			Gas:   "C3H8",
			Scale: 100,
		},
		{
			Name:  "13.01",
			Gas:   "C3H8",
			Scale: 100,
		},
		{
			Name:  "14.00",
			Gas:   "CH4",
			Scale: 100,
		},
		{
			Name:  "16.00",
			Gas:   "C3H8",
			Scale: 100,
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
