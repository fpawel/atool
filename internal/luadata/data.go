package luadata

type DataSections []*DataSection

type DataSection struct {
	Name   string
	Params []DataParam
}

type DataParam = [2]string

func (x *DataSection) AddParam(key, name string) {
	x.Params = append(x.Params, DataParam{key, name})
}

func (xs DataSections) Keys() map[string]struct{} {
	r := map[string]struct{}{}
	for _, x := range xs {
		for _, p := range x.Params {
			r[p[0]] = struct{}{}
		}
	}
	return r
}

func (xs DataSections) HasKey(key string) bool {
	for _, x := range xs {
		for _, p := range x.Params {
			if p[0] == key {
				return true
			}
		}
	}
	return false
}
