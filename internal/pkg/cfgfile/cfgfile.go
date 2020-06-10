package cfgfile

import (
	"github.com/ansel1/merry"
	"io/ioutil"
	"os"
	"path/filepath"
)

type MarshalFunc = func(in interface{}) (out []byte, err error)
type UnmarshalFunc = func(in []byte, out interface{}) error

type F struct {
	name      string
	marshal   MarshalFunc
	unmarshal UnmarshalFunc
}

func New(name string, marshal MarshalFunc, unmarshal UnmarshalFunc) *F {
	return &F{
		name:      name,
		marshal:   marshal,
		unmarshal: unmarshal,
	}
}

func (x *F) Set(in interface{}) error {
	data, err := x.marshal(in)
	if err != nil {
		return x.err(err)
	}
	if err := ioutil.WriteFile(x.Filename(), data, 0666); err != nil {
		return x.err(err)
	}
	return nil
}

func (x *F) Get(out interface{}) error {
	data, err := ioutil.ReadFile(x.Filename())
	if err != nil {
		return err
	}
	return x.err(x.unmarshal(data, out))
}

func (x *F) err(err error) error {
	return merry.Append(err, x.name)
}

func (x *F) Filename() string {
	return filepath.Join(filepath.Dir(os.Args[0]), x.name)
}
