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

func (t *perfTimer) End(args ...any) {
	if t.start == nil {
		// nothing to do
		return
	}
	if t.ended {
		msg := fmt.Sprintf("%s timer already ended", t.qualifier)
		panic(msg)
	}
	duration := time.Since(*t.start)
	msg := fmt.Sprintf("%s ended in %s", t.qualifier, duration)
	//t.logger.Perf(msg)
	t.logger.log(context.Background(), LevelPerf, msg, args...)
	t.ended = true
}

type zLogger struct {
	*slog.Logger

	level *slog.LevelVar
}

func (l *zLogger) Trace(msg string, args ...any) {
	l.log(context.Background(), LevelTrace, msg, args...)
}

func (l *zLogger) TraceContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelTrace, msg, args...)
}

func (l *zLogger) Perf(msg string, args ...any) {
	l.log(context.Background(), LevelPerf, msg, args...)
}

func (l *zLogger) PerfContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelPerf, msg, args...)
}

func (l *zLogger) Fatal(msg string, args ...any) {
	l.log(context.Background(), LevelFatal, msg, args...)
	os.Exit(1)
}

func (l *zLogger) FatalContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelFatal, msg, args...)
	os.Exit(1)
}

func (l *zLogger) Panic(msg string, args ...any) {
	l.log(context.Background(), LevelFatal, msg, args...)
	panic(msg)
}

func (l *zLogger) PanicContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelFatal, msg, args...)
	panic(msg)
}

func (l *zLogger) QualifiedPerfTimer(qualifier string, args ...any) *perfTimer {
	var t perfTimer
	if l.level.Level() > LevelPerf {
		return &t
	}

	msg := fmt.Sprintf("%s timer started ...", qualifier)
	l.log(context.Background(), LevelTrace, msg)
	t.logger = l
	t.qualifier = qualifier
	now := time.Now()
	t.start = &now
	return &t
}

func (l *zLogger) PerfTimer(args ...any) *perfTimer {
	if l.level.Level() > LevelPerf {
		var t perfTimer
		return &t
	}

	_, qualifier := CallerInfos(1)
	qualifier += "()"

	return l.QualifiedPerfTimer(qualifier, args...)
}

// log is the low-level logging method for methods that take ...any.
// It must always be called directly by an exported logging method
// or function, because it uses a fixed call depth to obtain the pc.
func (l *zLogger) log(ctx context.Context, level slog.Level, msg string, args ...any) {
	if !l.Enabled(ctx, level) {
		return
	}
	var pc uintptr
	if !IgnorePC {
		var pcs [1]uintptr
		// skip [runtime.Callers, this function, this function's caller]
		runtime.Callers(3, pcs[:])
		pc = pcs[0]
	}
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)
	if ctx == nil {
		ctx = context.Background()
	}
	_ = l.Handler().Handle(ctx, r)
}

// logAttrs is like [Logger.log], but for methods that take ...Attr.
func (l *zLogger) logAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	if !l.Enabled(ctx, level) {
		return
	}
	var pc uintptr
	if !IgnorePC {
		var pcs [1]uintptr
		// skip [runtime.Callers, this function, this function's caller]
		runtime.Callers(3, pcs[:])
		pc = pcs[0]
	}
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.AddAttrs(attrs...)
	if ctx == nil {
		ctx = context.Background()
	}
	_ = l.Handler().Handle(ctx, r)
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

var defaultHandlers []qualifiedHandlerProxy

type qualifiedHandlerProxy struct {
	*handlerProxy
	qualifier, pkgName string
}

func (h qualifiedHandlerProxy) Set(new slog.Handler) {
	updated := new.WithAttrs([]slog.Attr{
		slog.String(PackageKey, h.pkgName),
		slog.String(QualifierKey, h.qualifier),
	})
	h.handlerProxy.Set(updated)
}

func new(defaultHandler slog.Handler, inputs ...any) *zLogger {
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

	pkgName, _ := CallerInfos(2)
	if handler == nil {
		handler = defaultHandler
	}

	if qualifier == "" {
		qualifier = pkgName
	}

	/*
		handler = handler.WithAttrs([]slog.Attr{
			slog.String(PackageKey, pkgName),
			slog.String(QualifierKey, qualifier),
		})
	*/
	proxyHandler := qualifiedHandlerProxy{
		handlerProxy: &handlerProxy{},
		qualifier:    qualifier,
		pkgName:      pkgName,
	}
	proxyHandler.Set(handler)
	defaultHandlers = append(defaultHandlers, proxyHandler)

	logger := slog.New(proxyHandler)
	zLogger := zLogger{
		Logger: logger,
		level:  defaultLogLevel,
	}
	return &zLogger
}

func New(inputs ...any) *zLogger {
	return new(DefaultHandler(), inputs...)
}

/*
func NewUnstructured(inputs ...any) *zLogger {
	return new(NewUnstructuredHandler(defaultOutput, defaultHandlerOptions), inputs...)
}

func NewColored(inputs ...any) *zLogger {
	return new(NewColoredHandler(defaultOutput, defaultHandlerOptions), inputs...)
}
*/

func DefaultConfig(attrs ...slog.Attr) {
	handler := DefaultHandler()
	SetDefaultHandler(handler, attrs...)
	logger := slog.New(handler)
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func updateDefaultHandlers(newDefault slog.Handler) {
	// Update default handlers
	for _, h := range defaultHandlers {
		h.Set(newDefault)
	}
}

func UnstructuredConfig(attrs ...slog.Attr) bool {
	handler := NewUnstructuredHandler(defaultOutput, defaultHandlerOptions).WithAttrs(attrs)
	updateDefaultHandlers(handler)
	logger := slog.New(handler)
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelDebug)
	return true
}

func ColoredConfig(attrs ...slog.Attr) bool {
	handler := NewColoredHandler(defaultOutput, defaultHandlerOptions).WithAttrs(attrs)
	updateDefaultHandlers(handler)
	logger := slog.New(handler)
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelDebug)
	return true
}
