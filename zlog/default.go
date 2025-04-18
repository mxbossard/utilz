package zlog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"text/template"

	"github.com/mxbossard/utilz/inoutz"
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
	partKey      = "part"
)

var (
	defaultLogLevel                  *slog.LevelVar
	defaultHandlerOptions            *slog.HandlerOptions
	defaultOutput                    *inoutz.WriterProxy
	defaultHandlerProxy              *handlerProxy
	defaultPart                      string
	displayPerfStartTimerAsTrace     = true
	loggingFilepath                  = ""
	fileOutputLoggingReportedAlready = false
	IgnorePC                         = false
	QualifierPadding                 = 30
	TruncatedArgsLength              = 64
)

type handlerProxy struct {
	slog.Handler
}

func (h *handlerProxy) Set(new slog.Handler) {
	if p, ok := new.(handlerProxy); ok {
		h.Handler = p.Handler
	} else {
		h.Handler = new
	}
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
	defaultOutput = &inoutz.WriterProxy{}
	defaultHandlerProxy = &handlerProxy{}

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

	// handler := NewUnstructuredHandler(defaultOutput, defaultHandlerOptions)
	// //handler := NewColoredHandler(defaultOutput, defaultHandlerOptions)
	// SetDefaultHandler(handler)

	if defaultHandlerProxy.Handler == nil {
		UnstructuredConfig()
	}
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

func GetLogLevelThreshold() slog.Level {
	return defaultLogLevel.Level()
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

func SetLogLevelThreshold0IsTrace6IsFatal(n int) {
	lvl := slog.Level(int(LevelTrace) + 4*n)
	validateThresholdLevel(lvl)
	defaultLogLevel.Set(lvl)
}

func SetLogLevelThreshold0IsFatal6IsTrace(n int) {
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

func expandFilepathTemplate(filepath string) string {
	tmpl, err := template.New("filepath").Parse(filepath)
	if err != nil {
		panic(err)
	}

	var sw bytes.Buffer
	data := struct {
		Pid int
	}{
		Pid: os.Getpid(),
	}

	err = tmpl.Execute(&sw, data)
	if err != nil {
		panic(err)
	}
	return sw.String()
}

func SetDefaultTruncatingFileOutput(filepath string) {
	filepath = expandFilepathTemplate(filepath)
	out, err := os.OpenFile(filepath, os.O_CREATE+os.O_WRONLY+os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	out.WriteString("\n==================== Truncated log file ====================\n")
	SetDefaultOutput(out)
	loggingFilepath = filepath
}

func SetDefaultAppendingFileOutput(filepath string) {
	filepath = expandFilepathTemplate(filepath)
	out, err := os.OpenFile(filepath, os.O_CREATE+os.O_WRONLY+os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	out.WriteString("\n==================== Appending to log file ====================\n")
	SetDefaultOutput(out)
	loggingFilepath = filepath
}

func reportFileOutputLogging() {
	if fileOutputLoggingReportedAlready || defaultLogLevel.Level() < slog.LevelDebug {
		return
	}
	fileOutputLoggingReportedAlready = true
	fmt.Fprintf(os.Stderr, "Appending logs into file: [%s]\n", loggingFilepath)
}

func SetDefaultHandler(handler slog.Handler, attrs ...slog.Attr) {
	attrs = append(attrs, slog.String(QualifierKey, "default"))
	handler = handler.WithAttrs(attrs)
	defaultHandlerProxy.Set(handler)
	updateDefaultHandlers(handler)
}

func DefaultHandler() slog.Handler {
	return defaultHandlerProxy
}

func DefaultConfig(attrs ...slog.Attr) {
	handler := defaultHandlerProxy.Handler
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

func ColoredConfig(attrs ...slog.Attr) {
	handler := NewColoredHandler(defaultOutput, defaultHandlerOptions)
	SetDefaultHandler(handler, attrs...)
	logger := slog.New(handler)
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func UnstructuredConfig(attrs ...slog.Attr) {
	var handler slog.Handler
	handler = NewUnstructuredHandler(defaultOutput, defaultHandlerOptions)
	if len(attrs) > 0 {
		handler = handler.WithAttrs(attrs)
	}
	SetDefaultHandler(handler, attrs...)
	logger := slog.New(handler)
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func UncoloredConfig(attrs ...slog.Attr) {
	UnstructuredConfig(attrs...)
}

func SetPart(part string) {
	defaultPart = part
}

func SetQualifierPadding(n int) {
	QualifierPadding = n
}

func SetTruncatedArgsLength(n int) {
	TruncatedArgsLength = n
}

func PerfTimerStartAsTrace(b bool) {
	displayPerfStartTimerAsTrace = b
}
