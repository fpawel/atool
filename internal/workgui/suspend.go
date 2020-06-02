package workgui

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
)

func SuspendWork(log comm.Logger, ctx context.Context, err error) {
	var ctxResumeWork context.Context
	ctxResumeWork, ResumeWork = context.WithCancel(ctx)
	NotifyWorkSuspended(err)
	<-ctxResumeWork.Done()
	ResumeWork()
	if ctx.Err() == nil {
		NotifyErr(log, merry.Prepend(err, "ошибка проигнорирована: выполнение продолжено"))
	}
}

var ResumeWork = func() {}
