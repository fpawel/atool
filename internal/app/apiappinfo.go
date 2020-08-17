package app

import (
	"context"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
)

type appInfoSvc struct{}

var _ api.AppInfoService = new(appInfoSvc)

func (*appInfoSvc) BuildInfo(context.Context) (*apitypes.BuildInfo, error) {
	return &apitypes.BuildInfo{
		Commit:    buildInfo.Commit,
		CommitGui: buildInfo.CommitGui,
		UUID:      buildInfo.UUID,
		Date:      buildInfo.Date,
		Time:      buildInfo.Time,
	}, nil
}
