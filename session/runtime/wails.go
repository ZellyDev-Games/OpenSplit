package runtime

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type WailsRuntime struct{}

func (w *WailsRuntime) OpenFileDialog(ctx context.Context, options runtime.OpenDialogOptions) (string, error) {
	return runtime.OpenFileDialog(ctx, options)
}

func (w *WailsRuntime) SaveFileDialog(ctx context.Context, options runtime.SaveDialogOptions) (string, error) {
	return runtime.SaveFileDialog(ctx, options)
}

func Quit(ctx context.Context) {
	runtime.Quit(ctx)
}

func WindowSetAlwaysOnTop(ctx context.Context, onTop bool) {
	runtime.WindowSetAlwaysOnTop(ctx, onTop)
}
