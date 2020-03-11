package mil82

import (
	"github.com/fpawel/atool/internal/devtypes/devdata"
	"sort"
)

var Device = devdata.Device{

	GetCalcSectionsFunc: getConcentrationErrors,

	ProductTypes: func() (result []string) {
		for name := range prodTypes {
			result = append(result, name)
		}
		sort.Strings(result)
		return
	}(),

	DataSections: dataSections(),
}
