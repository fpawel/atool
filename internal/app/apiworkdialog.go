package app

import (
	"context"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/atool/internal/workgui"
)

type workDialogSvc struct{}

var _ api.WorkDialogService = new(workDialogSvc)

func (_ *workDialogSvc) SelectWork(_ context.Context, workIndex int32) (err error) {
	workgui.ChanSelectedWork <- int(workIndex)
	return nil
}

func (_ *workDialogSvc) SelectWorks(_ context.Context, works []bool) (err error) {
	workgui.ChanSelectedWorks <- works
	return nil
}

func (_ *workDialogSvc) IgnoreError(context.Context) error {
	workgui.IgnoreError()
	return nil
}

func (_ *workDialogSvc) SetConfigParamValues(_ context.Context, configParamValues []*apitypes.ConfigParamValue) error {
	workgui.ConfigParamValues = configParamValues
	return nil
}

func (_ *workDialogSvc) GetConfigParamValues(_ context.Context) ([]*apitypes.ConfigParamValue, error) {
	return workgui.ConfigParamValues, nil
}
