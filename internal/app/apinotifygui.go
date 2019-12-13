package app

import (
	"context"
	"github.com/fpawel/atool/gui"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/lxn/win"
)

type notifyGuiSvc struct{}

var _ api.NotifyGuiService = new(notifyGuiSvc)

func (h *notifyGuiSvc) Open(_ context.Context, hWnd int64) error {
	gui.SetHWndTargetSendMessage(win.HWND(hWnd))
	return nil
}

func (h *notifyGuiSvc) Close(_ context.Context) error {
	gui.SetHWndTargetSendMessage(win.HWND_TOP)
	return nil
}
