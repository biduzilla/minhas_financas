package loggerutils

import (
	"context"
	"log/slog"
	"path/filepath"
	"runtime"
)

type ErrorAwareHandler struct {
	slog.Handler
}

func (h *ErrorAwareHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level >= slog.LevelError {
		if pc, file, line, ok := runtime.Caller(3); ok {
			funcName := runtime.FuncForPC(pc).Name()
			r.AddAttrs(
				slog.Group("source",
					slog.String("function", funcName),
					slog.String("file", filepath.Base(file)),
					slog.Int("line", line),
				),
			)
		}
	}
	return h.Handler.Handle(ctx, r)
}

func (h *ErrorAwareHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ErrorAwareHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *ErrorAwareHandler) WithGroup(name string) slog.Handler {
	return &ErrorAwareHandler{Handler: h.Handler.WithGroup(name)}
}
