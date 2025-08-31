package logger

import (
	"context"
	"log/slog"
	"runtime"
	"time"
)

var log = new(openSplitLog)

func AddHandler(h slog.Handler) {
	log.handlers = append(log.handlers, h)
}

func Debug(s string) {
	handle(log, s, slog.LevelDebug)
}

func Info(s string) {
	handle(log, s, slog.LevelInfo)
}

func Warn(s string) {
	handle(log, s, slog.LevelWarn)
}

func Error(s string) {
	handle(log, s, slog.LevelError)
}

type openSplitLog struct {
	context  context.Context
	handlers []slog.Handler
}

func handle(o *openSplitLog, s string, level slog.Level) {
	r := slog.NewRecord(time.Now(), level, s, getPCs())
	for _, h := range o.handlers {
		r.Clone()
		err := h.Handle(o.context, r)
		if err != nil {
			return
		}
	}
}

func getPCs() uintptr {
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])
	return pcs[0]
}
