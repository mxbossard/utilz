package zlog

import (
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

func UseColoredDefaultHandler() {
	handler := NewColoredHandler(defaultOutput, defaultHandlerOptions)
	SetDefaultHandler(handler)
}

/*
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
	record.Message = levelLabel(record.Level) + " "
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
*/
