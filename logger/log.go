package logger

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"
)

var log = new(openSplitLog)

func AddHandler(h slog.Handler) {
	log.handlers = append(log.handlers, h)
}

func Debug(module string, s string) {
	handle(log, module, s, slog.LevelDebug)
}

func Debugf(module string, s string, opts ...any) {
	handle(log, module, fmt.Sprintf(s, opts...), slog.LevelDebug)
}

func Info(module string, s string) {
	handle(log, module, s, slog.LevelInfo)
}

func Infof(module string, s string, opts ...any) {
	handle(log, module, fmt.Sprintf(s, opts...), slog.LevelInfo)
}

func Warn(module string, s string) {
	handle(log, module, s, slog.LevelWarn)
}

func Warnf(module string, s string, opts ...any) {
	handle(log, module, fmt.Sprintf(s, opts...), slog.LevelWarn)
}

func Error(module string, s string) {
	handle(log, module, s, slog.LevelError)
}

func Errorf(module string, s string, opts ...any) {
	handle(log, module, fmt.Sprintf(s, opts...), slog.LevelError)
}

type openSplitLog struct {
	context  context.Context
	handlers []slog.Handler
}

func handle(o *openSplitLog, module string, s string, level slog.Level, opts ...any) {
	r := slog.NewRecord(time.Now(), level, s, getPCs())
	r.AddAttrs(slog.String("module", module))

	for _, h := range o.handlers {
		clone := r.Clone()
		err := h.Handle(o.context, clone)
		if err != nil {
			continue
		}
	}
}

func getPCs() uintptr {
	var pcs [1]uintptr
	runtime.Callers(4, pcs[:])
	return pcs[0]
}
