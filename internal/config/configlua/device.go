package configlua

type Device struct {
	Sections Sections
}

func (x *Device) AddSection(name string) *Section {
	s := &Section{Name: name}
	x.Sections = append(x.Sections, s)
	return s
}

type Section struct {
	Name   string
	Params []Param
}

func (x *Section) AddParam(key, name string) {
	x.Params = append(x.Params, Param{
		Key:  key,
		Name: name,
	})
}

type Param struct {
	Key  string
	Name string
}

type Sections []*Section

func (xs Sections) Keys() map[string]struct{} {
	r := map[string]struct{}{}
	for _, x := range xs {
		for _, p := range x.Params {
			r[p.Key] = struct{}{}
		}
	}
	return r
}

func (xs Sections) HasKey(key string) bool {
	for _, x := range xs {
		for _, p := range x.Params {
			if p.Key == key {
				return true
			}
		}
	}
	return false
}

type devices struct {
	xs map[string]*Device
}

func (x devices) Add(name string) *Device {
	x.xs[name] = new(Device)
	return x.xs[name]
}
