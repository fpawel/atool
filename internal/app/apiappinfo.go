package app

import (
	"context"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
)

type appInfoSvc struct{}

var _ api.AppInfoService = new(appInfoSvc)

func (*appInfoSvc) BuildInfo(context.Context) (*apitypes.BuildInfo, error) {
	orNone := func(s string) string {
		if s != ""{
			return s
		}
		return "DEVELOP"
	}
	return &apitypes.BuildInfo{
		Commit:    orNone(buildInfo.Commit),
		CommitGui: orNone(buildInfo.CommitGui),
		UUID:      orNone(buildInfo.UUID),
		Date:      orNone(buildInfo.Date),
		Time:      orNone(buildInfo.Time),
	}, nil
}
