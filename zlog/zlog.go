package zlog

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"
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
		msg := fmt.Sprintf("%s() timer already ended", t.qualifier)
		panic(msg)
	}
	duration := time.Since(*t.start)
	msg := fmt.Sprintf("%s() ended in %s", t.qualifier, duration)
	t.logger.Perf(msg)
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
	os.Exit(1)
}

func (l *zLogger) FatalContext(ctx context.Context, msg string, args ...any) {
	l.Log(ctx, LevelFatal, msg, args...)
	os.Exit(1)
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
		_, qualifier = CallerInfos(1)
	}

	msg := fmt.Sprintf("%s() started ...", qualifier)
	l.Perf(msg)
	t.logger = l
	t.qualifier = qualifier
	now := time.Now()
	t.start = &now
	return &t
}

func CallerInfos(skip int) (pkgName, funcName string) {
	pc, _, _, ok := runtime.Caller(skip + 1)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		names := strings.Split(details.Name(), ".")
		c := len(names)
		funcName = names[c-1]
		pkgName = strings.Join(names[:c-1], ".")

	} else {
		panic("cannot find caller infos")
	}
	return
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

	if qualifier == "" {
		qualifier, _ = CallerInfos(1)
	}
	if handler == nil {
		handler = *defaultHandler.Load()
	}

	handler = handler.WithAttrs([]slog.Attr{slog.String(QualifierKey, qualifier)})
	logger := slog.New(handler)
	zLogger := zLogger{
		Logger: logger,
		level:  defaultLogLevel,
	}
	return &zLogger
}
