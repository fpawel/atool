package guiwork

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/atool/internal/pkg/comports"
	"github.com/powerman/structlog"
	"sync"
	"sync/atomic"
	"time"
)

type WorkFunc func(*structlog.Logger, context.Context) (string, error)

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

func RunTask(what string, task func() (string, error)) {
	go func() {
		gui.Popup(false, what+": выполняется")
		str, err := task()
		if err != nil {
			gui.PopupError(false, merry.Append(err, what))
			return
		}
		if len(what) == 0 {
			gui.Popup(false, what+": "+str)
			return
		}
		gui.Popup(false, what+": выполнено")

	}()
}

func PerformNewNamedWork(log *structlog.Logger, ctx context.Context, newWorkName string, work WorkFunc) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	muNamedWorksStack.Lock()
	namedWorksStack = append(namedWorksStack, newWorkName)
	level := len(namedWorksStack)
	muNamedWorksStack.Unlock()

	go gui.NotifyPushWork(newWorkName)

	log = pkg.LogPrependSuffixKeys(log, fmt.Sprintf("work%d", level), newWorkName)

	result, err := work(log, ctx)

	muNamedWorksStack.Lock()
	namedWorksStack = namedWorksStack[:len(namedWorksStack)-1]
	muNamedWorksStack.Unlock()

	go gui.NotifyPopWork()

	return result, err
}

func InterruptDelay(log *structlog.Logger) {
	muInterruptDelay.Lock()
	interruptDelay()
	muInterruptDelay.Unlock()
	log.Info("delay: skipped by user")
}

func Delay(log *structlog.Logger, ctx context.Context, duration time.Duration, name string, backgroundWork DelayBackgroundWorkFunc) error {
	startTime := time.Now()
	log = pkg.LogPrependSuffixKeys(log, "delay_start", startTime.Format("15:04:05"))

	// сохранить ссылку на изначальный контекст
	ctxParent := ctx

	// установить коллбэк прерывания задаржки
	muInterruptDelay.Lock()
	ctx, interruptDelay = context.WithTimeout(ctx, duration)
	muInterruptDelay.Unlock()

	_, err := PerformNewNamedWork(log, ctx, fmt.Sprintf("%s: %s", name, duration), func(log *structlog.Logger, ctx context.Context) (string, error) {
		log.Info("delay: begin")
		go gui.NotifyBeginDelay(duration)
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
				return "", ctxParent.Err()
			}
			if ctx.Err() != nil {
				// задержка истекла или прервана
				return "", nil
			}
			if err != nil {
				go gui.PopupError(false, err)
				return "", nil
			}
		}
	})
	return err
}

func performWork(log *structlog.Logger, ctx context.Context, workName string, work WorkFunc) {
	go gui.NotifyStartWork()
	go gui.Popup(false, workName+": выполняется")

	muNamedWorksStack.Lock()
	namedWorksStack = nil
	muNamedWorksStack.Unlock()

	result, err := PerformNewNamedWork(log, ctx, workName, work)

	muNamedWorksStack.Lock()
	namedWorksStack = nil
	muNamedWorksStack.Unlock()

	if err != nil {
		go gui.PopupError(false, merry.Append(err, workName))
		log.PrintErr(err, "stack", pkg.FormatMerryStacktrace(err), "work", workName)
	} else {
		if len(workName) == 0 {
			gui.Popup(false, workName+": "+result)
			return
		}
		go gui.Popup(false, workName+": выполнено")
	}
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
