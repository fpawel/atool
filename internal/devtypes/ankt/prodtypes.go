package ankt

var (
	productTypesList = []productType{
		prodT1(10, "CO₂", 2, true),
		prodT1(10, "CO₂", 5, true),
		prodT1(10, "CO₂", 10, true),
		prodT1(12, "∑CH", 100, true),
		prodT1(13, "C₃H₈", 100, true),
		prodT1(14, "CH₄", 100, true),
		prodT1(15, "CO₂", 2, true),
		prodT1(15, "CO₂", 5, true),
		prodT1(15, "CO₂", 10, true),
		prodT1(15, "CH₄", 100, true),
		prodT1(15, "C₃H₈", 100, true),
		prodT1(15, "∑CH", 100, true),
		prodT1(16, "CO₂", 2, true),
		prodT1(16, "CO₂", 5, true),
		prodT1(16, "CO₂", 10, true),
		prodT1(16, "CH₄", 100, true),
		prodT1(16, "C₃H₈", 100, true),
		prodT1(16, "∑CH", 100, true),
		prodT1(26, "CH₄", 100, false),
		prodT1(26, "C₃H₈", 100, false),
		prodT1(26, "∑CH", 100, false),
		prodT1(27, "CH₄", 100, false),
		prodT1(27, "C₃H₈", 100, false),
		prodT1(27, "∑CH", 100, false),
		prodT1(28, "CH₄", 100, false),
		prodT1(28, "C₃H₈", 100, false),
		prodT1(28, "∑CH", 100, false),
		prodT1(29, "CH₄", 100, false),
		prodT1(29, "C₃H₈", 100, false),
		prodT1(29, "∑CH", 100, false),
		prodT1(30, "CH₄", 100, false),
		prodT1(30, "C₃H₈", 100, false),
		prodT1(30, "∑CH", 100, false),
		prodT1(31, "CH₄", 100, false),
		prodT1(31, "C₃H₈", 100, false),
		prodT1(31, "∑CH", 100, false),
		prodT1(32, "CH₄", 100, false),
		prodT1(32, "C₃H₈", 100, false),
		prodT1(32, "∑CH", 100, false),
		prodT1(42, "CH₄", 100, false),
		prodT1(42, "C₃H₈", 100, false),
		prodT1(42, "∑CH", 100, false),
		prodT1(43, "CH₄", 100, false),
		prodT1(43, "C₃H₈", 100, false),
		prodT1(43, "∑CH", 100, false),
		prodT1(44, "CH₄", 100, false),
		prodT1(44, "C₃H₈", 100, false),
		prodT1(44, "∑CH", 100, false),

		prodT2(11, "CO₂", 2, "CH₄", 100, true),
		prodT2(11, "CO₂", 5, "CH₄", 100, true),
		prodT2(11, "CO₂", 10, "CH₄", 100, true),
		prodT2(33, "CO₂", 2, "CH₄", 100, false),
		prodT2(33, "CO₂", 5, "CH₄", 100, false),
		prodT2(33, "CO₂", 10, "CH₄", 100, false),
		prodT2(33, "CH₄", 100, "CH₄", 100, false),
		prodT2(33, "C₃H₈", 100, "CH₄", 100, false),
		prodT2(33, "∑CH", 100, "CH₄", 100, false),
		prodT2(34, "CO₂", 2, "CH₄", 100, false),
		prodT2(34, "CO₂", 5, "CH₄", 100, false),
		prodT2(34, "CO₂", 10, "CH₄", 100, false),
		prodT2(34, "CH₄", 100, "CH₄", 100, false),
		prodT2(34, "C₃H₈", 100, "CH₄", 100, false),
		prodT2(34, "∑CH", 100, "CH₄", 100, false),
		prodT2(35, "CO₂", 2, "CH₄", 100, false),
		prodT2(35, "CO₂", 5, "CH₄", 100, false),
		prodT2(35, "CO₂", 10, "CH₄", 100, false),
		prodT2(35, "CH₄", 100, "CH₄", 100, false),
		prodT2(35, "C₃H₈", 100, "CH₄", 100, false),
		prodT2(35, "∑CH", 100, "CH₄", 100, false),
	}

	productTypes = func() (xs map[string]productType) {
		xs = make(map[string]productType)
		for _, x := range productTypesList {
			xs[x.String()] = x
		}
		return
	}()
)

func prodT1(n int, gas gasT, scale float64, pressure bool) productType {
	return productType{
		N:        n,
		Pressure: pressure,
		Chan2:    false,
		Chan: [2]chanT{
			{
				gas:   gas,
				scale: scale,
			},
		},
	}
}

func prodT2(n int, gas gasT, scale float64, gas2 gasT, scale2 float64, pressure bool) productType {
	return productType{
		N:        n,
		Pressure: pressure,
		Chan2:    true,
		Chan: [2]chanT{
			{
				gas:   gas,
				scale: scale,
			},
			{
				gas:   gas2,
				scale: scale2,
			},
		},
	}
}
