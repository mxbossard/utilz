package zlog

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"sync/atomic"

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
)

var (
	defaultLogLevel       *slog.LevelVar
	defaultHandlerOptions *slog.HandlerOptions
	defaultOutput         *inout.WriterRef
	defaultHandler        atomic.Pointer[slog.Handler]
)

func init() {
	defaultLogLevel = &slog.LevelVar{}
	defaultHandlerOptions = &slog.HandlerOptions{}
	defaultOutput = &inout.WriterRef{}

	SetDefaultLogLevel(LevelError)

	SetDefaultHandlerOptions(&slog.HandlerOptions{
		Level: defaultLogLevel,
	})

	SetDefaultOutput(os.Stderr)

	//handler := slog.NewTextHandler(defaultOutput, defaultHandlerOptions)
	handler := NewDefaultHandler()
	SetDefaultHandler(handler)
}

func levelString(l slog.Level) string {
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
		return str("PERF", l-LevelPerf)
	case l < LevelInfo:
		return str("DEBUG", l-LevelDebug)
	case l < LevelWarn:
		return str("INFO", l-LevelInfo)
	case l < LevelError:
		return str("WARN", l-LevelWarn)
	case l < LevelFatal:
		return str("ERROR", l-LevelError)
	default:
		return str("FATAL", l-LevelFatal)
	}
}

type DefaultHandler struct {
	slog.Handler

	qualifier string
}

func NewDefaultHandler() DefaultHandler {
	handler := DefaultHandler{
		Handler: slog.Default().Handler(),
	}
	handler.Enabled(context.Background(), defaultLogLevel.Level())
	return handler
}

func (h DefaultHandler) Handle(ctx context.Context, record slog.Record) error {
	rawMsg := record.Message
	record.Message = levelString(record.Level) + " "
	if h.qualifier != "" {
		//record.Message = fmt.Sprintf("%s [%s] %s", levelString(record.Level), h.qualifier, record.Message)
		record.Message += fmt.Sprintf("[%s] ", h.qualifier)
	}
	record.Message += rawMsg
	return h.Handler.Handle(ctx, record)
}

func (h DefaultHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var qualifier string
	var filtered []slog.Attr
	for _, attr := range attrs {
		if attr.Key == QualifierKey {
			qualifier = attr.Value.String()
		} else {
			filtered = append(filtered, attr)
		}
	}
	newHandler := h.Handler.WithAttrs(filtered)
	return DefaultHandler{
		Handler:   newHandler,
		qualifier: qualifier,
	}
}

func SetDefaultLogLevel(lvl slog.Level) {
	defaultLogLevel.Set(lvl)
	slog.SetLogLoggerLevel(lvl)
}

func SetDefaultHandlerOptions(opts *slog.HandlerOptions) {
	*defaultHandlerOptions = *opts
}

func SetDefaultOutput(out io.Writer) {
	defaultOutput.Set(out)
	log.SetOutput(out)
}

func SetDefaultHandler(handler slog.Handler) {
	defaultHandler.Store(&handler)
}
