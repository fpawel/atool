package mil82

import (
	"fmt"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/devtypes/devdata"
	"github.com/fpawel/atool/internal/pkg/must"
	"math"
)

func getConcentrationErrors(party data.PartyValues, sections *devdata.CalcSections) error {
	prodT, ok := prodTypes[party.ProductType]
	if !ok {
		return fmt.Errorf("не правильное исполнение МИЛ-82: %s", party.ProductType)
	}
	for _, pt := range []section{
		{key: "test_t_norm", name: "нормальная температура"},
		{key: "test_t_low", name: "низкая температура", tNorm: ptrFloat(20)},
		{key: "test_t_high", name: "высокая температура", tNorm: ptrFloat(20)},
		{key: "test2", name: "возврат НКУ"},
		{key: "test_t80", name: "80⁰C", tNorm: ptrFloat(80)},
		{key: "tex1", name: "перед технологическим прогоном"},
		{key: "tex1", name: "после технологического прогона"},
	} {
		sect := devdata.AddSect(sections, "Расчёт погрешности: "+pt.name)
		gases := []int{1, 3, 4}
		if pt.key == "test_t_norm" {
			gases = []int{1, 2, 3, 4}
		}
		for _, gas := range gases {
			prm := devdata.AddParam(sect, fmt.Sprintf("газ %d", gas))
			pgs := valOrNaN(party.Values, fmt.Sprintf("c%d", gas))

			for _, p := range party.Products {
				V := func(k string) float64 {
					return valOrNaN(p.Values, k)
				}
				nominal := pgs
				absErrLim, var2, tNorm := math.NaN(), math.NaN(), math.NaN()
				if prodT.gas == "CO2" {
					switch prodT.scale {
					case 4:
						absErrLim = 0.2 + 0.05*nominal
					case 10:
						absErrLim = 0.5
					case 20:
						absErrLim = 1
					default:
						absErrLim = math.NaN()
					}
				} else {
					absErrLim = 2.5 + 0.05*nominal
				}
				if pt.tNorm != nil {
					tNorm = *pt.tNorm
					k := "test_t_norm"
					if tNorm == 80 {
						k = "test_t80"
					}
					nominal = V(keyGasVar(k, gas, 0))
					var2 = V(keyGasVar(k, gas, 2))
					if prodT.gas == "CO2" {
						absErrLim = 0.5 * math.Abs(absErrLim*(var2-tNorm)) / 10
					} else {
						if gas == 1 {
							absErrLim = 5
						} else {
							absErrLim = math.Abs(nominal * 0.15)
						}
					}
				}
				valueKey := keyGasVar(pt.key, gas, 0)
				value := V(valueKey)
				absErr := value - nominal
				relErr := 100 * absErr / absErrLim

				v := devdata.AddValue(prm)

				v.Detail = string(must.MarshalJsonIndent(map[string]interface{}{
					"газ":                gas,
					"концентрация":       jsonNaN(value),
					"номинал":            jsonNaN(nominal),
					"погрешность":        jsonNaN(absErr),
					"предел погрешности": jsonNaN(absErrLim),
					"db_key":             valueKey,
					"tNorm":              jsonNaN(tNorm),
					"product_type":       party.ProductType,
					"gas":                prodT.gas,
					"ПГС":                jsonNaN(pgs),
					"var2":               jsonNaN(var2),
				}, "", "\t"))

				if !math.IsNaN(relErr) {
					v.Validated = true
					v.Valid = math.Abs(absErr) < math.Abs(absErrLim)
					v.Value = fmt.Sprintf("%.2f", relErr)
				}
			}
		}
	}
	return nil
}

func jsonNaN(v float64) interface{} {
	if math.IsNaN(v) {
		return nil
	}
	return v
}

func keyGasVar(k string, gas, Var int) string {
	return fmt.Sprintf("%s_gas%d_var%d", k, gas, Var)
}

func valOrNaN(m map[string]float64, k string) float64 {
	if v, ok := m[k]; ok {
		return v
	}
	return math.NaN()
}

type section struct {
	key, name string
	tNorm     *float64
}

func ptrFloat(v float64) *float64 {
	return &v
}
