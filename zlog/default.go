package zlog

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	"mby.fr/utils/inout"
)

const (
	LevelTrace slog.Level = -12
	LevelPerf  slog.Level = -8
	LevelDebug slog.Level = -4
	LevelInfo  slog.Level = 0
	LevelWarn  slog.Level = 4
	LevelError slog.Level = 8
	LevelFatal slog.Level = 12

	QualifierKey = "qualifier"
	PackageKey   = "pkg"
)

var (
	defaultLogLevel       *slog.LevelVar
	defaultHandlerOptions *slog.HandlerOptions
	defaultOutput         *inout.WriterProxy
	defaultHandler        *handlerProxy
	IgnorePC              = false
)

type handlerProxy struct {
	slog.Handler
}

func (h *handlerProxy) Set(new slog.Handler) {
	h.Handler = new
}

func (h handlerProxy) Enabled(c context.Context, l slog.Level) bool {
	return h.Handler.Enabled(c, l)
}

func (h handlerProxy) Handle(ctx context.Context, r slog.Record) error {
	return h.Handler.Handle(ctx, r)
}

func (h handlerProxy) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &handlerProxy{
		Handler: h.Handler.WithAttrs(attrs),
	}
}

func (h handlerProxy) WithGroup(name string) slog.Handler {
	return &handlerProxy{
		Handler: h.Handler.WithGroup(name),
	}
}

func init() {
	defaultLogLevel = &slog.LevelVar{}
	defaultHandlerOptions = &slog.HandlerOptions{}
	defaultOutput = &inout.WriterProxy{}
	defaultHandler = &handlerProxy{}

	SetLogLevelThreshold(LevelError)

	SetDefaultHandlerOptions(&slog.HandlerOptions{
		Level:     defaultLogLevel,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				levelLabel := levelLabel(level)
				a.Value = slog.StringValue(levelLabel)
			}
			return a
		},
	})

	SetDefaultOutput(os.Stderr)

	//handler := slog.NewTextHandler(defaultOutput, defaultHandlerOptions)
	handler := NewUnstructuredHandler(defaultOutput, defaultHandlerOptions)
	//handler := NewColoredHandler(defaultOutput, defaultHandlerOptions)
	SetDefaultHandler(handler)
}

func levelLabel(l slog.Level) string {
	str := func(base string, val slog.Level) string {
		if val == 0 {
			return base
		}
		return fmt.Sprintf("%s%+d", base, val)
	}

	switch {
	case l < LevelPerf:
		return str("TRACE", l-LevelTrace)
	case l < LevelDebug:
		return str(" PERF", l-LevelPerf)
	case l < LevelInfo:
		return str("DEBUG", l-LevelDebug)
	case l < LevelWarn:
		return str(" INFO", l-LevelInfo)
	case l < LevelError:
		return str(" WARN", l-LevelWarn)
	case l < LevelFatal:
		return str("ERROR", l-LevelError)
	default:
		return str("FATAL", l-LevelFatal)
	}
}

func levelShortLabel(l slog.Level) string {
	str := func(base string, val slog.Level) string {
		if val == 0 {
			return base
		}
		return fmt.Sprintf("%s%+d", base, val)
	}

	switch {
	case l < LevelPerf:
		return str("TRA", l-LevelTrace)
	case l < LevelDebug:
		return str("PRF", l-LevelPerf)
	case l < LevelInfo:
		return str("DBG", l-LevelDebug)
	case l < LevelWarn:
		return str("INF", l-LevelInfo)
	case l < LevelError:
		return str("WAR", l-LevelWarn)
	case l < LevelFatal:
		return str("ERR", l-LevelError)
	default:
		return str("FAT", l-LevelFatal)
	}
}

func SetLogLevelThreshold(lvl slog.Level) {
	defaultLogLevel.Set(lvl)
}

func validateThresholdLevel(lvl slog.Level) {
	if lvl < LevelTrace {
		msg := fmt.Sprintf("zlog threshold level: %d is < LevelTrace", lvl)
		panic(msg)
	} else if lvl > LevelFatal {
		msg := fmt.Sprintf("zlog threshold level: %d is > LevelFatal", lvl)
		panic(msg)
	}
}

func SetLogLevelThreshold0IsTrace(n int) {
	lvl := slog.Level(int(LevelTrace) + 4*n)
	validateThresholdLevel(lvl)
	defaultLogLevel.Set(lvl)
}

func SetLogLevelThreshold0IsFatal(n int) {
	lvl := slog.Level(int(LevelFatal) - 4*n)
	validateThresholdLevel(lvl)
	defaultLogLevel.Set(lvl)
}

func SetDefaultHandlerOptions(opts *slog.HandlerOptions) {
	*defaultHandlerOptions = *opts
}

func SetDefaultOutput(out io.Writer) {
	defaultOutput.Set(out)
	log.SetOutput(out)
}

func SetDefaultHandler(handler slog.Handler, attrs ...slog.Attr) {
	attrs = append(attrs, slog.String(QualifierKey, "default"))
	handler = handler.WithAttrs(attrs)
	defaultHandler.Set(handler)
	updateDefaultHandlers(handler)
}

func DefaultHandler() slog.Handler {
	return defaultHandler
}
