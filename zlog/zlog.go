package zlog

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/mxbossard/utilz/utilz"
)

func init() {
	// if flag.Lookup("test.v") != nil {
	// 	SetLogLevelThreshold(LevelDebug)
	// }
}

type perfTimer struct {
	logger    *zLogger
	level     slog.Level
	qualifier string
	args      []any
	start     *time.Time
	ended     bool
	uid       string
}

func (t *perfTimer) End(args ...any) {
	if t.start == nil {
		// nothing to do
		return
	}
	if t.ended {
		// If already ended simply do nothing : allow end with defer and end prematurely.
		// msg := fmt.Sprintf("%s timer already ended", t.qualifier)
		// panic(msg)
		return
	}
	duration := time.Since(*t.start)
	msg := fmt.Sprintf("%s{%s} ended in %s", t.qualifier, t.uid, duration)
	allArgs := append(t.args, args...)
	t.logger.log(context.Background(), t.level, msg, allArgs...)
	t.ended = true
}

func (t perfTimer) SinceStart() time.Duration {
	if t.start == nil {
		return -1 * time.Millisecond
	}
	return time.Since(*t.start)
}

type zLogger struct {
	*slog.Logger

	level *slog.LevelVar
}

func (l *zLogger) Trace(msg string, args ...any) {
	l.log(context.Background(), LevelTrace, msg, args...)
}

func (l *zLogger) Tracef(format string, a ...any) {
	l.log(context.Background(), LevelTrace, fmt.Sprintf(format, a...))
}

func (l *zLogger) TraceContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelTrace, msg, args...)
}

func (l *zLogger) TraceContextf(ctx context.Context, format string, a ...any) {
	l.log(ctx, LevelTrace, fmt.Sprintf(format, a...))
}

func (l *zLogger) Perf(msg string, args ...any) {
	l.log(context.Background(), LevelPerf, msg, args...)
}

func (l *zLogger) Perff(format string, a ...any) {
	l.log(context.Background(), LevelPerf, fmt.Sprintf(format, a...))
}

func (l *zLogger) PerfContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelPerf, msg, args...)
}

func (l *zLogger) PerfContextf(ctx context.Context, format string, a ...any) {
	l.log(ctx, LevelPerf, fmt.Sprintf(format, a...))
}

func (l *zLogger) Fatal(msg string, args ...any) {
	l.log(context.Background(), LevelFatal, msg, args...)
	os.Exit(1)
}

func (l *zLogger) Fatalf(format string, a ...any) {
	l.log(context.Background(), LevelFatal, fmt.Sprintf(format, a...))
	os.Exit(1)
}

func (l *zLogger) FatalContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelFatal, msg, args...)
	os.Exit(1)
}

func (l *zLogger) FatalContextf(ctx context.Context, format string, a ...any) {
	l.log(ctx, LevelFatal, fmt.Sprintf(format, a...))
	os.Exit(1)
}

func (l *zLogger) Panic(msg string, args ...any) {
	l.log(context.Background(), LevelFatal, msg, args...)
	panic(msg)
}

func (l *zLogger) Panicf(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	l.log(context.Background(), LevelFatal, msg)
	panic(msg)
}

func (l *zLogger) PanicContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelFatal, msg, args...)
	panic(msg)
}

func (l *zLogger) PanicContextf(ctx context.Context, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	l.log(ctx, LevelFatal, msg)
	panic(msg)
}

func (l *zLogger) Debugf(format string, a ...any) {
	l.Debug(fmt.Sprintf(format, a...))
}

func (l *zLogger) DebugContextf(ctx context.Context, format string, a ...any) {
	l.DebugContext(ctx, fmt.Sprintf(format, a...))
}

func (l *zLogger) Errorf(format string, a ...any) {
	l.Error(fmt.Sprintf(format, a...))
}

func (l *zLogger) ErrorContextf(ctx context.Context, format string, a ...any) {
	l.ErrorContext(ctx, fmt.Sprintf(format, a...))
}

func (l *zLogger) Infof(format string, a ...any) {
	l.Info(fmt.Sprintf(format, a...))
}

func (l *zLogger) InfoContextf(ctx context.Context, format string, a ...any) {
	l.InfoContext(ctx, fmt.Sprintf(format, a...))
}

func (l *zLogger) Logf(ctx context.Context, level slog.Level, format string, a ...any) {
	l.Log(ctx, level, fmt.Sprintf(format, a...))
}

func (l *zLogger) Warnf(format string, a ...any) {
	l.Warn(fmt.Sprintf(format, a...))
}

func (l *zLogger) WarnContextf(ctx context.Context, format string, a ...any) {
	l.WarnContext(ctx, fmt.Sprintf(format, a...))
}

