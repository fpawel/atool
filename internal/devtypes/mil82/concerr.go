package mil82

import (
	"fmt"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/devtypes/devdata"
	"github.com/fpawel/atool/internal/pkg/must"
	"math"
	"strconv"
)

func getProdOut(party data.PartyValues, sections *devdata.CalcSections) {
	sect := devdata.AddSect(sections, "Выпуск в эксплуатацию")
	dec2 := func(p data.ProductValues, k int) (int, int, bool) {
		v, ok := p.Values[fmt.Sprintf("K%d", k)]
		if !ok {
			return 0, 0, false
		}
		n := int(v)
		return n / 10000, n % 10000, true
	}
	prmYear := devdata.AddParam(sect, "Год")
	prmMonth := devdata.AddParam(sect, "Месяц")
	prmSerial := devdata.AddParam(sect, "Сер.№")
	prmKind := devdata.AddParam(sect, "Исполнение")
	for _, p := range party.Products {
		vYear := devdata.AddValue(prmYear)
		vMonth := devdata.AddValue(prmMonth)
		vSerial := devdata.AddValue(prmSerial)
		vKind := devdata.AddValue(prmKind)
		if y, s, f := dec2(p, 40); f && y > 0 {
			vYear.Value = strconv.Itoa(y + 2000)
			vSerial.Value = strconv.Itoa(s)
		}
		if m, i, f := dec2(p, 47); f && m > 0 && i > 0 {
			vMonth.Value = strconv.Itoa(m)
			vKind.Value = strconv.Itoa(i)
			for name, pt := range prodTypes {
				if pt.index == i {
					vKind.Value = name
				}
			}
		}
	}
}

func getConcentrationErrors(party data.PartyValues, sections *devdata.CalcSections) error {
	prodT, ok := prodTypes[party.ProductType]
	if !ok {
		return fmt.Errorf("не правильное исполнение МИЛ-82: %s", party.ProductType)
	}
	for _, pt := range []section{
		{key: "test_t_norm", name: "НКУ"},
		{key: "test_t_low", name: "Т-", tNorm: ptrFloat(20)},
		{key: "test_t_high", name: "Т+", tNorm: ptrFloat(20)},
		{key: "test2", name: "возврат НКУ"},
		{key: "test_t80", name: "90⁰C", tNorm: ptrFloat(80)},
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
					"исполнение":              party.ProductType + ", " + prodT.gas,
					fmt.Sprintf("ПГС%d", gas): jsonNaN(pgs),
					"концентрация":            fmt.Sprintf("%s: %v", valueKey, jsonNaN(value)),
				}

				nominal := pgs
				absErrLimit20, var2, tNorm := math.NaN(), math.NaN(), math.NaN()
				if prodT.gas == "CO2" {
					switch prodT.scale {
					case 4:
						absErrLimit20 = 0.2 + 0.05*nominal
					case 10:
						absErrLimit20 = 0.5
					case 20:
						absErrLimit20 = 1
					default:
						absErrLimit20 = math.NaN()
					}
				} else {
					absErrLimit20 = 2.5 + 0.05*nominal
				}

				absErrLimit := absErrLimit20

				if pt.tNorm != nil {
					info["номинал"] = jsonNaN(nominal)
					info["предел при 20⁰C"] = jsonNaN(absErrLimit20)
					tNorm = *pt.tNorm

					if tNorm == 80 {
						k := keyGasVar("test_t80", gas, 0)
						nominal = V(k)
						info["номинал"] = fmt.Sprintf("%s: %v", k, nominal)
					} else {
						k := keyGasVar("test_t_norm", gas, 0)
						nominal = V(k)
						info["номинал"] = fmt.Sprintf("%s: %v", k, nominal)
					}
					info["Tn"] = tNorm

					var2k := keyGasVar(pt.key, gas, 2)
					var2 = V(var2k)

					info["T"] = fmt.Sprintf("%s: %v", var2k, var2)

					if prodT.gas == "CO2" {
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
				info["предел"] = jsonNaN(absErrLimit)

				absErr := value - nominal
				info["погрешность"] = jsonNaN(absErr)

				relErr := 100 * absErr / absErrLimit

				v := devdata.AddValue(prm)

				v.Detail = string(must.MarshalJsonIndent(info, "", "\t"))

				if !math.IsNaN(relErr) {
					v.Validated = true
					v.Valid = math.Abs(absErr) < math.Abs(absErrLimit)
					v.Value = fmt.Sprintf("%.2f", relErr)
				}
			}
		}
	}
	getProdOut(party, sections)
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
