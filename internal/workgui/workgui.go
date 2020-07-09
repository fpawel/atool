package workgui

import (
	"context"
	"fmt"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/fpawel/comm"
	"github.com/powerman/structlog"
	"strings"
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
func SetWorkLogRecordCompleted(log comm.Logger, workLogRecordID int64) {
	_, err := data.DB.Exec(`UPDATE work_log SET complete_at = ? WHERE record_id = ?`, time.Now(), workLogRecordID)
	if err != nil {
		NotifyWarn(log, err.Error())
	}
}

func Delay(duration time.Duration, name string, backgroundWork WorkFunc) WorkFunc {
	return func(log comm.Logger, ctx context.Context) error {
		what := fmt.Sprintf("задержка: %s %s", name, duration)
		if duration == 0 {
			return nil
		}
		startTime := time.Now()
		log = pkg.LogPrependSuffixKeys(log, "delay_start", startTime.Format("15:04:05"))

		workLogRecordID, err := AddNewWorkLogRecord(what)
		if err != nil {
			return err
		}

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
			SetWorkLogRecordCompleted(log, workLogRecordID)
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

func AddNewWorkLogRecord(workName string) (int64, error) {
	muNamedWorksStack.Lock()
	works := append([]string{}, namedWorksStack...)
	muNamedWorksStack.Unlock()
	if len(workName) > 0 {
		works = append(works, workName)
	}
	s := strings.Join(works, ":")
	return data.AddNewWorkLogRecord(s)
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