func (l *zLogger) startTimer(t *perfTimer, qualifier string) {
	t.logger = l
	t.qualifier = qualifier
	now := time.Now()
	t.start = &now
}

func (l *zLogger) QualifiedTraceTimer(qualifier string, args ...any) *perfTimer {
	uid := utilz.ShortUidOrPanic()
	t := perfTimer{level: LevelTrace, args: args, uid: uid}
	if l.level.Level() > LevelTrace {
		return &t
	}

	l.startTimer(&t, qualifier)
	msg := fmt.Sprintf("%s{%s} timer started ...", qualifier, t.uid)
	l.log(context.Background(), LevelTrace, msg, args...)
	return &t
}

func (l *zLogger) TraceTimer(args ...any) *perfTimer {
	uid := utilz.ShortUidOrPanic()
	t := perfTimer{level: LevelTrace, args: args, uid: uid}
	if l.level.Level() > LevelTrace {
		return &t
	}

	_, qualifier := CallerInfos(1)
	qualifier += "()"

	l.startTimer(&t, qualifier)
	msg := fmt.Sprintf("%s{%s} timer started ...", qualifier, t.uid)
	l.log(context.Background(), LevelTrace, msg, args...)
	return &t
}

func (l *zLogger) QualifiedPerfTimer(qualifier string, args ...any) *perfTimer {
	uid := utilz.ShortUidOrPanic()
	t := perfTimer{level: LevelPerf, args: args, uid: uid}
	if l.level.Level() > LevelPerf {
		return &t
	}

	l.startTimer(&t, qualifier)
	msg := fmt.Sprintf("%s{%s} timer started ...", qualifier, t.uid)
	l.log(context.Background(), LevelTrace, msg, args...)
	return &t
}

func (l *zLogger) PerfTimer(args ...any) *perfTimer {
	uid := utilz.ShortUidOrPanic()
	t := perfTimer{level: LevelPerf, args: args, uid: uid}
	if l.level.Level() > LevelPerf {
		return &t
	}

	_, qualifier := CallerInfos(1)
	qualifier += "()"

	l.startTimer(&t, qualifier)
	msg := fmt.Sprintf("%s{%s} timer started ...", qualifier, t.uid)
	lvl := LevelPerf
	if displayPerfStartTimerAsTrace {
		lvl = LevelTrace
	}
	l.log(context.Background(), lvl, msg, args...)
	return &t
}

func (l *zLogger) QualifiedDebugTimer(qualifier string, args ...any) *perfTimer {
	uid := utilz.ShortUidOrPanic()
	t := perfTimer{level: LevelDebug, args: args, uid: uid}
	if l.level.Level() > LevelDebug {
		return &t
	}

	l.startTimer(&t, qualifier)
	msg := fmt.Sprintf("%s{%s} timer started ...", qualifier, t.uid)
	l.log(context.Background(), LevelTrace, msg, args...)
	return &t
}

func (l *zLogger) DebugTimer(args ...any) *perfTimer {
	uid := utilz.ShortUidOrPanic()
	t := perfTimer{level: LevelDebug, args: args, uid: uid}
	if l.level.Level() > LevelDebug {
		return &t
	}

	_, qualifier := CallerInfos(1)
	qualifier += "()"

	l.startTimer(&t, qualifier)
	msg := fmt.Sprintf("%s{%s} timer started ...", qualifier, t.uid)
	l.log(context.Background(), LevelTrace, msg, args...)
	return &t
}

func (l *zLogger) QualifiedInfoTimer(qualifier string, args ...any) *perfTimer {
	uid := utilz.ShortUidOrPanic()
	t := perfTimer{level: LevelInfo, args: args, uid: uid}
	if l.level.Level() > LevelInfo {
		return &t
	}

	l.startTimer(&t, qualifier)
	msg := fmt.Sprintf("%s{%s} timer started ...", qualifier, t.uid)
	l.log(context.Background(), LevelTrace, msg, args...)
	return &t
}

func (l *zLogger) InfoTimer(args ...any) *perfTimer {
	uid := utilz.ShortUidOrPanic()
	t := perfTimer{level: LevelInfo, args: args, uid: uid}
	if l.level.Level() > LevelInfo {
		return &t
	}

	_, qualifier := CallerInfos(1)
	qualifier += "()"

	l.startTimer(&t, qualifier)
	msg := fmt.Sprintf("%s{%s} timer started ...", qualifier, t.uid)
	l.log(context.Background(), LevelTrace, msg, args...)
	return &t
}

// log is the low-level logging method for methods that take ...any.
// It must always be called directly by an exported logging method
// or function, because it uses a fixed call depth to obtain the pc.
func (l *zLogger) log(ctx context.Context, level slog.Level, msg string, args ...any) {
	if !l.Enabled(ctx, level) {
		return
	}
	reportFileOutputLogging()
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
