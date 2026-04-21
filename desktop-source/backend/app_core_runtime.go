package backend

import "github.com/wailsapp/wails/v2/pkg/runtime"

func (a *App) CopyText(text string) error {
	return runtime.ClipboardSetText(a.ctx, text)
}
