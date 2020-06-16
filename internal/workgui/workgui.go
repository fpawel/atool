package workgui

import (
	"context"
	"fmt"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/comm"
	"github.com/powerman/structlog"
	"sync"
	"sync/atomic"
	"time"
)

func IsConnected() bool {
	return atomic.LoadInt32(&atomicConnected) != 0
}

func Interrupt() {
	interrupt()
}

func Wait() {
	wg.Wait()
}

func InterruptDelay(log *structlog.Logger) {
	muInterruptDelay.Lock()
	interruptDelay()
	name := delayName
	muInterruptDelay.Unlock()
	NotifyWarn(log, name+" - задержка прервана")
}

func Delay(duration time.Duration, name string, backgroundWork WorkFunc) WorkFunc {
	return func(log comm.Logger, ctx context.Context) error {
		what := fmt.Sprintf("задержка: %s %s", name, duration)
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
		delayName = name
		muInterruptDelay.Unlock()

		log.Info("delay: begin")
		go gui.NotifyBeginDelay(duration, what)
		defer func() {
			muInterruptDelay.Lock()
			interruptDelay()
			muInterruptDelay.Unlock()

			log.Info("delay: end", "delay_elapsed", time.Since(startTime))
			go gui.NotifyEndDelay()
		}()

		for {
			var err error
			if backgroundWork != nil {
				err = backgroundWork(log, ctx)
			}
			if ctxParent.Err() != nil {
				// выполнение прервано
				return ctxParent.Err()
			}
			if ctx.Err() != nil {
				// задержка истекла или прервана
				return nil
			}
			if err != nil {
				NotifyErr(log, err)
				return nil
			}
		}
	}
}

func IgnoreError() {
	ignoreError()
}

func currentWorkLevel() int {
	muNamedWorksStack.Lock()
	defer muNamedWorksStack.Unlock()
	return len(namedWorksStack)
}

var (
	ignoreError = func() {}

	atomicConnected   int32
	interrupt         = func() {}
	wg                sync.WaitGroup
	namedWorksStack   []string
	muNamedWorksStack sync.Mutex

	interruptDelay   = func() {}
	delayName        string
	muInterruptDelay sync.Mutex
)
