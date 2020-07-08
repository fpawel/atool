package main

import (
	"github.com/fpawel/atool/internal/app"
	"github.com/fpawel/atool/internal/pkg"
)

var (
	GitCommit string
	BuildUUID string
	BuildDate string
	BuildTime string
)

func main() {
	pkg.InitLog()
	app.Main(app.BuildInfo{
		Commit: GitCommit,
		UUID:   BuildUUID,
		Date:   BuildDate,
		Time:   BuildTime,
	})
}
