package log

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
)

// Level aliases.
const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

var (
	mu   sync.RWMutex
	log  *slog.Logger
	ctx  = context.Background()
)

func init() {
	SetOutput(os.Stderr, LevelInfo)
}

// SetOutput configures the logger with the given writer and minimum level.
func SetOutput(w io.Writer, level slog.Level) {
	mu.Lock()
	defer mu.Unlock()
	log = slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}))
}

// SetLevel changes the minimum log level at runtime.
func SetLevel(level slog.Level) {
	mu.Lock()
	defer mu.Unlock()
	handler := log.Handler()
	log = slog.New(handler.WithAttrs(nil))
	_ = handler // keep for potential re-create
	SetOutput(os.Stderr, level)
}

func Debug(msg string, args ...any) {
	mu.RLock()
	l := log
	mu.RUnlock()
	l.DebugContext(ctx, msg, args...)
}

func Info(msg string, args ...any) {
	mu.RLock()
	l := log
	mu.RUnlock()
	l.InfoContext(ctx, msg, args...)
}

func Warn(msg string, args ...any) {
	mu.RLock()
	l := log
	mu.RUnlock()
	l.WarnContext(ctx, msg, args...)
}

func Error(msg string, args ...any) {
	mu.RLock()
	l := log
	mu.RUnlock()
	l.ErrorContext(ctx, msg, args...)
}
