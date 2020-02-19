package devdata

type DataSections []DataSection

type DataSection struct {
	Name   string
	Params []DataParam
}

type DataParam struct {
	Key, Name string
}

func (xs DataSections) Keys() map[string]struct{} {
	r := map[string]struct{}{}
	for _, x := range xs {
		for _, p := range x.Params {
			r[p.Key] = struct{}{}
		}
	}
	return r
}

func (xs DataSections) HasKey(key string) bool {
	for _, x := range xs {
		for _, p := range x.Params {
			if p.Key == key {
				return true
			}
		}
	}
	return false
}
