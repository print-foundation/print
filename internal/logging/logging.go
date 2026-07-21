package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

func (l Level) slog() slog.Level {
	switch strings.ToLower(string(l)) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
}

type Options struct {
	Level     Level
	Writer    io.Writer
	JSON      bool
	AddSource bool
}

type logger struct {
	sl *slog.Logger
}

func New(opts Options) Logger {
	w := opts.Writer
	if w == nil {
		w = os.Stderr
	}
	handlerOpts := &slog.HandlerOptions{
		Level:     opts.Level.slog(),
		AddSource: opts.AddSource,
	}
	var h slog.Handler
	if opts.JSON {
		h = slog.NewJSONHandler(w, handlerOpts)
	} else {
		h = slog.NewTextHandler(w, handlerOpts)
	}
	return &logger{sl: slog.New(h)}
}

func NopLogger() Logger {
	return &logger{sl: slog.New(slog.NewTextHandler(io.Discard, nil))}
}

func (l *logger) Debug(msg string, args ...any) { l.sl.Debug(msg, args...) }
func (l *logger) Info(msg string, args ...any)  { l.sl.Info(msg, args...) }
func (l *logger) Warn(msg string, args ...any)  { l.sl.Warn(msg, args...) }
func (l *logger) Error(msg string, args ...any) { l.sl.Error(msg, args...) }

func (l *logger) With(args ...any) Logger {
	return &logger{sl: l.sl.With(args...)}
}

type Tee struct {
	targets []Logger
}

func NewTee(targets ...Logger) Logger {
	return &Tee{targets: targets}
}

func (t *Tee) Debug(msg string, args ...any) { t.each(func(l Logger) { l.Debug(msg, args...) }) }
func (t *Tee) Info(msg string, args ...any)  { t.each(func(l Logger) { l.Info(msg, args...) }) }
func (t *Tee) Warn(msg string, args ...any)  { t.each(func(l Logger) { l.Warn(msg, args...) }) }
func (t *Tee) Error(msg string, args ...any) { t.each(func(l Logger) { l.Error(msg, args...) }) }

func (t *Tee) With(args ...any) Logger {
	next := make([]Logger, len(t.targets))
	for i, tgt := range t.targets {
		next[i] = tgt.With(args...)
	}
	return &Tee{targets: next}
}

func (t *Tee) each(fn func(Logger)) {
	for _, tgt := range t.targets {
		fn(tgt)
	}
}

type FileSink struct {
	mu sync.Mutex
	f  *os.File
}

func OpenFileSink(path string) (*FileSink, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}
	return &FileSink{f: f}, nil
}

func (s *FileSink) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.f.Write(p)
}

func (s *FileSink) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.f.Close()
}

type contextKey struct{}

func IntoContext(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

func FromContext(ctx context.Context) Logger {
	if l, ok := ctx.Value(contextKey{}).(Logger); ok {
		return l
	}
	return NopLogger()
}
