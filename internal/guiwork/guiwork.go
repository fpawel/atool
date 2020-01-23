package guiwork

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/journal"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/comports"
	"github.com/powerman/structlog"
	"sync"
	"sync/atomic"
	"time"
)

type WorkFunc func(*structlog.Logger, context.Context) error

type DelayBackgroundWorkFunc func(*structlog.Logger, context.Context) error

func IsConnected() bool {
	return atomic.LoadInt32(&atomicConnected) != 0
}

func Interrupt() {
	interrupt()
}

func Wait() {
	wg.Wait()
}

func RunWork(log *structlog.Logger, ctx context.Context, workName string, work WorkFunc) error {
	if IsConnected() {
		return merry.New("already connected")
	}
	wg.Add(1)
	atomic.StoreInt32(&atomicConnected, 1)
	ctx, interrupt = context.WithCancel(ctx)
	go performWork(log, ctx, workName, work)
	return nil
}

func RunTask(log *structlog.Logger, what string, task func() error) {
	go func() {
		_ = PerformNewNamedWork(log, context.Background(), what, func(*structlog.Logger, context.Context) error {
			return task()
		})
	}()
}

func PerformNewNamedWork(log *structlog.Logger, ctx context.Context, newWorkName string, work WorkFunc) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	muNamedWorksStack.Lock()
	isMainWork := len(namedWorksStack) == 0
	namedWorksStack = append(namedWorksStack, newWorkName)
	level := len(namedWorksStack)
	muNamedWorksStack.Unlock()

	journal.Info(log, newWorkName+": выполняется")

	log = pkg.LogPrependSuffixKeys(log, fmt.Sprintf("work%d", level), newWorkName)

	err := work(log, ctx)

	muNamedWorksStack.Lock()
	namedWorksStack = namedWorksStack[:len(namedWorksStack)-1]
	muNamedWorksStack.Unlock()
	if err == nil {
		journal.Info(log, newWorkName+": выполнение окончено")
	} else {
		if isMainWork {
			log = pkg.LogPrependSuffixKeys(log, "stack", pkg.FormatMerryStacktrace(err), "error", err)
			err = merry.New(newWorkName + ": завершено с ошибкой").WithCause(err)
		} else {
			err = merry.Append(err, newWorkName)
		}
		journal.Err(log, err)
	}
	return err
}

func InterruptDelay(log *structlog.Logger) {
	muInterruptDelay.Lock()
	interruptDelay()
	muInterruptDelay.Unlock()
	journal.Warn(log, "текущая задержка прервана пользователем")
}

func Delay(log *structlog.Logger, ctx context.Context, duration time.Duration, name string, backgroundWork DelayBackgroundWorkFunc) error {

	if duration == 0 {
		return nil
	}

	startTime := time.Now()
	log = pkg.LogPrependSuffixKeys(log, "delay_start", startTime.Format("15:04:05"))

	// сохранить ссылку на изначальный контекст
	ctxParent := ctx

	// установить коллбэк прерывания задержки
	muInterruptDelay.Lock()
	ctx, interruptDelay = context.WithTimeout(ctx, duration)
	muInterruptDelay.Unlock()

	s1 := fmt.Sprintf("%s %s", name, duration)

	err := PerformNewNamedWork(log, ctx, s1, func(log *structlog.Logger, ctx context.Context) error {
		log.Info("delay: begin")
		go gui.NotifyBeginDelay(duration, s1)
		defer func() {
			muInterruptDelay.Lock()
			interruptDelay()
			muInterruptDelay.Unlock()

			log.Info("delay: end", "delay_elapsed", time.Since(startTime))
			go gui.NotifyEndDelay()
		}()

		for {
			err := backgroundWork(log, ctx)
			if ctxParent.Err() != nil {
				// выполнение прервано
				return ctxParent.Err()
			}
			if ctx.Err() != nil {
				// задержка истекла или прервана
				return nil
			}
			if err != nil {
				journal.Err(log, err)
				return nil
			}
		}
	})
	return err
}

func performWork(log *structlog.Logger, ctx context.Context, workName string, work WorkFunc) {
	go gui.NotifyStartWork()

	muNamedWorksStack.Lock()
	namedWorksStack = nil
	muNamedWorksStack.Unlock()

	_ = PerformNewNamedWork(log, ctx, workName, work)

	muNamedWorksStack.Lock()
	if len(namedWorksStack) != 0 {
		panic("len(namedWorksStack) != 0")
	}
	muNamedWorksStack.Unlock()

	interrupt()
	atomic.StoreInt32(&atomicConnected, 0)
	comports.CloseAllComports()
	wg.Done()
	go gui.NotifyStopWork()
}

var (
	atomicConnected   int32
	interrupt         = func() {}
	wg                sync.WaitGroup
	namedWorksStack   []string
	muNamedWorksStack sync.Mutex

	interruptDelay   = func() {}
	muInterruptDelay sync.Mutex
)
