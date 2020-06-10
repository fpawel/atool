package workgui

import (
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
)

type ConfigParamValue = *apitypes.ConfigParamValue

var (
	ConfigParamValues []ConfigParamValue
	ChanSelectedWorks = make(chan []bool)
)
