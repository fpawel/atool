package workgui

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/comports"
	"github.com/fpawel/comm"
	"github.com/powerman/structlog"
	"sync/atomic"
)

type WorkFunc func(log comm.Logger, ctx context.Context) error

type Work struct {
	Name string
	Func WorkFunc
}

type Works []Work

func New(name string, work WorkFunc) Work {
	return Work{name, work}
}

func NewFunc(name string, work WorkFunc) WorkFunc {
	return Work{name, work}.Perform
}

func NewWorks(works ...Work) Works {
	return works
}

func (x WorkFunc) Work(name string) Work {
	return New(name, x)
}

func (x WorkFunc) DoWarn(log comm.Logger, ctx context.Context) error {
	err := x(log, ctx)
	if err == nil || merry.Is(err, context.Canceled) {
		return err
	}
	var ctxIgnoreError context.Context
	ctxIgnoreError, ignoreError = context.WithCancel(ctx)
	NotifyWorkSuspended(err)
	<-ctxIgnoreError.Done()
	ignoreError()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	NotifyWarn(log, "ошибка проигнорирована")
	return nil
}

func (x Work) WithWarn() Work {
	return Work{
		Name: x.Name,
		Func: x.Func.DoWarn,
	}
}

func (x Work) Perform(log *structlog.Logger, ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	NotifyInfo(log, "🛠 "+x.Name)

	muNamedWorksStack.Lock()
	isMainWork := len(namedWorksStack) == 0
	namedWorksStack = append(namedWorksStack, x.Name)
	level := len(namedWorksStack)
	muNamedWorksStack.Unlock()

	log = pkg.LogPrependSuffixKeys(log, fmt.Sprintf("work%d", level), x.Name)

	err := x.Func(log, ctx)

	muNamedWorksStack.Lock()
	if len(namedWorksStack) > 0 {
		namedWorksStack = namedWorksStack[:len(namedWorksStack)-1]
	}
	muNamedWorksStack.Unlock()

	if err != nil {
		if isMainWork {
			pkg.LogPrependSuffixKeys(log, "stack", pkg.FormatStacktrace(merry.Stack(err), "\n\t")).PrintErr(err)
		}
		err = merry.Prepend(err, "🚫 "+x.Name)
		NotifyErr(log, err)
		return err
	}

	if isMainWork {
		NotifyInfo(log, "✅ "+x.Name)
	}
	return nil
}

func (x Work) Run(log *structlog.Logger, ctx context.Context) error {
	if IsConnected() {
		return merry.New("already connected")
	}
	wg.Add(1)
	atomic.StoreInt32(&atomicConnected, 1)
	ctx, interrupt = context.WithCancel(ctx)
	go x.run(log, ctx)
	return nil
}

func (x Works) Work(name string) Work {
	return Work{
		Name: name,
		Func: x.Do,
	}
}

func (x Works) Do(log comm.Logger, ctx context.Context) error {
	for _, w := range x {
		w := w
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := w.Perform(log, ctx); err != nil {
			return err
		}
	}
	return nil
}

type WorkFuncList []WorkFunc

func NewWorkFuncList(args ...WorkFunc) WorkFuncList {
	return args
}

func (xs WorkFuncList) Do(log *structlog.Logger, ctx context.Context) error {
	for _, w := range xs {
		w := w
		if err := w(log, ctx); err != nil {
			return err
		}
	}
	return nil
}

func (x Work) run(log *structlog.Logger, ctx context.Context) {
	go gui.NotifyStartWork()

	muNamedWorksStack.Lock()
	namedWorksStack = nil
	muNamedWorksStack.Unlock()

	_ = x.Perform(log, ctx)

	interrupt()
	atomic.StoreInt32(&atomicConnected, 0)
	comports.CloseAllComports()
	wg.Done()
	go gui.NotifyStopWork()
}
