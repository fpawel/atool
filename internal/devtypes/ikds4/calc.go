package ikds4

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/devtypes/devdata"
	"github.com/fpawel/atool/internal/devtypes/mil82"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/must"
	"math"
)

func calcSections(party data.PartyValues, sections *devdata.CalcSections) error {
	prodT, ok := prodTypes[party.ProductType]
	if !ok {
		return merry.Errorf("не правильное исполнение ИКД-С4: %s", party.ProductType)
	}
	for _, pt := range []section{
		{key: "test_t_norm", name: "НКУ"},
		{key: "test_t_low", name: "Т-", tNorm: ptrFloat(20)},
		{key: "test_t_high", name: "Т+", tNorm: ptrFloat(20)},
		{key: "test2", name: "возврат НКУ"},
		{key: "tex1", name: "перед техпрогоном"},
		{key: "tex1", name: "после техпрогона"},
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

				valueKey := keyGasVar(pt.key, gas, 0)
				value := V(valueKey)

				info := map[string]interface{}{
					"исполнение":              party.ProductType + ", " + prodT.Gas,
					fmt.Sprintf("ПГС%d", gas): jsonNaN(pgs),
					"концентрация":            fmt.Sprintf("%s: %v", valueKey, jsonNaN(value)),
				}

				nominal := pgs
				absErrLimit20, var2, tNorm := math.NaN(), math.NaN(), math.NaN()

				if prodT.Gas == "CO2" {
					absErrLimit20 = prodT.limD
				} else {
					absErrLimit20 = 2.5 + 0.05*nominal
				}

				absErrLimit := absErrLimit20

				if pt.tNorm != nil {
					info["номинал"] = jsonNaN(nominal)
					info["предел при 20⁰C"] = jsonNaN(absErrLimit20)
					tNorm = *pt.tNorm

					k := keyGasVar("test_t_norm", gas, 0)
					nominal = V(k)
					info["номинал"] = fmt.Sprintf("%s: %v", k, nominal)

					info["Tn"] = tNorm

					var2k := keyGasVar(pt.key, gas, 2)
					var2 = V(var2k)

					info["T"] = fmt.Sprintf("%s: %v", var2k, var2)

					if prodT.Gas == "CO2" {
						info["расчёт_предела"] = fmt.Sprintf("CO2: LIMIT20 * 0.5 * 0.1 * |T-Tn| = %v * 0.5 * 0.1 * |%v-%v|",
							absErrLimit, var2, tNorm)
						absErrLimit = 0.5 * math.Abs(absErrLimit*(var2-tNorm)) / 10

					} else {
						if gas == 1 {
							absErrLimit = 5
							info["расчёт_предела"] = "CH: ПГС1: 5"
						} else {
							absErrLimit = math.Abs(nominal * 0.15)
							info["расчёт_предела"] = fmt.Sprintf("CH: 0.15 * Cn = 0.15 * %v", nominal)
						}
					}
				}
				info["предел"] = jsonNaN(round3(absErrLimit))

				absErr := value - nominal
				info["погрешность"] = jsonNaN(round3(absErr))

				relErr := 100 * absErr / absErrLimit

				v := devdata.AddValue(prm)

				v.Detail = string(must.MarshalJsonIndent(info, "", "\t"))

				if !math.IsNaN(relErr) {
					v.Validated = true
					v.Valid = math.Abs(absErr) < math.Abs(absErrLimit)
					v.Value = pkg.FormatFloat(relErr, 2)
				}
			}
		}
	}
	mil82.AddSectionProdOut(party, sections)
	return nil
}

func round3(v float64) float64 {
	return math.Round(v*1000) / 1000
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
