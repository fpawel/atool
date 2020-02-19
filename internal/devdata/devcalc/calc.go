package devcalc

import "github.com/fpawel/atool/internal/thriftgen/apitypes"

type CalcSections []*CalcSection
type CalcParam = apitypes.CalcParam
type CalcValue = apitypes.CalcValue
type CalcSection = apitypes.CalcSection

func AddSect(x *CalcSections, name string) *CalcSection {
	c := &CalcSection{Name: name}
	*x = append(*x, c)
	return c
}

func AddPrm(x *CalcSection, name string) *CalcParam {
	v := &CalcParam{Name: name}
	x.Params = append(x.Params, v)
	return v
}

func AddVal(x *CalcParam) *CalcValue {
	v := new(CalcValue)
	x.Values = append(x.Values, v)
	return v
}
