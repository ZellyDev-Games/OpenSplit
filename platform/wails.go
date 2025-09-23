package platform

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type WailsRuntime struct {
	ctx context.Context
}

func NewWailsRuntime() *WailsRuntime {
	return &WailsRuntime{}
}

func (w *WailsRuntime) Startup(ctx context.Context) {
	w.ctx = ctx
}

func (w *WailsRuntime) OpenFileDialog(options runtime.OpenDialogOptions) (string, error) {
	return runtime.OpenFileDialog(w.ctx, options)
}

func (w *WailsRuntime) SaveFileDialog(options runtime.SaveDialogOptions) (string, error) {
	return runtime.SaveFileDialog(w.ctx, options)
}

func (w *WailsRuntime) Quit() {
	runtime.Quit(w.ctx)
}

func (w *WailsRuntime) WindowSetAlwaysOnTop(onTop bool) {
	runtime.WindowSetAlwaysOnTop(w.ctx, onTop)
}

func (w *WailsRuntime) MessageDialog(options runtime.MessageDialogOptions) (string, error) {
	return runtime.MessageDialog(w.ctx, options)
}

func (w *WailsRuntime) EventsEmit(message string, payload ...interface{}) {
	runtime.EventsEmit(w.ctx, message, payload...)
}
