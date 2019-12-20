package guiwork

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/pkg/comports"
	"sync"
	"sync/atomic"
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

func RunWork(ctx context.Context, what string, work func(context.Context) (string, error)) error {
	if IsConnected() {
		return merry.New("already connected")
	}

	wg.Add(1)
	atomic.StoreInt32(&atomicConnected, 1)

	ctx, interrupt = context.WithCancel(ctx)
	go func() {

		go gui.NotifyStartWork()
		go gui.Popup(false, what+": выполняется")

		result, err := work(ctx)
		if err != nil {
			go gui.PopupError(false, merry.Append(err, what))
		} else {
			if len(what) == 0 {
				gui.Popup(false, what+": "+result)
				return
			}
			go gui.Popup(false, what+": выполнено")
		}
		interrupt()
		atomic.StoreInt32(&atomicConnected, 0)
		comports.CloseAllComports()
		wg.Done()
		go gui.NotifyStopWork()
	}()
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

var (
	atomicConnected int32
	interrupt       = func() {}
	wg              sync.WaitGroup
)
