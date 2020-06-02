package devtypes

import (
	"github.com/fpawel/atool/internal/devtypes/devdata"
	"github.com/fpawel/atool/internal/devtypes/ikds4"
	"github.com/fpawel/atool/internal/devtypes/mil82"
)

var (
	DeviceTypes = map[string]devdata.Device{
		"МИЛ-82": mil82.Device,
		"ИКД-С4": ikds4.Device,
	}
)

func init() {
	for name, d := range DeviceTypes {
		d.Name = name
		d.ListProductTypes()
		DeviceTypes[name] = d
	}
}
