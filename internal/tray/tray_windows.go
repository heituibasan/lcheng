//go:build windows

package tray

import (
	"context"
	goruntime "runtime"

	"github.com/getlantern/systray"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"vpn_clash/internal/appmeta"
)

type Handler struct {
	ctx context.Context
}

func Start(ctx context.Context) {
	if ctx == nil {
		return
	}
	handler := &Handler{ctx: ctx}
	go func() {
		goruntime.LockOSThread()
		systray.Run(handler.onReady, handler.onExit)
	}()
}

func (h *Handler) onReady() {
	if len(iconData) > 0 {
		systray.SetIcon(iconData)
	}
	systray.SetTitle(appmeta.Name)
	systray.SetTooltip(appmeta.Name)

	showItem := systray.AddMenuItem("显示主窗口", "打开绿橙")
	systray.AddSeparator()
	quitItem := systray.AddMenuItem("退出", "退出绿橙")

	go func() {
		for {
			select {
			case <-showItem.ClickedCh:
				wailsruntime.WindowShow(h.ctx)
			case <-quitItem.ClickedCh:
				systray.Quit()
				wailsruntime.Quit(h.ctx)
			}
		}
	}()
}

func (h *Handler) onExit() {}
