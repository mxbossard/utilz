package zlog

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"
)

const (
	LevelTrace slog.Level = -12
	LevelPerf  slog.Level = -8
	LevelDebug slog.Level = -4
	LevelInfo  slog.Level = 0
	LevelWarn  slog.Level = 4
	LevelError slog.Level = 8
	LevelFatal slog.Level = 12
)

type perfTimer struct {
	logger    *zLogger
	qualifier string
	start     *time.Time
	ended     bool
}

func (t *perfTimer) End() {
	if t.start == nil {
		// nothing to do
		return
	}
	if t.ended {
		msg := fmt.Sprintf("<%s> timer already ended", t.qualifier)
		panic(msg)
	}
	duration := time.Since(*t.start)
	msg := fmt.Sprintf("<%s> timer stopped", t.qualifier)
	t.logger.Perf(msg, "duration", duration)
	t.ended = true
}

type zLogger struct {
	*slog.Logger

	level *slog.LevelVar
}

func (l *zLogger) Trace(msg string, args ...any) {
	l.Log(context.Background(), LevelTrace, msg, args...)
}

func (l *zLogger) TraceContext(ctx context.Context, msg string, args ...any) {
	l.Log(ctx, LevelTrace, msg, args...)
}

func (l *zLogger) Fatal(msg string, args ...any) {
	l.Log(context.Background(), LevelFatal, msg, args...)
}

func (l *zLogger) FatalContext(ctx context.Context, msg string, args ...any) {
	l.Log(ctx, LevelFatal, msg, args...)
}

func (l *zLogger) Perf(msg string, args ...any) {
	l.Log(context.Background(), LevelPerf, msg, args...)
}

func (l *zLogger) PerfContext(ctx context.Context, msg string, args ...any) {
	l.Log(ctx, LevelPerf, msg, args...)
}

func (l *zLogger) StartPerf(inputs ...any) *perfTimer {
	var t perfTimer
	if l.level.Level() > LevelPerf {
		return &t
	}

	var qualifier string
	if len(inputs) > 1 {
		panic("zlog.StartPerf() should not take more than one string as arg")
	} else if len(inputs) == 1 {
		if q, ok := inputs[0].(string); ok {
			qualifier = q
		} else {
			panic("not supported non string param supplied to zlog.StartPerf()")
		}
	} else {
		// TODO: forge qualifier from caller method
		qualifier = "TODO"
	}

	msg := fmt.Sprintf("<%s> timer started ...", qualifier)
	t.logger.Perf(msg)
	t.logger = l
	t.qualifier = "todo"
	now := time.Now()
	t.start = &now
	return &t
}

func New(inputs ...any) *zLogger {
	if len(inputs) > 2 {
		panic("zlog.New() should not take more than two args")
	}

	var qualifier string
	var handler slog.Handler
	for _, i := range inputs {
		switch v := i.(type) {
		case string:
			qualifier = v
		case slog.Handler:
			handler = v
		}
	}

	var level *slog.LevelVar

	if qualifier == "" {
		// TODO: forge qualifier from caller package
		qualifier = "TODO"
	}
	if handler == nil {
		// By default set log level to Error
		level.Set(LevelError)
		// By default log to Stderr
		opts := &slog.HandlerOptions{
			Level: level,
		}
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	logger := slog.New(handler)
	logger = logger.With("qualifier", qualifier)
	zLogger := zLogger{
		Logger: logger,
		level:  level,
	}
	return &zLogger
}
