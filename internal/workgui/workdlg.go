package workgui

import (
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
)

type ConfigParamValue = *apitypes.ConfigParamValue

var (
	ConfigParamValues []ConfigParamValue
	ChanSelectedWorks = make(chan []bool)
)

func (x Works) ExecuteSelectWorksDialog(done <-chan struct{}) (result Works) {
	var names = make([]string, len(x))
	for i := range x {
		names[i] = x[i].Name
	}

	go gui.ExecuteSelectWorksDialog(names)

	select {
	case <-done:
		return
	case xs := <-ChanSelectedWorks:
		for i, f := range xs {
			if f {
				result = append(result, Work{
					Name: x[i].Name,
					Func: x[i].Func,
				})
			}
		}
	}
	return
}
