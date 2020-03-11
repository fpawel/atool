package devtypes

import (
	"github.com/fpawel/atool/internal/devtypes/devdata"
	"github.com/fpawel/atool/internal/devtypes/mil82"
)

var DeviceTypes = map[string]devdata.Device{
	"МИЛ-82": mil82.Device,
}
