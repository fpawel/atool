package devcalc

type CalcSections []*CalcSection

func (x *CalcSections) AddSect(name string) *CalcSection {
	c := &CalcSection{Name: name}
	*x = append(*x, c)
	return c
}

type CalcSection struct {
	Name   string
	Params []*CalcParam
}

func (x *CalcSection) AddPrm(name string) *CalcParam {
	v := &CalcParam{Name: name}
	x.Params = append(x.Params, v)
	return v
}

type CalcParam struct {
	Name   string
	Values []*CalcValue
}

func (x *CalcParam) AddVal() *CalcValue {
	v := new(CalcValue)
	x.Values = append(x.Values, v)
	return v
}

type CalcValue struct {
	Validated bool
	Valid     bool
	Detail    string
	Value     string
}
