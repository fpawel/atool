package calcmil82

import (
	"fmt"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/devdata/devcalc"
	"math"
)

func Calc(party data.PartyValues, calc *devcalc.CalcSections) error {
	prodT, ok := prodTypes[party.ProductType]
	if !ok {
		return fmt.Errorf("не правильное исполнение МИЛ-82: %s", party.ProductType)
	}
	for ptKey, pt := range sections {
		sect := calc.AddSect(pt.name)
		for _, gas := range []int{1, 3, 4} {
			prm := sect.AddPrm(fmt.Sprintf("газ %d", gas))
			nominal := valOrNaN(party.Values, fmt.Sprintf("c%d", gas))
			for _, p := range party.Products {
				V := func(k string) float64 {
					return valOrNaN(p.Values, k)
				}
				absErrLim := math.NaN()
				switch prodT.gas {
				case "CO2":
					switch prodT.scale {
					case 4:
						absErrLim = 0.2 + 0.05*nominal
					case 10:
						absErrLim = 0.5
					case 20:
						absErrLim = 1
					}
				default:
					absErrLim = 2.5 + 0.05*nominal
				}
			}

		}
	}
}

func valOrNaN(m map[string]float64, k string) float64 {
	if v, ok := m[k]; ok {
		return v
	}
	return math.NaN()
}

type section struct {
	key, name string
	t         *float64
}

func ptrFloat(v float64) *float64 {
	return &v
}

var sections = []section{
	{key: "test_t_norm", name: "нормальная температура"},
	{key: "test_t_low", name: "низкая температура", t: ptrFloat(20)},
	{key: "test_t_high", name: "высокая температура", t: ptrFloat(20)},
	{key: "test2", name: "возврат НКУ"},
	{key: "test_t80", name: "80⁰C", t: ptrFloat(80)},
	{key: "tex1", name: "перед технологическим прогоном"},
	{key: "tex1", name: "после технологического прогона"},
}
